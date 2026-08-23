package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/dashboard"
	"github.com/behaviorengineering/gitboard/internal/triage"
)

//go:embed static/*
var staticFS embed.FS

// Options configures the HTTP server.
type Options struct {
	Addr        string
	Projects    []config.Project
	Local       config.Local
	Dash        *dashboard.Service
	Triage      *triage.Analyzer
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
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		payload := opts.Dash.Collect(ctx, config.File{Projects: opts.Projects, Local: opts.Local})
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

	mux.Handle("/", fileServer)
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
