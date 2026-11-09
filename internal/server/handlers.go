package server

import (
	"context"
	"net/http"
	"strings"

	ollyerrors "github.com/behaviorengineering/olly/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/behaviorengineering/gitboard/internal/ai/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/ai/triage"
	"github.com/behaviorengineering/gitboard/internal/server/timing"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

func (a *api) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodGet {
		return
	}
	if _, err := w.Write([]byte("ok\n")); err != nil {
		a.log.Errorf("write health response: %v", err)
	}
}

func (a *api) meta(w http.ResponseWriter, r *http.Request) {
	doc, poll, ok := a.liveDoc(w)
	if !ok {
		return
	}
	resp := metaResponse{
		PollIntervalSeconds: poll,
		UI:                  boardUIConfig(doc),
		Views:               dashboard.ViewSummaries(doc),
		Token:               a.opts.Token,
	}
	if a.opts.Dash != nil {
		ctx, cancel := context.WithTimeout(r.Context(), metaBudget)
		defer cancel()
		tooling := a.opts.Dash.ToolingStatus(ctx)
		resp.Tooling = &tooling
	}
	writeJSON(w, a.log, resp)
}

func (a *api) dashboard(w http.ResponseWriter, r *http.Request) {
	if !a.requireDash(w) {
		return
	}
	opts := dashboard.CollectOpts{
		Fresh:     queryFlag(r, "fresh"),
		ViewID:    strings.TrimSpace(r.URL.Query().Get("view")),
		RowBudget: projectRowBudget,
	}
	if queryFlag(r, "stream") {
		a.dashboardStream(w, r, opts)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), dashboardBudget)
	defer cancel()
	doc, poll, ok := a.liveDoc(w)
	if !ok {
		return
	}
	rec := timing.New()
	var payload board.Dashboard
	var err error
	rec.Track("collect", func() {
		payload, err = a.opts.Dash.CollectStream(ctx, doc, opts, nil)
	})
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		writeError(w, a.log, "dashboard.collect", err)
		return
	}
	payload.PollIntervalSeconds = poll
	payload.UI = boardUIConfig(doc)
	cacheDesc := "warm"
	if opts.Fresh {
		cacheDesc = "fresh"
	}
	rec.Add("cache", 0, cacheDesc)
	rec.WriteHeader(w)
	writeJSON(w, a.log, payload)
}

func (a *api) failures(w http.ResponseWriter, r *http.Request) {
	if !a.requireDash(w) {
		return
	}
	doc, _, ok := a.liveDoc(w)
	if !ok {
		return
	}
	projectID := strings.TrimSpace(r.URL.Query().Get("project"))
	runID := strings.TrimSpace(r.URL.Query().Get("run_id"))
	ctx, cancel := context.WithTimeout(r.Context(), failuresBudget)
	defer cancel()
	jobs, err := a.opts.Dash.FailedJobs(ctx, doc, projectID, runID)
	if err != nil {
		writeError(w, a.log, "failures.list", err)
		return
	}
	writeJSON(w, a.log, failuresResponse{Jobs: jobs})
}

func (a *api) triage(w http.ResponseWriter, r *http.Request) {
	if a.opts.Dash == nil || a.opts.Triage == nil {
		http.Error(w, "triage unavailable", http.StatusServiceUnavailable)
		return
	}
	var body triage.Request
	if err := decodeJSONBody(w, r, &body); err != nil {
		return
	}
	doc, _, ok := a.liveDoc(w)
	if !ok {
		return
	}
	if _, ok := dashboard.FindProject(doc.Projects, body.ProjectID); !ok {
		writeError(w, a.log, "triage", gberr.New(gberr.CodeInvalid, "server.triage", "unknown project"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), triageBudget)
	defer cancel()
	if strings.TrimSpace(body.Log) == "" {
		logText, err := a.opts.Dash.JobLog(ctx, doc, body.ProjectID, body.RunID, body.JobID)
		if err != nil {
			writeError(w, a.log, "triage.job_log", err)
			return
		}
		body.Log = logText
	}
	resp, err := a.opts.Triage.Analyze(ctx, body)
	if err != nil {
		writeError(w, a.log, "triage.analyze", gberr.Wrap(err, gberr.CodeUnavailable, "server.triage", "triage failed").
			With("project_id", body.ProjectID).With("run_id", body.RunID).With("job_id", body.JobID))
		return
	}
	writeJSON(w, a.log, resp)
}

func (a *api) pruneSafe(w http.ResponseWriter, r *http.Request) {
	if !a.requireCommands(w) {
		return
	}
	var body dashboard.PruneSafeRequest
	if err := decodeJSONBody(w, r, &body); err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), pruneBudget)
	defer cancel()
	doc, _, ok := a.liveDoc(w)
	if !ok {
		return
	}
	if err := a.opts.Commands.PruneSafe(ctx, doc, body); err != nil {
		writeError(w, a.log, "prune.safe", err)
		return
	}
	writeJSON(w, a.log, okResponse{OK: true})
}

func (a *api) pullFF(w http.ResponseWriter, r *http.Request) {
	if !a.requireCommands(w) {
		return
	}
	var body dashboard.PullFFRequest
	if err := decodeJSONBody(w, r, &body); err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), pullBudget)
	defer cancel()
	doc, _, ok := a.liveDoc(w)
	if !ok {
		return
	}
	rec := timing.New()
	var result dashboard.PullFFResult
	var err error
	rec.Track("pull", func() {
		result, err = a.opts.Commands.PullFF(ctx, doc, body)
	})
	if err != nil {
		writeError(w, a.log, "pull.ff", err)
		return
	}
	addPullTimings(rec, result.PullPhases)
	rec.WriteHeader(w)
	writeJSON(w, a.log, result)
}

// addPullTimings adds one Server-Timing entry per pull phase that ran.
func addPullTimings(rec *timing.Recorder, p localgit.PullPhases) {
	for _, t := range p.Timings() {
		rec.Add("pull_"+t.Name, t.Duration, "")
	}
}

func (a *api) investigateSync(w http.ResponseWriter, r *http.Request) {
	if !a.requireCommands(w) {
		return
	}
	var body dashboard.SyncInvestigationRequest
	if err := decodeJSONBody(w, r, &body); err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), investigateBudget)
	defer cancel()
	doc, _, ok := a.liveDoc(w)
	if !ok {
		return
	}
	result, err := a.opts.Commands.InvestigateSync(ctx, doc, body)
	if err != nil {
		writeError(w, a.log, "local.sync.investigate", err)
		return
	}
	writeJSON(w, a.log, result)
}

func (a *api) investigatePrune(w http.ResponseWriter, r *http.Request) {
	const op = "agents.prune.investigate"
	if a.opts.Prune == nil {
		http.Error(w, errPruneUnavailable, http.StatusServiceUnavailable)
		return
	}
	if a.opts.Dash == nil || a.opts.Commands == nil {
		http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
		return
	}
	var body pruneagent.Request
	if err := decodeJSONBody(w, r, &body); err != nil {
		return
	}
	if strings.TrimSpace(body.ProjectID) == "" {
		writeError(w, a.log, op, gberr.New(gberr.CodeInvalid, "server.investigatePrune", "project_id is required"))
		return
	}
	doc, _, ok := a.liveDoc(w)
	if !ok {
		return
	}
	if _, ok := dashboard.FindProject(doc.Projects, body.ProjectID); !ok {
		writeError(w, a.log, op, gberr.New(gberr.CodeInvalid, "server.investigatePrune", "unknown project"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), agentBudget)
	defer cancel()
	abs, err := a.opts.Commands.RequireMappedPath(ctx, doc, body.ProjectID, body.WorktreePath)
	if err != nil {
		writeError(w, a.log, op, err)
		return
	}
	body.WorktreePath = abs
	resp, err := a.runPruneAgent(ctx, body)
	if err != nil {
		writeError(w, a.log, op, err)
		return
	}
	writeJSON(w, a.log, resp)
}

// runPruneAgent runs the prune agent under its own span.
func (a *api) runPruneAgent(ctx context.Context, body pruneagent.Request) (resp *pruneagent.Result, err error) {
	ctx, span := otel.Tracer(defaultServiceName).Start(ctx, "agents.prune.investigate")
	defer func() { ollyerrors.EndWithStatus(span, err) }()
	resp, err = a.opts.Prune.Investigate(ctx, body)
	return resp, err
}

// agentSession serves GET /api/agents/sessions/<id> (the session view) and
// GET /api/agents/sessions/<id>/failure (the stored LLM failure).
func (a *api) agentSession(w http.ResponseWriter, r *http.Request) {
	if a.opts.Prune == nil {
		http.Error(w, errPruneUnavailable, http.StatusServiceUnavailable)
		return
	}
	id, sub := sessionPath(r.URL.Path)
	if id == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
	switch sub {
	case "":
		a.sessionView(w, r, id)
	case "failure":
		a.serveFailureDump(w, r, id)
	default:
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

// sessionPath splits /api/agents/sessions/<id>[/<sub>] into id and sub.
func sessionPath(path string) (id, sub string) {
	rest := strings.Trim(strings.TrimPrefix(path, "/api/agents/sessions/"), "/")
	id, sub, _ = strings.Cut(rest, "/")
	return id, sub
}

func (a *api) sessionView(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), sessionBudget)
	defer cancel()
	view, err := a.opts.Prune.Session(ctx, id)
	if err != nil {
		writeError(w, a.log, "agents.sessions.get", err)
		return
	}
	if view.PartialErr != nil {
		a.log.With("op", "agents.sessions.get", "session_id", id).Errorf("partial session read: %v", view.PartialErr)
	}
	writeJSON(w, a.log, view)
}

func (a *api) serveFailureDump(w http.ResponseWriter, r *http.Request, id string) {
	path, err := a.opts.Prune.FailurePath(id)
	if err != nil {
		writeError(w, a.log, "agents.sessions.failure", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// G703: FailurePath joins the sessions root with an id the session store
	// sanitises (letters, digits, '-' and '_' only), so the URL segment
	// cannot leave the sessions directory.
	http.ServeFile(w, r, path) //nolint:gosec // G703: id sanitised by the session store; see above.
}
