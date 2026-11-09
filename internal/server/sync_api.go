package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
)

const errConfigNotWritable = "config path not set; cannot mutate tracked projects"

// registerSyncRoutes adds the Manage routes: discovery, tracked projects,
// sync sources, local roots and views.
func (a *api) registerSyncRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/sync/candidates", allow(a.syncCandidates, http.MethodGet))
	mux.HandleFunc("/api/sync/projects", allow(configWrite[projectAction]{
		op:     "api.sync.projects",
		apply:  a.applyProjectAction,
		render: renderSyncState,
	}.handler(a), http.MethodPost))
	mux.HandleFunc("/api/sync/selection", allow(configWrite[selectionBody]{
		op:     "api.sync.selection",
		budget: syncBudget,
		apply:  a.applySelection,
		render: renderSyncState,
	}.handler(a), http.MethodPut))
	mux.HandleFunc("/api/sync/sources", getOrPut(
		a.readConfig(func(doc boardconfig.File) any { return syncSourcesOf(doc) }),
		configWrite[dashboard.SyncSourceLists]{
			op: "api.sync.sources",
			apply: func(_ context.Context, doc boardconfig.File, body dashboard.SyncSourceLists) (boardconfig.File, error) {
				return a.opts.Commands.SetSyncSources(doc, body)
			},
			render: func(doc boardconfig.File) any { return syncSourcesOf(doc) },
		}.handler(a),
	))
	mux.HandleFunc("/api/local/roots", getOrPut(
		a.readConfig(renderRoots),
		configWrite[rootsResponse]{
			op: "api.local.roots",
			apply: func(_ context.Context, doc boardconfig.File, body rootsResponse) (boardconfig.File, error) {
				return a.opts.Commands.SetLocalRoots(doc, body.Roots)
			},
			render: renderRoots,
		}.handler(a),
	))
	mux.HandleFunc("/api/views", getOrPut(
		a.readConfig(renderSyncState),
		configWrite[viewsBody]{
			op: "api.views",
			apply: func(_ context.Context, doc boardconfig.File, body viewsBody) (boardconfig.File, error) {
				return a.opts.Commands.SetViews(doc, body.Views)
			},
			render: renderSyncState,
		}.handler(a),
	))
}

func renderSyncState(doc boardconfig.File) any { return syncStateOf(doc) }

func renderRoots(doc boardconfig.File) any { return rootsResponse{Roots: doc.Local.Roots} }

// projectAction is the POST /api/sync/projects body.
type projectAction struct {
	Action    string           `json:"action"`
	Host      boardconfig.Host `json:"host"`
	Path      string           `json:"path"`
	ID        string           `json:"id"`
	LocalPath string           `json:"local_path"`
}

// selectionBody is the PUT /api/sync/selection body.
type selectionBody struct {
	Selected []projectsync.RepoRef `json:"selected"`
	Host     string                `json:"host"`
}

// viewsBody is the PUT /api/views body.
type viewsBody struct {
	Views []boardconfig.View `json:"views"`
}

func (a *api) syncCandidates(w http.ResponseWriter, r *http.Request) {
	if !a.requireCommands(w) {
		return
	}
	doc, _, ok := a.liveDoc(w)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), syncBudget)
	defer cancel()
	res, err := a.opts.Commands.DiscoverCandidates(ctx, doc, strings.TrimSpace(r.URL.Query().Get("host")))
	if err != nil {
		writeError(w, a.log, "api.sync.candidates", err)
		return
	}
	writeJSON(w, a.log, candidatesResponse{Candidates: res.Candidates, Warnings: res.Warnings})
}

func (a *api) applyProjectAction(_ context.Context, doc boardconfig.File, body projectAction) (boardconfig.File, error) {
	switch strings.ToLower(strings.TrimSpace(body.Action)) {
	case "add":
		return a.opts.Commands.AddTrackedProject(doc, body.Host, body.Path)
	case "remove":
		return a.opts.Commands.RemoveTrackedProject(doc, body.ID)
	case "set_local_path":
		return a.opts.Commands.SetProjectLocalPath(doc, body.ID, body.LocalPath)
	}
	return doc, gberr.New(gberr.CodeInvalid, "server.applyProjectAction", `action must be "add", "remove", or "set_local_path"`)
}

func (a *api) applySelection(ctx context.Context, doc boardconfig.File, body selectionBody) (boardconfig.File, error) {
	return a.opts.Commands.ApplySyncSelection(ctx, doc, body.Host, body.Selected)
}

// readConfig serves a read-only view of the live config.
func (a *api) readConfig(render func(boardconfig.File) any) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		doc, _, ok := a.liveDoc(w)
		if !ok {
			return
		}
		writeJSON(w, a.log, render(doc))
	}
}

// configWrite is one config-mutating route. Every such route runs the same
// flow: guards, decode, apply, save, reload, render. The response is built
// from the reloaded file, so it shows what the next process start will load.
type configWrite[B any] struct {
	op string
	// budget bounds apply when it calls forges. Zero leaves the request
	// context as it is.
	budget time.Duration
	apply  func(ctx context.Context, doc boardconfig.File, body B) (boardconfig.File, error)
	render func(boardconfig.File) any
}

func (cw configWrite[B]) handler(a *api) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.requireCommands(w) {
			return
		}
		if !a.live.writable() {
			http.Error(w, errConfigNotWritable, http.StatusServiceUnavailable)
			return
		}
		var body B
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		doc, _, ok := a.liveDoc(w)
		if !ok {
			return
		}
		ctx := r.Context()
		if cw.budget > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, cw.budget)
			defer cancel()
		}
		doc, err := cw.apply(ctx, doc, body)
		if err != nil {
			writeError(w, a.log, cw.op, err)
			return
		}
		if err := a.live.replace(doc); err != nil {
			writeError(w, a.log, cw.op, err)
			return
		}
		if doc, _, ok = a.liveDoc(w); ok {
			writeJSON(w, a.log, cw.render(doc))
		}
	}
}
