package server

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	ollyhttp "github.com/behaviorengineering/olly/pkg/http"

	"github.com/behaviorengineering/gitboard/internal/ai/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/ai/triage"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

//go:embed static/*
var staticFS embed.FS

const maxJSONBody = 4 << 20 // 4 MiB.

// Request budgets, one per route family. Each handler sets its budget once
// from the request context; the services below take the deadline as given.
const (
	metaBudget        = 5 * time.Second
	dashboardBudget   = 60 * time.Second
	projectRowBudget  = 15 * time.Second
	failuresBudget    = 45 * time.Second
	triageBudget      = 120 * time.Second
	pruneBudget       = 60 * time.Second
	pullBudget        = 90 * time.Second
	investigateBudget = 90 * time.Second
	agentBudget       = 120 * time.Second
	sessionBudget     = 15 * time.Second
	syncBudget        = 3 * time.Minute
)

// HTTP error message constants for repeated handler responses.
const (
	errMethodNotAllowed     = "method not allowed"
	errDashboardUnavailable = "dashboard unavailable"
	errPruneUnavailable     = "prune agent unavailable"
)

// defaultServiceName names the server spans when Options.ServiceName is empty.
const defaultServiceName = "gitboard"

// Options configures the HTTP server.
type Options struct {
	// Addr is the listen address. The API accepts only Host headers that
	// reach this address (see hostsFor).
	Addr string
	// Token, when set, must arrive in X-Gitboard-Token on every mutating API
	// request. GET /api/meta serves it to the same-origin UI. serve sets it
	// with --allow-non-localhost.
	Token string
	// RedactErrorDetail swaps error and detail text in responses for a fixed
	// message and logs the original, because that text can hold local paths
	// and CLI stderr. serve sets it with --allow-non-localhost.
	RedactErrorDetail bool
	// ServiceName names the API server spans. Empty uses "gitboard".
	ServiceName string
	// ConfigPath, when set, reloads the full config file when mtime advances
	// (so gitboard sync and ui edits update a running board).
	ConfigPath string
	// Doc is the initial full config snapshot (from config.Load).
	Doc         boardconfig.File
	Dash        *dashboard.Service
	Commands    *dashboard.Commands
	Triage      *triage.Analyzer
	Prune       *pruneagent.Service
	PollSeconds int
	Log         observability.Logger
}

// api holds the handlers' shared state: the options and the live config.
type api struct {
	opts   Options
	live   *configLive
	log    observability.Logger
	redact detailRedactor
}

// newAPI validates opts and builds the live config. Panics if opts.Log is
// nil (required dependency at wire-up).
func newAPI(opts Options) *api {
	if opts.Log == nil {
		panic("server.NewMux: Log is required")
	}
	poll := opts.PollSeconds
	if poll == 0 {
		poll = config.EffectivePollSeconds(opts.Doc)
	}
	clearCaches := func() {
		if opts.Dash != nil {
			opts.Dash.ClearCaches()
		}
	}
	return &api{
		opts:   opts,
		live:   newConfigLive(opts.ConfigPath, opts.Doc, poll, clearCaches, opts.Log),
		log:    opts.Log,
		redact: detailRedactor{on: opts.RedactErrorDetail, log: opts.Log},
	}
}

// NewMux returns the gitboard HTTP handler: the /api/ routes behind the
// same-origin guard and a server span, and the embedded UI. Panics if
// opts.Log is nil (required dependency at wire-up).
func NewMux(opts Options) http.Handler {
	a := newAPI(opts)
	routes := http.NewServeMux()
	routes.HandleFunc("/api/health", allow(a.health, http.MethodGet, http.MethodHead))
	routes.HandleFunc("/api/meta", allow(a.meta, http.MethodGet))
	routes.HandleFunc("/api/dashboard", allow(a.dashboard, http.MethodGet))
	a.registerSyncRoutes(routes)
	routes.HandleFunc("/api/failures", allow(a.failures, http.MethodGet))
	routes.HandleFunc("/api/triage", allow(a.triage, http.MethodPost))
	routes.HandleFunc("/api/prune/safe", allow(a.pruneSafe, http.MethodPost))
	routes.HandleFunc("/api/pull/ff", allow(a.pullFF, http.MethodPost))
	routes.HandleFunc("/api/local/sync/investigate", allow(a.investigateSync, http.MethodPost))
	routes.HandleFunc("/api/agents/prune/investigate", allow(a.investigatePrune, http.MethodPost))
	routes.HandleFunc("/api/agents/sessions/", a.agentSession)
	routes.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	serviceName := opts.ServiceName
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	guarded := apiGuard{hosts: hostsFor(opts.Addr), token: opts.Token, next: routes, log: a.log}
	mux := http.NewServeMux()
	mux.Handle("/api/", ollyhttp.WrapHandler(guarded, serviceName, ollyhttp.WithSkipPaths("/api/health")))
	mux.Handle("/", staticHandler())
	return mux
}

// staticHandler serves the embedded UI.
func staticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("server: embed static: " + err.Error())
	}
	return http.FileServer(http.FS(sub))
}

// allow answers 405 unless the request method is one of methods.
func allow(h http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, m := range methods {
			if r.Method == m {
				h(w, r)
				return
			}
		}
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

// getOrPut routes GET to get and PUT to put, and answers 405 otherwise.
func getOrPut(get, put http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			get(w, r)
		case http.MethodPut:
			put(w, r)
		default:
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	}
}

// requireDash answers 503 when the dashboard service is not wired.
func (a *api) requireDash(w http.ResponseWriter) bool {
	if a.opts.Dash == nil {
		http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
		return false
	}
	return true
}

// requireCommands answers 503 when the dashboard commands are not wired.
func (a *api) requireCommands(w http.ResponseWriter) bool {
	if a.opts.Commands == nil {
		http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
		return false
	}
	return true
}

// liveDoc returns the current config snapshot and poll interval, or writes
// a 503 and reports false.
func (a *api) liveDoc(w http.ResponseWriter) (boardconfig.File, int, bool) {
	doc, poll, err := a.live.snapshot()
	if err != nil {
		writeError(w, a.log, "config.snapshot", gberr.Wrap(err, gberr.CodeUnavailable, "server.liveDoc", "config unavailable"))
		return boardconfig.File{}, 0, false
	}
	return doc, poll, true
}

// queryFlag reports whether the query parameter name is "1" or "true".
func queryFlag(r *http.Request, name string) bool {
	v := r.URL.Query().Get(name)
	return v == "1" || strings.EqualFold(v, "true")
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	defer func() { _ = r.Body.Close() }()
	limited := io.LimitReader(r.Body, maxJSONBody+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return err
	}
	if len(raw) > maxJSONBody {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return errors.New("body too large")
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return err
	}
	return nil
}

// boardUIConfig copies presentation settings the UI may display (read-only).
func boardUIConfig(doc boardconfig.File) board.UIConfig {
	patterns := doc.UI.HideBranches
	out := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return board.UIConfig{HideBranches: out}
}
