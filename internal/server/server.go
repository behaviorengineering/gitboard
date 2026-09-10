package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/dashboard"
	"github.com/behaviorengineering/gitboard/internal/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/triage"
	"github.com/behaviorengineering/strop/agentsession"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

//go:embed static/*
var staticFS embed.FS

const maxJSONBody = 4 << 20 // 4 MiB

// HTTP error message constants for repeated handler responses.
const (
	errMethodNotAllowed     = "method not allowed"
	errDashboardUnavailable = "dashboard unavailable"
	errForgeClientMissing   = "forge client missing"
	errUnknownProject       = "unknown project"
)

// Options configures the HTTP server.
type Options struct {
	Addr string
	// ConfigPath, when set, reloads the full config file when mtime advances
	// (so gitboard sync and ui edits update a running board).
	ConfigPath string
	// Doc is the initial full config snapshot (from config.Load).
	Doc         config.File
	Dash        *dashboard.Service
	Commands    *dashboard.Commands
	Triage      *triage.Analyzer
	Prune       *pruneagent.Service
	PollSeconds int
}

// NewMux returns the gitboard HTTP handler.
func NewMux(opts Options) http.Handler {
	mux := http.NewServeMux()
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("server: embed static: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	clearCaches := func() {
		if opts.Dash != nil {
			opts.Dash.ClearCaches()
		}
	}
	poll := opts.PollSeconds
	if poll == 0 {
		poll = opts.Doc.EffectivePollSeconds()
	}
	live := newConfigLive(opts.ConfigPath, opts.Doc, poll, clearCaches)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte("ok\n"))
		}
	})

	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		_, poll := live.snapshot()
		writeJSON(w, map[string]any{
			"poll_interval_seconds": poll,
		})
	})

	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		fresh := r.URL.Query().Get("fresh") == "1" || strings.EqualFold(r.URL.Query().Get("fresh"), "true")
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		doc, poll := live.snapshot()
		payload := opts.Dash.Collect(ctx, doc, fresh)
		payload.PollIntervalSeconds = poll
		if err := ctx.Err(); err != nil {
			http.Error(w, "dashboard timed out or canceled", http.StatusGatewayTimeout)
			return
		}
		writeJSON(w, payload)
	})

	mux.HandleFunc("/api/failures", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		doc, _ := live.snapshot()
		projectID := strings.TrimSpace(r.URL.Query().Get("project"))
		runID := strings.TrimSpace(r.URL.Query().Get("run_id"))
		p, ok := dashboard.FindProject(doc.Projects, projectID)
		if !ok {
			http.Error(w, errUnknownProject, http.StatusBadRequest)
			return
		}
		client := dashboard.ClientFor(opts.Dash, p)
		if client == nil {
			http.Error(w, errForgeClientMissing, http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
		defer cancel()
		jobs, err := client.FailedJobs(ctx, p, runID)
		if err != nil {
			http.Error(w, "failed to list jobs", http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]any{"jobs": jobs})
	})

	mux.HandleFunc("/api/triage", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil || opts.Triage == nil {
			http.Error(w, "triage unavailable", http.StatusServiceUnavailable)
			return
		}
		var body triage.Request
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		doc, _ := live.snapshot()
		p, ok := dashboard.FindProject(doc.Projects, body.ProjectID)
		if !ok {
			http.Error(w, errUnknownProject, http.StatusBadRequest)
			return
		}
		client := dashboard.ClientFor(opts.Dash, p)
		if client == nil {
			http.Error(w, errForgeClientMissing, http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()
		if strings.TrimSpace(body.Log) == "" {
			logText, err := client.JobLog(ctx, p, body.RunID, body.JobID)
			if err != nil {
				http.Error(w, "failed to fetch job log", http.StatusBadGateway)
				return
			}
			body.Log = logText
		}
		resp, err := opts.Triage.Analyze(ctx, body)
		if err != nil {
			http.Error(w, "triage failed", http.StatusBadGateway)
			return
		}
		writeJSON(w, resp)
	})

	mux.HandleFunc("/api/prune/safe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Commands == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		var body dashboard.PruneSafeRequest
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		doc, _ := live.snapshot()
		if err := opts.Commands.PruneSafe(ctx, doc, body); err != nil {
			code := http.StatusInternalServerError
			if dashboard.IsBadRequest(err) {
				code = http.StatusBadRequest
			}
			http.Error(w, err.Error(), code)
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/pull/ff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Commands == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		var body dashboard.PullFFRequest
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		doc, _ := live.snapshot()
		result, err := opts.Commands.PullFF(ctx, doc, body)
		if err != nil {
			code := http.StatusInternalServerError
			if dashboard.IsBadRequest(err) {
				code = http.StatusBadRequest
			}
			http.Error(w, err.Error(), code)
			return
		}
		writeJSON(w, result)
	})

	mux.HandleFunc("/api/local/sync/investigate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Commands == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		var body dashboard.SyncInvestigationRequest
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		doc, _ := live.snapshot()
		result, err := opts.Commands.InvestigateSync(ctx, doc, body)
		if err != nil {
			code := http.StatusInternalServerError
			if dashboard.IsBadRequest(err) {
				code = http.StatusBadRequest
			}
			http.Error(w, err.Error(), code)
			return
		}
		writeJSON(w, result)
	})

	mux.HandleFunc("/api/agents/prune/investigate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Prune == nil {
			http.Error(w, "prune agent unavailable", http.StatusServiceUnavailable)
			return
		}
		if opts.Dash == nil || opts.Commands == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		var body pruneagent.Request
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		if strings.TrimSpace(body.ProjectID) == "" {
			http.Error(w, "project_id is required", http.StatusBadRequest)
			return
		}
		doc, _ := live.snapshot()
		if _, ok := dashboard.FindProject(doc.Projects, body.ProjectID); !ok {
			http.Error(w, errUnknownProject, http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()
		abs, err := opts.Commands.RequireMappedPath(ctx, doc, body.ProjectID, body.WorktreePath)
		if err != nil {
			code := http.StatusBadRequest
			if !dashboard.IsBadRequest(err) {
				code = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), code)
			return
		}
		body.WorktreePath = abs
		tr := otel.Tracer("gitboard")
		ctx, span := tr.Start(ctx, "agents.prune.investigate")
		defer span.End()
		resp, err := opts.Prune.Investigate(ctx, body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, "investigate failed", http.StatusBadGateway)
			return
		}
		span.SetStatus(codes.Ok, "")
		writeJSON(w, resp)
	})

	mux.HandleFunc("/api/agents/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if opts.Prune == nil || opts.Prune.Store == nil {
			http.Error(w, "prune agent unavailable", http.StatusServiceUnavailable)
			return
		}
		rest := strings.TrimPrefix(r.URL.Path, "/api/agents/sessions/")
		rest = strings.Trim(rest, "/")
		parts := strings.Split(rest, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "missing session id", http.StatusBadRequest)
			return
		}
		id := parts[0]
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		if len(parts) == 2 && parts[1] == "failure" && r.Method == http.MethodGet {
			dir, err := opts.Prune.Store.Dir(id)
			if err != nil {
				http.Error(w, "invalid session id", http.StatusBadRequest)
				return
			}
			path := filepath.Join(dir, agentsession.FileFailure)
			st, err := os.Stat(path)
			if err != nil || st.IsDir() {
				http.Error(w, "failure dump not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			http.ServeFile(w, r, path)
			return
		}
		if r.Method != http.MethodGet || len(parts) != 1 {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		meta, err := opts.Prune.Store.Load(ctx, id)
		if err != nil {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		var card pruneagent.Card
		cardErr := opts.Prune.Store.LoadJSON(ctx, id, agentsession.FileCard, &card)
		turns, turnsErr := opts.Prune.Store.ReadTurns(ctx, id)
		out := map[string]any{
			"meta":  meta,
			"card":  card,
			"turns": turns,
		}
		if cardErr != nil && !errors.Is(cardErr, os.ErrNotExist) {
			out["card_error"] = "failed to load card"
		}
		if turnsErr != nil && !errors.Is(turnsErr, os.ErrNotExist) {
			out["turns_error"] = "failed to load turns"
		}
		writeJSON(w, out)
	})

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.Handle("/", fileServer)
	return mux
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

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("gitboard: write json: %v", err)
	}
}
