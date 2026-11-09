package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/ai/llm"
	"github.com/behaviorengineering/gitboard/internal/ai/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/ai/triage"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

type fakeExec struct {
	responses map[string][]byte
}

type syncLocalFake struct {
	result localgit.SyncInspection
}

func (f *syncLocalFake) ScanRoots(context.Context, []string) localgit.Discovery {
	return localgit.Discovery{}
}

func (f *syncLocalFake) InspectPath(_ context.Context, path string) localgit.Status {
	return localgit.Status{Mapped: true, Path: path, Branch: "main"}
}

func (f *syncLocalFake) OriginSync(context.Context, string) (localgit.RepoSync, error) {
	return localgit.RepoSync{}, nil
}

func (f *syncLocalFake) CommonGitDir(_ context.Context, path string) (string, error) {
	return path, nil
}

func (f *syncLocalFake) ListLocalHeads(context.Context, string) ([]string, error) {
	return nil, nil
}

func (f *syncLocalFake) FetchOriginSmart(context.Context, string, localgit.FetchPolicy, *localgit.OriginFetchCache) (bool, error) {
	return true, nil
}

func (f *syncLocalFake) OriginRemote(context.Context, string) (string, error) {
	return "https://example.test/acme/app.git", nil
}

func (f *syncLocalFake) PullFFOnly(context.Context, string, string) (localgit.PullPhases, error) {
	return localgit.PullPhases{}, nil
}

func (f *syncLocalFake) RemoveSafeCheckout(context.Context, string, string, string) error {
	return nil
}

func (f *syncLocalFake) EnsureWritableIndex(context.Context, string, bool) error {
	return nil
}

func (f *syncLocalFake) ContentOnDefault(context.Context, string, string, string) (bool, string, error) {
	return false, "", nil
}

func (f *syncLocalFake) CompareSync(context.Context, string, string) (localgit.SyncInspection, error) {
	return f.result, nil
}

func (f *syncLocalFake) CanonicalCheckoutPath(_ context.Context, path string) string {
	return path
}

func (f *syncLocalFake) Mutator(*localgit.MutationLease) (dashboard.LocalMutator, error) {
	return f, nil
}

func (f *fakeExec) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *fakeExec) match(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	var best string
	for substr := range f.responses {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	if best == "" {
		return nil, fmt.Errorf("unexpected argv: %s", key)
	}
	return f.responses[best], nil
}

func (f *fakeExec) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	return f.match(name, args...)
}

func (f *fakeExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := f.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(out)) == "" {
		return []byte("[]"), nil
	}
	return out, nil
}

// newDash wires every forge client on run plus local (nil for forge-only tests).
func newDash(run cliexec.Exec, local dashboard.LocalGit) *dashboard.Service {
	cfg := dashboard.DefaultServiceConfig(run)
	cfg.Local = local
	cfg.Mutations = localgit.NewMutationCoordinator(nil)
	return cfg.CreateService()
}

func testFake() *fakeExec {
	return &fakeExec{
		responses: map[string][]byte{
			"auth status":             []byte(""),
			"repos/acme/app --jq":     []byte(`{"default":"main"}`),
			"repos/acme/app/branches": []byte(`[{"name":"main","commit":{"commit":{"committer":{"date":"2026-01-01T00:00:00Z"}}}}]`),
			"run list":                []byte(`[]`),
			"--state open":            []byte(`[]`),
			"--state merged":          []byte(`[]`),
			"run view 99": []byte(`{"jobs":[
				{"databaseId":1001,"name":"test","conclusion":"failure","url":"https://example/job/1001"}
			]}`),
		},
	}
}

func testMux(t *testing.T, fx *fakeExec, triageA *triage.Analyzer) http.Handler {
	t.Helper()
	if fx == nil {
		fx = testFake()
	}
	projects := []boardconfig.Project{
		{ID: "gh-app", Label: "App", Host: boardconfig.HostGitHub, Path: "acme/app"},
	}
	dash := newDash(fx, nil)
	return server.NewMux(server.Options{
		Addr:        testAddr,
		Log:         observability.NewLogger(),
		Doc:         boardconfig.File{Projects: projects},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		Triage:      triageA,
		PollSeconds: 42,
	})
}

func TestDashboardHidesBranchesFromInitialDoc(t *testing.T) {
	fx := &fakeExec{
		responses: map[string][]byte{
			"auth status":         []byte(""),
			"repos/acme/app --jq": []byte(`{"default":"main"}`),
			"repos/acme/app/branches": []byte(`[
				{"name":"main","commit":{"commit":{"committer":{"date":"2026-01-01T00:00:00Z"}}}},
				{"name":"majordomo-context/gitboard","commit":{"commit":{"committer":{"date":"2026-01-02T00:00:00Z"}}}}
			]`),
			"run list":       []byte(`[]`),
			"--state open":   []byte(`[]`),
			"--state merged": []byte(`[]`),
		},
	}
	projects := []boardconfig.Project{
		{ID: "gh-app", Label: "App", Host: boardconfig.HostGitHub, Path: "acme/app"},
	}
	dash := newDash(fx, nil)
	mux := server.NewMux(server.Options{
		Addr: testAddr,
		Log:  observability.NewLogger(),
		Doc: boardconfig.File{
			Projects: projects,
			UI:       boardconfig.UI{HideBranches: []string{"majordomo-context/*"}},
		},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 30,
	})
	res := apiRequest(http.MethodGet, "/api/dashboard?fresh=1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var payload board.Dashboard
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Projects) != 1 {
		t.Fatalf("projects: %d", len(payload.Projects))
	}
	for _, b := range payload.Projects[0].Branches {
		if strings.HasPrefix(b.Name, "majordomo-context/") {
			t.Fatalf("hidden branch still present: %+v", payload.Projects[0].Branches)
		}
	}
	foundMain := false
	for _, b := range payload.Projects[0].Branches {
		if b.Name == "main" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Fatalf("main missing: %+v", payload.Projects[0].Branches)
	}
	if len(payload.UI.HideBranches) != 1 || payload.UI.HideBranches[0] != "majordomo-context/*" {
		t.Fatalf("ui.hide_branches: %+v", payload.UI.HideBranches)
	}
}

func TestHealthAndMeta(t *testing.T) {
	mux := testMux(t, nil, nil)

	res := apiRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("health: %d %q", rec.Code, rec.Body.String())
	}

	res = apiRequest(http.MethodGet, "/api/meta", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("meta status: %d", rec.Code)
	}
	var meta map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatal(err)
	}
	if meta["poll_interval_seconds"] != float64(42) {
		t.Fatalf("poll: %+v", meta)
	}
	ui, ok := meta["ui"].(map[string]any)
	if !ok {
		t.Fatalf("ui missing: %+v", meta)
	}
	hb, ok := ui["hide_branches"].([]any)
	if !ok {
		t.Fatalf("hide_branches: %+v", ui)
	}
	if len(hb) != 0 {
		t.Fatalf("want empty hide_branches, got %+v", hb)
	}
	views, ok := meta["views"].([]any)
	if !ok || len(views) != 1 {
		t.Fatalf("views: %+v", meta["views"])
	}
}

func TestMethodNotAllowed(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := apiRequest(http.MethodPost, "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestDashboardHappy(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := apiRequest(http.MethodGet, "/api/dashboard?fresh=1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var payload board.Dashboard
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Projects) != 1 || payload.Projects[0].ID != "gh-app" {
		t.Fatalf("projects: %+v", payload.Projects)
	}
	if payload.PollIntervalSeconds != 42 {
		t.Fatalf("poll: %d", payload.PollIntervalSeconds)
	}
}

func TestDashboardStream(t *testing.T) {
	fx := &fakeExec{
		responses: map[string][]byte{
			"auth status":         []byte(""),
			"repos/acme/app --jq": []byte(`{"default":"main"}`),
			"repos/acme/app/branches": []byte(`[
				{"name":"main","commit":{"commit":{"committer":{"date":"2026-01-01T00:00:00Z"}}}}
			]`),
			"run list":       []byte(`[]`),
			"--state open":   []byte(`[]`),
			"--state merged": []byte(`[]`),
		},
	}
	dash := newDash(fx, nil)
	mux := server.NewMux(server.Options{
		Addr: testAddr,
		Log:  observability.NewLogger(),
		Doc: boardconfig.File{
			Projects: []boardconfig.Project{
				{ID: "gh-app", Label: "App", Host: boardconfig.HostGitHub, Path: "acme/app"},
			},
		},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 30,
	})
	res := apiRequest(http.MethodGet, "/api/dashboard?stream=1&fresh=1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "ndjson") {
		t.Fatalf("content-type: %q", ct)
	}
	events := decodeStream(t, rec.Body.Bytes())
	var types []string
	for _, ev := range events {
		types = append(types, ev.Type)
		checkStreamEvent(t, ev)
	}
	if len(types) != 3 || types[0] != "shell" || types[1] != "project" || types[2] != "done" {
		t.Fatalf("event types: %v", types)
	}
	if n := len(events[0].Dashboard.Projects); n != 1 {
		t.Fatalf("shell projects: %d", n)
	}
	if id := events[1].Project.ID; id != "gh-app" {
		t.Fatalf("project id: %q", id)
	}
}

// streamEvent is the NDJSON event shape the UI reads.
type streamEvent struct {
	Type      string `json:"type"`
	Dashboard *struct {
		Projects []board.ProjectSummary `json:"projects"`
	} `json:"dashboard"`
	Project     *board.ProjectSummary `json:"project"`
	Error       string                `json:"error"`
	Tooling     *board.Tooling        `json:"tooling"`
	GeneratedAt string                `json:"generated_at"`
}

func decodeStream(t *testing.T, body []byte) []streamEvent {
	t.Helper()
	var out []streamEvent
	dec := json.NewDecoder(bytes.NewReader(body))
	for {
		var ev streamEvent
		err := dec.Decode(&ev)
		if errors.Is(err, io.EOF) {
			return out
		}
		if err != nil {
			t.Fatalf("decode: %v body=%s", err, body)
		}
		out = append(out, ev)
	}
}

// checkStreamEvent checks one event against what its type promises.
func checkStreamEvent(t *testing.T, ev streamEvent) {
	t.Helper()
	switch ev.Type {
	case "shell":
		if ev.Dashboard == nil {
			t.Fatal("shell missing dashboard")
		}
		for _, p := range ev.Dashboard.Projects {
			if len(p.Branches) > 0 {
				t.Fatalf("shell branches not empty: %+v", p.Branches)
			}
		}
	case "project":
		if ev.Project == nil || len(ev.Project.Branches) == 0 {
			t.Fatalf("project event missing branches: %+v", ev.Project)
		}
	case "done":
		if ev.Tooling == nil || ev.GeneratedAt == "" {
			t.Fatalf("done missing tooling or generated_at: %+v", ev)
		}
	case "error":
		t.Fatalf("stream error: %+v", ev)
	}
}

func TestFailuresUnknownProject(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := apiRequest(http.MethodGet, "/api/failures?project=missing&run_id=99", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestFailuresHappy(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := apiRequest(http.MethodGet, "/api/failures?project=gh-app&run_id=99", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Jobs []board.FailedJob `json:"jobs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Jobs) != 1 || body.Jobs[0].ID != "github:1001" {
		t.Fatalf("jobs: %+v", body.Jobs)
	}
}

func TestPruneSafeBadJSON(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := apiRequest(http.MethodPost, "/api/prune/safe", strings.NewReader("{"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestPruneSafeValidation(t *testing.T) {
	fx := testFake()
	projects := []boardconfig.Project{
		{ID: "gh-app", Label: "App", Host: boardconfig.HostGitHub, Path: "acme/app"},
	}
	dash := newDash(fx, dashboard.NewLocalGit(fx))
	mux := server.NewMux(server.Options{
		Addr:        testAddr,
		Log:         observability.NewLogger(),
		Doc:         boardconfig.File{Projects: projects},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 42,
	})
	body := `{"project_id":"gh-app","branch":"feat","worktree_path":"/tmp/x"}`
	res := apiRequest(http.MethodPost, "/api/prune/safe", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 validation, got %d %s", rec.Code, rec.Body.String())
	}
}

type timedPullLocalFake struct {
	syncLocalFake
	phases localgit.PullPhases
}

func (f *timedPullLocalFake) PullFFOnly(context.Context, string, string) (localgit.PullPhases, error) {
	return f.phases, nil
}

func (f *timedPullLocalFake) Mutator(*localgit.MutationLease) (dashboard.LocalMutator, error) {
	return f, nil
}

func TestPullFFEmitsPhaseTimings(t *testing.T) {
	path := t.TempDir()
	local := &timedPullLocalFake{phases: localgit.PullPhases{FetchMs: 11, MergeMs: 5}}
	projects := []boardconfig.Project{{
		ID:        "gh-app",
		Label:     "App",
		Host:      boardconfig.HostGitHub,
		Path:      "acme/app",
		LocalPath: path,
	}}
	fx := testFake()
	dash := newDash(fx, local)
	mux := server.NewMux(server.Options{
		Addr:     testAddr,
		Log:      observability.NewLogger(),
		Doc:      boardconfig.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"main","repo_path":%q}`, path)
	res := apiRequest(http.MethodPost, "/api/pull/ff", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var got dashboard.PullFFResult
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.FetchMs != 11 || got.MergeMs != 5 {
		t.Fatalf("phases: %+v", got)
	}
	st := rec.Header().Get("Server-Timing")
	if !strings.Contains(st, "pull_fetch") || !strings.Contains(st, "pull_merge") {
		t.Fatalf("server-timing: %q", st)
	}
}

func TestLocalSyncInvestigate(t *testing.T) {
	path := t.TempDir()
	local := &syncLocalFake{
		result: localgit.SyncInspection{
			Path:        path,
			Branch:      "main",
			Relation:    "diverged",
			AheadCount:  1,
			BehindCount: 6,
		},
	}
	projects := []boardconfig.Project{{
		ID:        "gh-app",
		Label:     "App",
		Host:      boardconfig.HostGitHub,
		Path:      "acme/app",
		LocalPath: path,
	}}
	fx := testFake()
	dash := newDash(fx, local)
	mux := server.NewMux(server.Options{
		Addr:     testAddr,
		Log:      observability.NewLogger(),
		Doc:      boardconfig.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"main","repo_path":%q}`, path)
	res := apiRequest(http.MethodPost, "/api/local/sync/investigate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var got dashboard.SyncInvestigation
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ProjectID != "gh-app" || got.Result.Relation != "diverged" || got.Result.BehindCount != 6 {
		t.Fatalf("result: %+v", got)
	}
}

func TestTriageWithLogBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "test-model",
			"choices": []map[string]any{
				{"message": map[string]string{
					"content": "SUMMARY: tests failed\nROOT_CAUSE: nil pointer\nFIX_STEPS:\n1. add check\nCONFIDENCE: high\n",
				}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	analyzer := &triage.Analyzer{
		LLM: &llm.Client{
			BaseURL: srv.URL,
			Model:   "test-model",
			HTTP:    srv.Client(),
		},
	}
	mux := testMux(t, nil, analyzer)
	payload := `{"project_id":"gh-app","run_id":"99","job_id":"github:1001","log":"panic: nil"}`
	res := apiRequest(http.MethodPost, "/api/triage", bytes.NewReader([]byte(payload)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var resp triage.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Summary != "tests failed" || resp.Confidence != "high" || len(resp.FixSteps) != 1 {
		t.Fatalf("resp: %+v", resp)
	}
}

func TestPruneInvestigateUnavailable(t *testing.T) {
	mux := testMux(t, nil, nil) // Prune nil.
	res := apiRequest(http.MethodPost, "/api/agents/prune/investigate", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d %s", rec.Code, rec.Body.String())
	}
}

func investigateFakeGit(t *testing.T) (*fakeExec, string) {
	t.Helper()
	dir := t.TempDir()
	fx := &fakeExec{
		responses: map[string][]byte{
			"status --porcelain":            []byte(""),
			"merge-base":                    []byte("abc"),
			"rev-list --left-right --count": []byte("0\t0"),
			"log --oneline":                 []byte(""),
			"diff --name-only":              []byte(""),
		},
	}
	return fx, dir
}

func TestPruneInvestigateLLM(t *testing.T) {
	fx, dir := investigateFakeGit(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "prune-model",
			"choices": []map[string]any{
				{"message": map[string]string{
					"content": "VERDICT: drop\nSUMMARY: LLM says drop it.\nBULLETS:\n- clean\n- empty ahead\nCOMMAND:\n",
				}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	agentsRoot := t.TempDir()
	prune, err := pruneagent.Config{
		AgentsDir: agentsRoot,
		Exec:      fx,
		LLM: &llm.Client{
			BaseURL: srv.URL,
			Model:   "prune-model",
			HTTP:    srv.Client(),
		},
		Mutations: localgit.NewMutationCoordinator(nil),
	}.CreateService()
	if err != nil {
		t.Fatal(err)
	}
	projects := []boardconfig.Project{
		{ID: "gh-app", Label: "App", Host: boardconfig.HostGitHub, Path: "acme/app", LocalPath: dir},
	}
	dash := newDash(fx, dashboard.NewLocalGit(fx))
	mux := server.NewMux(server.Options{
		Addr:     testAddr,
		Log:      observability.NewLogger(),
		Doc:      boardconfig.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
		Prune:    prune,
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"feat/x","worktree_path":%q,"default_branch":"main"}`, dir)
	res := apiRequest(http.MethodPost, "/api/agents/prune/investigate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var got pruneagent.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Source != "llm" || got.Model != "prune-model" {
		t.Fatalf("source/model: %+v", got)
	}
	if got.Card.Verdict != "drop" || !strings.Contains(got.Card.Summary, "LLM says drop") {
		t.Fatalf("card: %+v", got.Card)
	}
	if got.Card.Command == "" {
		t.Fatal("expected filled drop command")
	}
	if got.SessionID == "" {
		t.Fatal("missing session id")
	}
}

func TestPruneInvestigateLLMFallback(t *testing.T) {
	fx, dir := investigateFakeGit(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	agentsRoot := t.TempDir()
	prune, err := pruneagent.Config{
		AgentsDir: agentsRoot,
		Exec:      fx,
		LLM: &llm.Client{
			BaseURL: srv.URL,
			Model:   "prune-model",
			HTTP:    srv.Client(),
		},
		Mutations: localgit.NewMutationCoordinator(nil),
	}.CreateService()
	if err != nil {
		t.Fatal(err)
	}
	projects := []boardconfig.Project{
		{ID: "gh-app", Label: "App", Host: boardconfig.HostGitHub, Path: "acme/app", LocalPath: dir},
	}
	dash := newDash(fx, dashboard.NewLocalGit(fx))
	mux := server.NewMux(server.Options{
		Addr:     testAddr,
		Log:      observability.NewLogger(),
		Doc:      boardconfig.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
		Prune:    prune,
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"feat/x","worktree_path":%q,"default_branch":"main"}`, dir)
	res := apiRequest(http.MethodPost, "/api/agents/prune/investigate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var got pruneagent.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Source != "rules" {
		t.Fatalf("want rules fallback, got %+v", got)
	}
	if got.Card.Verdict != "drop" {
		t.Fatalf("rules drop card: %+v", got.Card)
	}
}

// Forge and git causes carry argv, stderr and local paths. Clients get the
// stable message and code only.
func TestHandlersSendStableMessageAndCode(t *testing.T) {
	mux := testMux(t, nil, nil)

	// run view 404 has no canned response, so the fake's error names the argv.
	res := apiRequest(http.MethodGet, "/api/failures?project=gh-app&run_id=404", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("failures: want 503, got %d %s", rec.Code, rec.Body.String())
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "failed to list jobs" {
		t.Fatalf("failures body = %q", body)
	}
	if got := rec.Header().Get("X-Gitboard-Error-Code"); got != "unavailable" {
		t.Fatalf("failures code header = %q", got)
	}

	res = apiRequest(http.MethodGet, "/api/dashboard?view=nope", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusBadRequest || strings.Contains(rec.Body.String(), "nope") {
		t.Fatalf("unknown view: %d %q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Gitboard-Error-Code"); got != "invalid" {
		t.Fatalf("unknown view code header = %q", got)
	}
}
