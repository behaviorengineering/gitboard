package server

import (
	"crypto/subtle"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/observability"
)

// TokenHeader carries the per-launch API token on mutating requests.
const TokenHeader = "X-Gitboard-Token" //nolint:gosec // G101: a header name, not a credential.

// apiGuard rejects requests a browser could send from another site. Every
// request must name this server in Host, which stops DNS rebinding on reads
// too. A request with any other method than GET or HEAD must also carry a
// same-site Origin (or none), a JSON body, and the launch token when the
// server has one. A text/plain POST is a simple request that browsers send
// cross-site without a preflight, so binding to loopback alone does not stop
// another page from triggering a prune or pull.
type apiGuard struct {
	hosts map[string]bool
	token string
	next  http.Handler
	log   observability.Logger
}

func (g apiGuard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if reason, status := g.refuse(r); status != 0 {
		g.log.With("op", "api.guard", "method", r.Method, "path", r.URL.Path, "host", r.Host).
			Infof("request refused: %s", reason)
		http.Error(w, reason, status)
		return
	}
	g.next.ServeHTTP(w, r)
}

// refuse returns the client message and status for a refused request, or a
// zero status when the request may proceed.
func (g apiGuard) refuse(r *http.Request) (string, int) {
	if !g.hosts[strings.ToLower(r.Host)] {
		return "host not allowed", http.StatusMisdirectedRequest
	}
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return "", 0
	}
	if origin := r.Header.Get("Origin"); origin != "" && !g.hosts[originHost(origin)] {
		return "cross-origin request refused", http.StatusForbidden
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		return "content type must be application/json", http.StatusUnsupportedMediaType
	}
	if g.token != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get(TokenHeader)), []byte(g.token)) != 1 {
		return "token required", http.StatusUnauthorized
	}
	return "", 0
}

// originHost returns the lower-case host:port of an Origin header value, or
// "" when it does not parse (including the opaque "null" origin).
func originHost(origin string) string {
	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

// hostsFor returns the Host header values that reach a server listening on
// addr: localhost, 127.0.0.1 and [::1] on its port, plus the bound host. A
// wildcard bind (empty host, 0.0.0.0 or ::) also accepts the machine's host
// name and every interface address, so --allow-non-localhost stays usable
// from the LAN without opening the API to rebinding through another name.
func hostsFor(addr string) map[string]bool {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host, port = addr, ""
	}
	names := []string{"localhost", "127.0.0.1", "::1"}
	if wildcard(host) {
		names = append(names, machineNames()...)
	} else {
		names = append(names, host)
	}
	out := make(map[string]bool, len(names))
	for _, n := range names {
		n = strings.ToLower(strings.Trim(n, "[]"))
		if n == "" {
			continue
		}
		if port == "" {
			out[hostHeader(n)] = true
			continue
		}
		out[net.JoinHostPort(n, port)] = true
	}
	return out
}

// hostHeader formats a bare host as a Host header value (IPv6 in brackets).
func hostHeader(host string) string {
	if strings.Contains(host, ":") {
		return "[" + host + "]"
	}
	return host
}

func wildcard(host string) bool {
	switch host {
	case "", "0.0.0.0", "::":
		return true
	}
	return false
}

// machineNames lists the host name and interface addresses. Lookup failures
// shrink the list; they never widen it.
func machineNames() []string {
	var out []string
	if name, err := os.Hostname(); err == nil && name != "" {
		out = append(out, name)
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return out
	}
	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok {
			out = append(out, ipNet.IP.String())
		}
	}
	return out
}
