package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
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

// Options configures the HTTP server.
type Options struct {
	Addr        string
	Projects    []config.Project
	Local       config.Local
	Upstream    config.Upstream
	Dash        *dashboard.Service
	Triage      *triage.Analyzer
	Prune       *pruneagent.Service
	PollSeconds int
}

// NewMux returns the gitboard HTTP handler.
func NewMux(opts Options) http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(staticFS, "static")
	fileServer := http.FileServer(http.FS(sub))

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]any{
			"poll_interval_seconds": opts.PollSeconds,
		})
	})

	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		fresh := r.URL.Query().Get("fresh") == "1" || strings.EqualFold(r.URL.Query().Get("fresh"), "true")
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		doc := config.File{
			Projects: opts.Projects,
			Local:    opts.Local,
			Upstream: opts.Upstream,
		}
		payload := opts.Dash.Collect(ctx, doc, fresh)
		payload.PollIntervalSeconds = opts.PollSeconds
		writeJSON(w, payload)
	})

	mux.HandleFunc("/api/failures", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		projectID := strings.TrimSpace(r.URL.Query().Get("project"))
		runID := strings.TrimSpace(r.URL.Query().Get("run_id"))
		p, ok := dashboard.FindProject(opts.Projects, projectID)
		if !ok {
			http.Error(w, "unknown project", http.StatusBadRequest)
			return
		}
		client := dashboard.ClientFor(opts.Dash, p)
		if client == nil {
			http.Error(w, "forge client missing", http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
		defer cancel()
		jobs, err := client.FailedJobs(ctx, p, runID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]any{"jobs": jobs})
	})

	mux.HandleFunc("/api/triage", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body triage.Request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		p, ok := dashboard.FindProject(opts.Projects, body.ProjectID)
		if !ok {
			http.Error(w, "unknown project", http.StatusBadRequest)
			return
		}
		client := dashboard.ClientFor(opts.Dash, p)
		if client == nil {
			http.Error(w, "forge client missing", http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()
		if strings.TrimSpace(body.Log) == "" {
			logText, err := client.JobLog(ctx, p, body.RunID, body.JobID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			body.Log = logText
		}
		resp, err := opts.Triage.Analyze(ctx, body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, resp)
	})

	mux.HandleFunc("/api/prune/safe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil {
			http.Error(w, "dashboard unavailable", http.StatusServiceUnavailable)
			return
		}
		var body dashboard.PruneSafeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		doc := config.File{
			Projects: opts.Projects,
			Local:    opts.Local,
			Upstream: opts.Upstream,
		}
		if err := opts.Dash.PruneSafe(ctx, doc, body); err != nil {
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
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil {
			http.Error(w, "dashboard unavailable", http.StatusServiceUnavailable)
			return
		}
		var body dashboard.PullFFRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		doc := config.File{
			Projects: opts.Projects,
			Local:    opts.Local,
			Upstream: opts.Upstream,
		}
		result, err := opts.Dash.PullFF(ctx, doc, body)
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
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if opts.Prune == nil {
			http.Error(w, "prune agent unavailable", http.StatusServiceUnavailable)
			return
		}
		var body pruneagent.Request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if _, ok := dashboard.FindProject(opts.Projects, body.ProjectID); body.ProjectID != "" && !ok {
			http.Error(w, "unknown project", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		tr := otel.Tracer("gitboard")
		ctx, span := tr.Start(ctx, "agents.prune.investigate")
		defer func() {
			span.End()
		}()
		resp, err := opts.Prune.Investigate(ctx, body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, err.Error(), http.StatusBadGateway)
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
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			path := filepath.Join(dir, agentsession.FileFailure)
			http.ServeFile(w, r, path)
			return
		}
		if r.Method != http.MethodGet || len(parts) != 1 {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		meta, err := opts.Prune.Store.Load(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		var card pruneagent.Card
		_ = opts.Prune.Store.LoadJSON(ctx, id, agentsession.FileCard, &card)
		turns, _ := opts.Prune.Store.ReadTurns(ctx, id)
		writeJSON(w, map[string]any{
			"meta":  meta,
			"card":  card,
			"turns": turns,
		})
	})

	mux.Handle("/", fileServer)
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
