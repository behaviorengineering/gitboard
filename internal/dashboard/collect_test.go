package dashboard

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/remotegit"
)

// fakeExec mirrors forge client_test fake (kept local to avoid test helper packages).
type fakeExec struct {
	responses map[string][]byte
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

func dualHostFake() *fakeExec {
	return &fakeExec{
		responses: map[string][]byte{
			"auth status":         []byte(""),
			"repos/acme/app --jq": []byte(`{"default":"main"}`),
			"repos/acme/app/branches": []byte(`[
				{"name":"main","commit":{"commit":{"committer":{"date":"2026-01-01T00:00:00Z"}}}},
				{"name":"feat/x","commit":{"commit":{"committer":{"date":"2026-01-02T00:00:00Z"}}}}
			]`),
			"run list":     []byte(`[]`),
			"--state open": []byte(`[]`),
			"--state merged": []byte(`[
				{"number":7,"headRefName":"feat/x","url":"https://example/pr/7","mergedAt":"2026-01-01T12:00:00Z"}
			]`),
			"projects/acme%2Flib?simple=true": []byte(`{"default_branch":"main"}`),
			"repository/branches": []byte(`[
				{"name":"main","commit":{"committed_date":"2026-01-01T00:00:00Z"}}
			]`),
			"ci list":                           []byte(`[]`),
			"mr list -R acme/lib --output json": []byte(`[]`),
			"--merged":                          []byte(`[]`),
		},
	}
}

func TestCollectWithFakeForge(t *testing.T) {
	fx := dualHostFake()
	s := New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), nil)
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
			{ID: "gl-lib", Label: "Lib", Host: config.HostGitLab, Path: "acme/lib"},
		},
	}
	out := s.Collect(context.Background(), doc, true)
	if !out.Tooling.GitHub.Installed || !out.Tooling.GitHub.Authed {
		t.Fatalf("github tooling: %+v", out.Tooling.GitHub)
	}
	if !out.Tooling.GitLab.Installed || !out.Tooling.GitLab.Authed {
		t.Fatalf("gitlab tooling: %+v", out.Tooling.GitLab)
	}
	if len(out.Projects) != 2 {
		t.Fatalf("projects: %d", len(out.Projects))
	}
	byID := map[string]board.ProjectSummary{}
	for _, p := range out.Projects {
		byID[p.ID] = p
	}
	gh := byID["gh-app"]
	if gh.Error != "" || len(gh.RemoteNames) == 0 {
		t.Fatalf("gh-app: %+v", gh)
	}
	if !gh.MergedOK || len(gh.Merged) != 1 {
		t.Fatalf("gh-app merged: ok=%v %+v", gh.MergedOK, gh.Merged)
	}
	gl := byID["gl-lib"]
	if gl.Error != "" {
		t.Fatalf("gl-lib: %+v", gl)
	}
}

func TestCollectNilClients(t *testing.T) {
	s := &Service{}
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh", Host: config.HostGitHub, Path: "acme/app"},
			{ID: "gl", Host: config.HostGitLab, Path: "acme/lib"},
		},
	}
	out := s.Collect(context.Background(), doc, true)
	if len(out.Projects) != 2 {
		t.Fatalf("projects: %d", len(out.Projects))
	}
	for _, p := range out.Projects {
		if p.Error == "" {
			t.Fatalf("want client-missing error for %s", p.ID)
		}
	}
}

func TestFindProject(t *testing.T) {
	projects := []config.Project{
		{ID: "a", Path: "acme/a"},
		{ID: "b", Path: "acme/b"},
	}
	p, ok := FindProject(projects, "b")
	if !ok || p.Path != "acme/b" {
		t.Fatalf("find b: ok=%v %+v", ok, p)
	}
	if _, ok := FindProject(projects, "missing"); ok {
		t.Fatal("missing should be false")
	}
}
