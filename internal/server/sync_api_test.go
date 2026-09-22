package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

func TestDashboardViewQuery(t *testing.T) {
	fx := testFake()
	fx.responses["projects/acme%2Flib?simple=true"] = []byte(`{"default_branch":"main"}`)
	fx.responses["repository/branches"] = []byte(`[{"name":"main","commit":{"committed_date":"2026-01-01T00:00:00Z"}}]`)
	fx.responses["ci list"] = []byte(`[]`)
	fx.responses["mr list -R acme/lib --output json"] = []byte(`[]`)
	fx.responses["--merged"] = []byte(`[]`)

	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), nil)
	mux := server.NewMux(server.Options{
		Doc: config.File{
			Projects: []config.Project{
				{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
				{ID: "gl-lib", Label: "Lib", Host: config.HostGitLab, Path: "acme/lib"},
			},
			Views: []config.View{
				{ID: "work", Label: "Work", Projects: []string{"gh-app"}},
			},
		},
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 30,
	})

	res := httptest.NewRequest(http.MethodGet, "/api/dashboard?view=work", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["active_view"] != "work" {
		t.Fatalf("active_view: %+v", payload["active_view"])
	}
	projects, ok := payload["projects"].([]any)
	if !ok {
		t.Fatalf("projects type: %T", payload["projects"])
	}
	if len(projects) != 1 {
		t.Fatalf("projects: %d", len(projects))
	}

	res = httptest.NewRequest(http.MethodGet, "/api/dashboard?view=nope", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestSyncProjectAddRemoveAndViews(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
		},
		Sync: config.SyncSources{
			GitHub: config.GitHubSync{Orgs: []string{"acme"}},
		},
	}
	if err := config.Save(path, doc); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	fx := testFake()
	dash := dashboard.New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), nil)
	mux := server.NewMux(server.Options{
		ConfigPath:  path,
		Doc:         loaded,
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		PollSeconds: 30,
	})

	res := httptest.NewRequest(http.MethodPost, "/api/sync/projects", bytes.NewBufferString(`{"action":"add","host":"github","path":"acme/beta"}`))
	res.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("add: %d %s", rec.Code, rec.Body.String())
	}

	res = httptest.NewRequest(http.MethodPut, "/api/views", bytes.NewBufferString(`{"views":[{"id":"work","label":"Work","projects":["gh-app","beta"]}]}`))
	res.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("views: %d %s", rec.Code, rec.Body.String())
	}

	_ = os.Chtimes(path, time.Now().Add(time.Second), time.Now().Add(time.Second))

	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Projects) != 2 {
		t.Fatalf("projects: %+v", got.Projects)
	}
	if len(got.Views) != 1 || got.Views[0].ID != "work" {
		t.Fatalf("views: %+v", got.Views)
	}

	res = httptest.NewRequest(http.MethodPost, "/api/sync/projects", bytes.NewBufferString(`{"action":"remove","id":"beta"}`))
	res.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("remove: %d %s", rec.Code, rec.Body.String())
	}
	got, err = config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Projects) != 1 || got.Projects[0].ID != "gh-app" {
		t.Fatalf("after remove: %+v", got.Projects)
	}
	if len(got.Views[0].Projects) != 1 || got.Views[0].Projects[0] != "gh-app" {
		t.Fatalf("view pruned: %+v", got.Views)
	}
}

func TestSyncSourcesPut(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	doc := config.File{Projects: []config.Project{}}
	if err := config.Save(path, doc); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	dash := dashboard.New(remotegit.NewGitHub(testFake()), remotegit.NewGitLab(testFake()), remotegit.NewAzureDevOps(testFake()), remotegit.NewBitbucket(testFake()), nil)
	mux := server.NewMux(server.Options{
		ConfigPath: path,
		Doc:        loaded,
		Dash:       dash,
		Commands:   dashboard.NewCommands(dash),
	})
	res := httptest.NewRequest(http.MethodPut, "/api/sync/sources", bytes.NewBufferString(`{"github_orgs":["acme"],"gitlab_groups":[]}`))
	res.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("sources: %d %s", rec.Code, rec.Body.String())
	}
	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sync.GitHub.Orgs) != 1 || got.Sync.GitHub.Orgs[0] != "acme" {
		t.Fatalf("sync: %+v", got.Sync)
	}
}

func TestLocalRootsAndProjectLocalPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
		},
	}
	if err := config.Save(path, doc); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	dash := dashboard.New(remotegit.NewGitHub(testFake()), remotegit.NewGitLab(testFake()), remotegit.NewAzureDevOps(testFake()), remotegit.NewBitbucket(testFake()), nil)
	mux := server.NewMux(server.Options{
		ConfigPath: path,
		Doc:        loaded,
		Dash:       dash,
		Commands:   dashboard.NewCommands(dash),
	})

	res := httptest.NewRequest(http.MethodPut, "/api/local/roots", bytes.NewBufferString(`{"roots":["~/code","~/code","~/work"]}`))
	res.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("roots: %d %s", rec.Code, rec.Body.String())
	}
	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Local.Roots) != 2 {
		t.Fatalf("roots: %+v", got.Local.Roots)
	}

	res = httptest.NewRequest(http.MethodPost, "/api/sync/projects", bytes.NewBufferString(`{"action":"set_local_path","id":"gh-app","local_path":"~/code/app"}`))
	res.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, res)
	if rec.Code != http.StatusOK {
		t.Fatalf("local_path: %d %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	projects, _ := payload["projects"].([]any)
	if len(projects) != 1 {
		t.Fatalf("projects: %+v", projects)
	}
	row, _ := projects[0].(map[string]any)
	if row["local_path"] != "~/code/app" {
		t.Fatalf("payload local_path: %+v", row)
	}
	got, err = config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Projects[0].LocalPath != "~/code/app" {
		t.Fatalf("saved local_path: %+v", got.Projects[0])
	}
}
