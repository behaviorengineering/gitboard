package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
)

// testAddr is the listen address every test mux is built for.
const testAddr = "127.0.0.1:1325"

// apiRequest is httptest.NewRequest addressed to testAddr. Writes carry a
// JSON content type, as the embedded UI sends them.
func apiRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Host = testAddr
	if method != http.MethodGet && method != http.MethodHead {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func guardMux(t *testing.T, token string) http.Handler {
	t.Helper()
	dash := newDash(testFake(), nil)
	return server.NewMux(server.Options{
		Addr:     testAddr,
		Token:    token,
		Log:      observability.NewLogger(),
		Doc:      boardconfig.File{Projects: []boardconfig.Project{{ID: "gh-app", Host: boardconfig.HostGitHub, Path: "acme/app"}}},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
	})
}

// A cross-site page can send a text/plain POST without a preflight, and DNS
// rebinding can point another name at 127.0.0.1. The guard refuses both
// before any handler runs.
func TestAPIGuardRefusesCrossSiteRequests(t *testing.T) {
	const token = "launch-token"
	body := `{"project_id":"gh-app","branch":"feat","worktree_path":"/tmp/x"}`
	cases := []struct {
		name   string
		token  string
		mutate func(*http.Request)
		want   int
	}{
		{"foreign origin", "", func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") }, http.StatusForbidden},
		{"opaque origin", "", func(r *http.Request) { r.Header.Set("Origin", "null") }, http.StatusForbidden},
		{"text/plain body", "", func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, http.StatusUnsupportedMediaType},
		{"missing content type", "", func(r *http.Request) { r.Header.Del("Content-Type") }, http.StatusUnsupportedMediaType},
		{"wrong host", "", func(r *http.Request) { r.Host = "rebind.example:1325" }, http.StatusMisdirectedRequest},
		{"missing token", token, func(*http.Request) {}, http.StatusUnauthorized},
		{"wrong token", token, func(r *http.Request) { r.Header.Set(server.TokenHeader, "nope") }, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := apiRequest(http.MethodPost, "/api/prune/safe", strings.NewReader(body))
			tc.mutate(req)
			rec := httptest.NewRecorder()
			guardMux(t, tc.token).ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

// Same-origin requests pass: the loopback names on the bound port, a
// matching Origin, and the token when the server has one. The prune handler
// then answers with its own 503, because the test mux has no local git.
func TestAPIGuardAllowsSameOrigin(t *testing.T) {
	const token = "launch-token"
	body := `{"project_id":"gh-app","branch":"feat","worktree_path":"/tmp/x"}`
	cases := []struct {
		name   string
		token  string
		mutate func(*http.Request)
	}{
		{"no origin", "", func(*http.Request) {}},
		{"same origin", "", func(r *http.Request) { r.Header.Set("Origin", "http://127.0.0.1:1325") }},
		{"localhost host", "", func(r *http.Request) {
			r.Host = "localhost:1325"
			r.Header.Set("Origin", "http://localhost:1325")
		}},
		{"json with charset", "", func(r *http.Request) { r.Header.Set("Content-Type", "application/json; charset=utf-8") }},
		{"token sent", token, func(r *http.Request) { r.Header.Set(server.TokenHeader, token) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := apiRequest(http.MethodPost, "/api/prune/safe", strings.NewReader(body))
			tc.mutate(req)
			rec := httptest.NewRecorder()
			guardMux(t, tc.token).ServeHTTP(rec, req)
			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want the handler's 503 (%s)", rec.Code, rec.Body.String())
			}
		})
	}
}

// Reads need no token, but the Host check still applies, so another site
// cannot read /api/meta (and the token in it) through DNS rebinding.
func TestAPIGuardServesTokenToSameOriginOnly(t *testing.T) {
	const token = "launch-token"
	mux := guardMux(t, token)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, apiRequest(http.MethodGet, "/api/meta", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("meta: %d %s", rec.Code, rec.Body.String())
	}
	var meta struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.Token != token {
		t.Fatalf("meta token = %q", meta.Token)
	}

	req := apiRequest(http.MethodGet, "/api/meta", nil)
	req.Host = "rebind.example:1325"
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMisdirectedRequest || strings.Contains(rec.Body.String(), token) {
		t.Fatalf("rebound meta: %d %q", rec.Code, rec.Body.String())
	}
}

// Without --allow-non-localhost there is no token and /api/meta omits it.
func TestMetaOmitsTokenOnLoopback(t *testing.T) {
	rec := httptest.NewRecorder()
	guardMux(t, "").ServeHTTP(rec, apiRequest(http.MethodGet, "/api/meta", nil))
	if strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("meta = %s", rec.Body.String())
	}
}
