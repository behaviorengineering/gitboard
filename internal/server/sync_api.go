package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/syncproj"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
)

const errConfigNotWritable = "config path not set; cannot mutate tracked projects"

func registerSyncRoutes(mux *http.ServeMux, live *configLive, opts Options) {
	mux.HandleFunc("/api/sync/candidates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Commands == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		doc, _ := live.snapshot()
		hostFilter := strings.TrimSpace(r.URL.Query().Get("host"))
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
		defer cancel()
		res, err := opts.Commands.DiscoverCandidates(ctx, doc, hostFilter)
		if err != nil {
			if dashboard.IsBadRequest(err) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]any{
			"candidates": res.Candidates,
			"warnings":   res.Warnings,
		})
	})

	mux.HandleFunc("/api/sync/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Commands == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		if !live.writable() {
			http.Error(w, errConfigNotWritable, http.StatusServiceUnavailable)
			return
		}
		var body struct {
			Action    string      `json:"action"`
			Host      config.Host `json:"host"`
			Path      string      `json:"path"`
			ID        string      `json:"id"`
			LocalPath string      `json:"local_path"`
		}
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		doc, _ := live.snapshot()
		action := strings.ToLower(strings.TrimSpace(body.Action))
		var err error
		switch action {
		case "add":
			doc, err = opts.Commands.AddTrackedProject(doc, body.Host, body.Path)
		case "remove":
			doc, err = opts.Commands.RemoveTrackedProject(doc, body.ID)
		case "set_local_path":
			doc, err = opts.Commands.SetProjectLocalPath(doc, body.ID, body.LocalPath)
		default:
			http.Error(w, `action must be "add", "remove", or "set_local_path"`, http.StatusBadRequest)
			return
		}
		if err != nil {
			writeSyncError(w, err)
			return
		}
		if err := live.replace(doc); err != nil {
			writeSyncError(w, err)
			return
		}
		doc, _ = live.snapshot()
		writeJSON(w, syncStatePayload(doc))
	})

	mux.HandleFunc("/api/sync/selection", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Commands == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		if !live.writable() {
			http.Error(w, errConfigNotWritable, http.StatusServiceUnavailable)
			return
		}
		var body struct {
			Selected []struct {
				Host config.Host `json:"host"`
				Path string      `json:"path"`
			} `json:"selected"`
			Host string `json:"host"`
		}
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		doc, _ := live.snapshot()
		refs := make([]syncproj.RepoRef, 0, len(body.Selected))
		for _, s := range body.Selected {
			refs = append(refs, syncproj.RepoRef{Host: s.Host, Path: s.Path})
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
		defer cancel()
		doc, err := opts.Commands.ApplySyncSelection(ctx, doc, body.Host, refs)
		if err != nil {
			if dashboard.IsBadRequest(err) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if err := live.replace(doc); err != nil {
			writeSyncError(w, err)
			return
		}
		doc, _ = live.snapshot()
		writeJSON(w, syncStatePayload(doc))
	})

	mux.HandleFunc("/api/sync/sources", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			doc, _ := live.snapshot()
			writeJSON(w, map[string]any{
				"github_orgs":          doc.Sync.GitHub.Orgs,
				"gitlab_groups":        doc.Sync.GitLab.Groups,
				"azuredevops_orgs":     doc.Sync.AzureDevOps.Orgs,
				"bitbucket_workspaces": doc.Sync.Bitbucket.Workspaces,
			})
		case http.MethodPut:
			if opts.Commands == nil {
				http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
				return
			}
			if !live.writable() {
				http.Error(w, errConfigNotWritable, http.StatusServiceUnavailable)
				return
			}
			var body struct {
				GitHubOrgs          []string `json:"github_orgs"`
				GitLabGroups        []string `json:"gitlab_groups"`
				AzureDevOpsOrgs     []string `json:"azuredevops_orgs"`
				BitbucketWorkspaces []string `json:"bitbucket_workspaces"`
			}
			if err := decodeJSONBody(w, r, &body); err != nil {
				return
			}
			doc, _ := live.snapshot()
			doc, err := opts.Commands.SetSyncSources(doc, body.GitHubOrgs, body.GitLabGroups, body.AzureDevOpsOrgs, body.BitbucketWorkspaces)
			if err != nil {
				writeSyncError(w, err)
				return
			}
			if err := live.replace(doc); err != nil {
				writeSyncError(w, err)
				return
			}
			doc, _ = live.snapshot()
			writeJSON(w, map[string]any{
				"github_orgs":          doc.Sync.GitHub.Orgs,
				"gitlab_groups":        doc.Sync.GitLab.Groups,
				"azuredevops_orgs":     doc.Sync.AzureDevOps.Orgs,
				"bitbucket_workspaces": doc.Sync.Bitbucket.Workspaces,
			})
		default:
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/local/roots", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			doc, _ := live.snapshot()
			writeJSON(w, map[string]any{"roots": doc.Local.Roots})
		case http.MethodPut:
			if opts.Commands == nil {
				http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
				return
			}
			if !live.writable() {
				http.Error(w, errConfigNotWritable, http.StatusServiceUnavailable)
				return
			}
			var body struct {
				Roots []string `json:"roots"`
			}
			if err := decodeJSONBody(w, r, &body); err != nil {
				return
			}
			doc, _ := live.snapshot()
			doc, err := opts.Commands.SetLocalRoots(doc, body.Roots)
			if err != nil {
				writeSyncError(w, err)
				return
			}
			if err := live.replace(doc); err != nil {
				writeSyncError(w, err)
				return
			}
			doc, _ = live.snapshot()
			writeJSON(w, map[string]any{"roots": doc.Local.Roots})
		default:
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/views", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			doc, _ := live.snapshot()
			writeJSON(w, syncStatePayload(doc))
		case http.MethodPut:
			if opts.Commands == nil {
				http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
				return
			}
			if !live.writable() {
				http.Error(w, errConfigNotWritable, http.StatusServiceUnavailable)
				return
			}
			var body struct {
				Views []config.View `json:"views"`
			}
			if err := decodeJSONBody(w, r, &body); err != nil {
				return
			}
			doc, _ := live.snapshot()
			doc, err := opts.Commands.SetViews(doc, body.Views)
			if err != nil {
				writeSyncError(w, err)
				return
			}
			if err := live.replace(doc); err != nil {
				writeSyncError(w, err)
				return
			}
			doc, _ = live.snapshot()
			writeJSON(w, syncStatePayload(doc))
		default:
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})
}

func writeSyncError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if writeCommandError(w, err) {
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func syncStatePayload(doc config.File) map[string]any {
	projects := make([]map[string]any, 0, len(doc.Projects))
	for _, p := range doc.Projects {
		projects = append(projects, map[string]any{
			"id":         p.ID,
			"label":      p.Label,
			"host":       p.Host,
			"path":       p.Path,
			"local_path": p.LocalPath,
		})
	}
	return map[string]any{
		"projects":  projects,
		"views":     dashboard.ViewSummaries(doc),
		"view_defs": effectiveViewDefs(doc),
		"sync": map[string]any{
			"github_orgs":          doc.Sync.GitHub.Orgs,
			"gitlab_groups":        doc.Sync.GitLab.Groups,
			"azuredevops_orgs":     doc.Sync.AzureDevOps.Orgs,
			"bitbucket_workspaces": doc.Sync.Bitbucket.Workspaces,
		},
		"local": map[string]any{
			"roots": doc.Local.Roots,
		},
	}
}

func effectiveViewDefs(doc config.File) []config.View {
	views := doc.EffectiveViews()
	out := make([]config.View, len(views))
	copy(out, views)
	return out
}
