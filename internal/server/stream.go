package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// dashboardStream emits NDJSON: shell stubs, then each project as it
// finishes, then done (or error).
func (a *api) dashboardStream(w http.ResponseWriter, r *http.Request, opts dashboard.CollectOpts) {
	doc, poll, ok := a.liveDoc(w)
	if !ok {
		return
	}
	shell, err := a.opts.Dash.Shell(doc, opts.ViewID)
	if err != nil {
		writeError(w, a.log, "dashboard.stream.shell", err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	out := &eventWriter{enc: json.NewEncoder(w), flush: flusher, a: a}
	if !out.write(streamEvent{Type: eventShell, Dashboard: &streamShell{
		GeneratedAt:         shell.GeneratedAt,
		PollIntervalSeconds: poll,
		UI:                  boardUIConfig(doc),
		Views:               shell.Views,
		ActiveView:          shell.ActiveView,
		Projects:            shell.Projects,
	}}) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), dashboardBudget)
	defer cancel()
	payload, err := a.opts.Dash.CollectStream(ctx, doc, opts, func(row board.ProjectSummary) {
		if !out.write(streamEvent{Type: eventProject, Project: &row}) {
			cancel()
		}
	})
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		// write logs a failed write; the stream ends either way.
		logError(a.log, "dashboard.stream", err)
		out.write(streamEvent{Type: eventError, Error: publicMessage(err), Code: gberr.CodeOf(err)})
		return
	}
	out.write(streamEvent{
		Type:          eventDone,
		Tooling:       &payload.Tooling,
		GeneratedAt:   payload.GeneratedAt,
		LocalWarnings: payload.LocalWarnings,
	})
}

// eventWriter serialises NDJSON events from concurrent project rows.
type eventWriter struct {
	mu    sync.Mutex
	enc   *json.Encoder
	flush http.Flusher
	a     *api
}

// write encodes ev and flushes it. It reports false once the client is gone.
func (e *eventWriter) write(ev streamEvent) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.enc.Encode(ev); err != nil {
		e.a.log.With("op", "dashboard.stream").Infof("stream write stopped: %v", err)
		return false
	}
	e.flush.Flush()
	return true
}
