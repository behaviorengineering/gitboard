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
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/llm"
	"github.com/behaviorengineering/gitboard/internal/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/internal/triage"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
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

func (f *syncLocalFake) EnrichOriginSync(context.Context, *localgit.Status) {}

func (f *syncLocalFake) CommonGitDir(_ context.Context, path string) (string, error) {
	return path, nil
}

func (f *syncLocalFake) ListLocalHeads(context.Context, string) ([]string, error) {
	return nil, nil
}

func (f *syncLocalFake) FetchOriginCached(context.Context, string, time.Duration, bool, *localgit.OriginFetchCache) (bool, error) {
	return false, nil
}

func (f *syncLocalFake) FetchOriginSmart(context.Context, string, time.Duration, bool, *localgit.OriginFetchCache, []string) (bool, error) {
	return true, nil
}

func (f *syncLocalFake) OriginRemote(context.Context, string) (string, error) {
	return "https://example.test/acme/app.git", nil
}

func (f *syncLocalFake) PullFFOnly(context.Context, string, string) error {
	return nil
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

func (f *syncLocalFake) InspectSync(context.Context, string, string) (localgit.SyncInspection, error) {
	return f.result, nil
}

func (f *syncLocalFake) CompareSync(context.Context, string, string) (localgit.SyncInspection, error) {
	return f.result, nil
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
	projects := []config.Project{
		{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
	}
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), nil)
	return server.NewMux(server.Options{
		Doc:         config.File{Projects: projects},
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
	projects := []config.Project{
		{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
	}
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), nil)
	mux := server.NewMux(server.Options{
		Doc: config.File{
			Projects: projects,
			UI:       config.UI{HideBranches: []string{"majordomo-context/*"}},
		},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 30,
	})
	res := httptest.NewRequest(http.MethodGet, "/api/dashboard?fresh=1", nil)
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

	res := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("health: %d %q", rec.Code, rec.Body.String())
	}

	res = httptest.NewRequest(http.MethodGet, "/api/meta", nil)
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
	res := httptest.NewRequest(http.MethodPost, "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestDashboardHappy(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := httptest.NewRequest(http.MethodGet, "/api/dashboard?fresh=1", nil)
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
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), nil)
	mux := server.NewMux(server.Options{
		Doc: config.File{
			Projects: []config.Project{
				{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
			},
		},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 30,
	})
	res := httptest.NewRequest(http.MethodGet, "/api/dashboard?stream=1&fresh=1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "ndjson") {
		t.Fatalf("content-type: %q", ct)
	}
	var types []string
	var shellProjects int
	var sawProjectID string
	dec := json.NewDecoder(bytes.NewReader(rec.Body.Bytes()))
	for {
		var ev map[string]any
		if err := dec.Decode(&ev); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("decode: %v body=%s", err, rec.Body.String())
		}
		typ, _ := ev["type"].(string)
		types = append(types, typ)
		switch typ {
		case "shell":
			dashObj, _ := ev["dashboard"].(map[string]any)
			projects, _ := dashObj["projects"].([]any)
			shellProjects = len(projects)
			for _, p := range projects {
				row, _ := p.(map[string]any)
				if br, ok := row["branches"]; ok && br != nil {
					if list, ok := br.([]any); ok && len(list) > 0 {
						t.Fatalf("shell branches not empty: %+v", br)
					}
				}
			}
		case "project":
			proj, _ := ev["project"].(map[string]any)
			sawProjectID, _ = proj["id"].(string)
			branches, _ := proj["branches"].([]any)
			if len(branches) == 0 {
				t.Fatalf("project event missing branches: %+v", proj)
			}
		case "done":
			if _, ok := ev["tooling"]; !ok {
				t.Fatal("done missing tooling")
			}
			if _, ok := ev["generated_at"]; !ok {
				t.Fatal("done missing generated_at")
			}
		case "error":
			t.Fatalf("stream error: %+v", ev)
		}
	}
	if len(types) < 3 || types[0] != "shell" || types[len(types)-1] != "done" {
		t.Fatalf("event types: %v", types)
	}
	if shellProjects != 1 {
		t.Fatalf("shell projects: %d", shellProjects)
	}
	if sawProjectID != "gh-app" {
		t.Fatalf("project id: %q", sawProjectID)
	}
	projectCount := 0
	for _, typ := range types {
		if typ == "project" {
			projectCount++
		}
	}
	if projectCount != 1 {
		t.Fatalf("project events: %d types=%v", projectCount, types)
	}
}

func TestFailuresUnknownProject(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := httptest.NewRequest(http.MethodGet, "/api/failures?project=missing&run_id=99", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestFailuresHappy(t *testing.T) {
	mux := testMux(t, nil, nil)
	res := httptest.NewRequest(http.MethodGet, "/api/failures?project=gh-app&run_id=99", nil)
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
	res := httptest.NewRequest(http.MethodPost, "/api/prune/safe", strings.NewReader("{"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestPruneSafeValidation(t *testing.T) {
	fx := testFake()
	projects := []config.Project{
		{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
	}
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), localgit.NewInspector(fx))
	mux := server.NewMux(server.Options{
		Doc:         config.File{Projects: projects},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 42,
	})
	body := `{"project_id":"gh-app","branch":"feat","worktree_path":"/tmp/x"}`
	res := httptest.NewRequest(http.MethodPost, "/api/prune/safe", strings.NewReader(body))
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

func (f *timedPullLocalFake) PullFFOnlyWithPhases(context.Context, string, string) (localgit.PullPhases, error) {
	return f.phases, nil
}

func TestPullFFEmitsPhaseTimings(t *testing.T) {
	path := t.TempDir()
	local := &timedPullLocalFake{phases: localgit.PullPhases{FetchMs: 11, MergeMs: 5}}
	projects := []config.Project{{
		ID:        "gh-app",
		Label:     "App",
		Host:      config.HostGitHub,
		Path:      "acme/app",
		LocalPath: path,
	}}
	fx := testFake()
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), local)
	mux := server.NewMux(server.Options{
		Doc:      config.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"main","repo_path":%q}`, path)
	res := httptest.NewRequest(http.MethodPost, "/api/pull/ff", strings.NewReader(body))
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
	projects := []config.Project{{
		ID:        "gh-app",
		Label:     "App",
		Host:      config.HostGitHub,
		Path:      "acme/app",
		LocalPath: path,
	}}
	fx := testFake()
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), local)
	mux := server.NewMux(server.Options{
		Doc:      config.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"main","repo_path":%q}`, path)
	res := httptest.NewRequest(http.MethodPost, "/api/local/sync/investigate", strings.NewReader(body))
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
	res := httptest.NewRequest(http.MethodPost, "/api/triage", bytes.NewReader([]byte(payload)))
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
	mux := testMux(t, nil, nil) // Prune nil
	res := httptest.NewRequest(http.MethodPost, "/api/agents/prune/investigate", strings.NewReader(`{}`))
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
	prune, err := pruneagent.New(agentsRoot, fx, &llm.Client{
		BaseURL: srv.URL,
		Model:   "prune-model",
		HTTP:    srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{
		{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app", LocalPath: dir},
	}
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), localgit.NewInspector(fx))
	mux := server.NewMux(server.Options{
		Doc:      config.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
		Prune:    prune,
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"feat/x","worktree_path":%q,"default_branch":"main"}`, dir)
	res := httptest.NewRequest(http.MethodPost, "/api/agents/prune/investigate", strings.NewReader(body))
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
	prune, err := pruneagent.New(agentsRoot, fx, &llm.Client{
		BaseURL: srv.URL,
		Model:   "prune-model",
		HTTP:    srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{
		{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app", LocalPath: dir},
	}
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), localgit.NewInspector(fx))
	mux := server.NewMux(server.Options{
		Doc:      config.File{Projects: projects},
		Dash:     dash,
		Commands: dashboard.NewCommands(dash),
		Prune:    prune,
	})
	body := fmt.Sprintf(`{"project_id":"gh-app","branch":"feat/x","worktree_path":%q,"default_branch":"main"}`, dir)
	res := httptest.NewRequest(http.MethodPost, "/api/agents/prune/investigate", strings.NewReader(body))
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
