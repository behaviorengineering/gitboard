package remotegit

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
)

// fakeExec returns canned CLI output matched by argv substrings (longest match wins).
type fakeExec struct {
	responses map[string][]byte
	errors    map[string]error
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
	for substr := range f.errors {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	if best == "" {
		return nil, fmt.Errorf("unexpected argv: %s", key)
	}
	if err, ok := f.errors[best]; ok {
		return nil, err
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

func githubHappyExec() *fakeExec {
	return &fakeExec{
		responses: map[string][]byte{
			"auth status":         []byte(""),
			"repos/acme/app --jq": []byte(`{"default":"main"}`),
			"repos/acme/app/branches": []byte(`[
				{"name":"main","commit":{"commit":{"committer":{"date":"2026-01-01T00:00:00Z"}}}},
				{"name":"feat/x","commit":{"commit":{"committer":{"date":"2026-01-02T00:00:00Z"}}}}
			]`),
			"run list": []byte(`[
				{"databaseId":99,"status":"completed","conclusion":"success","displayTitle":"CI","url":"https://example/run/99","headBranch":"feat/x","updatedAt":"2026-01-02T01:00:00Z","workflowName":"CI"}
			]`),
			"--state open": []byte(`[
				{"number":12,"headRefName":"feat/x","url":"https://example/pr/12","mergeable":"MERGEABLE","mergeStateStatus":"CLEAN","updatedAt":"2026-01-02T02:00:00Z","isDraft":false}
			]`),
			"--state merged": []byte(`[
				{"number":7,"headRefName":"feat/old","url":"https://example/pr/7","mergedAt":"2026-01-01T12:00:00Z"}
			]`),
			"run view 99": []byte(`{"jobs":[
				{"databaseId":1001,"name":"test","conclusion":"failure","url":"https://example/job/1001"},
				{"databaseId":1002,"name":"build","conclusion":"success","url":"https://example/job/1002"}
			]}`),
		},
	}
}

func gitlabHappyExec() *fakeExec {
	return &fakeExec{
		responses: map[string][]byte{
			"auth status":                     []byte(""),
			"projects/acme%2Fapp?simple=true": []byte(`{"default_branch":"main"}`),
			"repository/branches": []byte(`[
				{"name":"main","web_url":"https://gitlab.example/b/main","commit":{"committed_date":"2026-01-01T00:00:00Z"}},
				{"name":"feat/y","web_url":"https://gitlab.example/b/y","commit":{"committed_date":"2026-01-02T00:00:00Z"}}
			]`),
			"ci list": []byte(`[
				{"id":55,"status":"success","ref":"feat/y","sha":"abc","web_url":"https://gitlab.example/p/55","updated_at":"2026-01-02T01:00:00Z"}
			]`),
			"mr list -R acme/app --output json": []byte(`[
				{"iid":3,"source_branch":"feat/y","web_url":"https://gitlab.example/mr/3","updated_at":"2026-01-02T02:00:00Z","has_conflicts":false,"merge_status":"can_be_merged","draft":false,"work_in_progress":false}
			]`),
			"--merged": []byte(`[
				{"iid":2,"source_branch":"feat/done","web_url":"https://gitlab.example/mr/2","merged_at":"2026-01-01T12:00:00Z"}
			]`),
			"ci view 55": []byte(`{"jobs":[
				{"id":901,"name":"test","stage":"test","status":"failed","web_url":"https://gitlab.example/j/901"},
				{"id":902,"name":"build","stage":"build","status":"success","web_url":"https://gitlab.example/j/902"}
			]}`),
		},
	}
}

func TestGitHubProjectSummaryHappy(t *testing.T) {
	gh := NewGitHub(githubHappyExec())
	p := config.Project{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"}
	sum, err := gh.ProjectSummary(context.Background(), p, SummaryOpts{Fresh: true, Cache: NewTTLCache()})
	if err != nil {
		t.Fatal(err)
	}
	if sum.Error != "" {
		t.Fatalf("error: %s", sum.Error)
	}
	if len(sum.RemoteNames) != 2 {
		t.Fatalf("remote names: %v", sum.RemoteNames)
	}
	if sum.CI == nil || sum.CI.RunID != "99" || sum.CI.Conclusion != "success" {
		t.Fatalf("ci: %+v", sum.CI)
	}
	if sum.OpenItems.PullRequests != 1 {
		t.Fatalf("open prs: %d", sum.OpenItems.PullRequests)
	}
	if !sum.MergedOK || len(sum.Merged) != 1 || sum.Merged[0].Branch != "feat/old" {
		t.Fatalf("merged: ok=%v %+v", sum.MergedOK, sum.Merged)
	}
	if !sum.RemoteNamesOK {
		t.Fatal("expected RemoteNamesOK")
	}
	var feat *board.BranchRef
	for i := range sum.Branches {
		if sum.Branches[i].Name == "feat/x" {
			b := sum.Branches[i]
			feat = &b
			break
		}
	}
	if feat == nil || !feat.OpenReview || feat.ReviewID != 12 || feat.CIStatus != "success" {
		t.Fatalf("feat/x branch: %+v", feat)
	}
}

func TestGitHubFailedJobsHappy(t *testing.T) {
	gh := NewGitHub(githubHappyExec())
	p := config.Project{ID: "gh-app", Host: config.HostGitHub, Path: "acme/app"}
	jobs, err := gh.FailedJobs(context.Background(), p, "99")
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].ID != "github:1001" || jobs[0].Name != "test" {
		t.Fatalf("jobs: %+v", jobs)
	}
}

func TestGitLabProjectSummaryHappy(t *testing.T) {
	gl := NewGitLab(gitlabHappyExec())
	p := config.Project{ID: "gl-app", Label: "App", Host: config.HostGitLab, Path: "acme/app"}
	sum, err := gl.ProjectSummary(context.Background(), p, SummaryOpts{Fresh: true, Cache: NewTTLCache()})
	if err != nil {
		t.Fatal(err)
	}
	if sum.Error != "" {
		t.Fatalf("error: %s", sum.Error)
	}
	if !sum.RemoteNamesOK || len(sum.RemoteNames) != 2 {
		t.Fatalf("remote names: ok=%v %v", sum.RemoteNamesOK, sum.RemoteNames)
	}
	if sum.CI == nil || sum.CI.RunID != "55" || sum.CI.Status != "success" {
		t.Fatalf("ci: %+v", sum.CI)
	}
	if sum.OpenItems.MergeRequests != 1 {
		t.Fatalf("open mrs: %d", sum.OpenItems.MergeRequests)
	}
	if !sum.MergedOK || len(sum.Merged) != 1 || sum.Merged[0].Branch != "feat/done" {
		t.Fatalf("merged: ok=%v %+v", sum.MergedOK, sum.Merged)
	}
	for i := range sum.Branches {
		if sum.Branches[i].Name == "feat/y" {
			feat := sum.Branches[i]
			if !feat.OpenReview || feat.ReviewID != 3 || feat.CIStatus != "success" {
				t.Fatalf("feat/y branch: %+v", feat)
			}
			return
		}
	}
	t.Fatal("feat/y branch not found")
}

func TestGitLabFailedJobsHappy(t *testing.T) {
	gl := NewGitLab(gitlabHappyExec())
	p := config.Project{ID: "gl-app", Host: config.HostGitLab, Path: "acme/app"}
	jobs, err := gl.FailedJobs(context.Background(), p, "55")
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].ID != "gitlab:901" || jobs[0].Name != "test" || jobs[0].Stage != "test" {
		t.Fatalf("jobs: %+v", jobs)
	}
}

func TestGitlabHasConflict(t *testing.T) {
	if !gitlabHasConflict(true, "can_be_merged") {
		t.Fatal("has_conflicts true")
	}
	if !gitlabHasConflict(false, "cannot_be_merged") {
		t.Fatal("merge_status cannot_be_merged")
	}
	if gitlabHasConflict(false, "can_be_merged") {
		t.Fatal("clean should not conflict")
	}
}

func TestGithubHasConflict(t *testing.T) {
	if !githubHasConflict("CONFLICTING", "") {
		t.Fatal("mergeable CONFLICTING")
	}
	if !githubHasConflict("MERGEABLE", "DIRTY") {
		t.Fatal("state DIRTY")
	}
	if githubHasConflict("MERGEABLE", "CLEAN") {
		t.Fatal("clean should not conflict")
	}
}
