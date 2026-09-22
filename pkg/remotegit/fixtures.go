package remotegit

import (
	"context"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
)

// FakeExec returns canned CLI output matched by argv substrings (longest match wins).
// Used by package tests and internal/conformity.
type FakeExec struct {
	Responses map[string][]byte
	Errors    map[string]error
}

func (f *FakeExec) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *FakeExec) match(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	var best string
	for substr := range f.Responses {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	for substr := range f.Errors {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	if best == "" {
		return nil, fmt.Errorf("unexpected argv: %s", key)
	}
	if err, ok := f.Errors[best]; ok {
		return nil, err
	}
	return f.Responses[best], nil
}

func (f *FakeExec) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	return f.match(name, args...)
}

func (f *FakeExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := f.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(out)) == "" {
		return []byte("[]"), nil
	}
	return out, nil
}

var _ cliexec.Exec = (*FakeExec)(nil)

// ConformityGitHubExec returns a deterministic GitHub fake for conformity tests.
func ConformityGitHubExec() cliexec.Exec {
	return &FakeExec{Responses: map[string][]byte{
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
	}}
}

// ConformityGitLabExec returns a deterministic GitLab fake for conformity tests.
func ConformityGitLabExec() cliexec.Exec {
	return &FakeExec{Responses: map[string][]byte{
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
	}}
}

// ConformityAzureExec returns a deterministic Azure DevOps fake for conformity tests.
func ConformityAzureExec() cliexec.Exec {
	return &FakeExec{Responses: map[string][]byte{
		"account show": []byte(`{"id":"sub"}`),
		"repos show":   []byte(`{"defaultBranch":"refs/heads/main"}`),
		"ref list": []byte(`[
			{"name":"refs/heads/main"},
			{"name":"refs/heads/feat/z"}
		]`),
		"pipelines runs list": []byte(`[]`),
		"pr list":             []byte(`[]`),
	}}
}

// ConformityBitbucketExec returns a deterministic Bitbucket fake for conformity tests.
// Callers must set BITBUCKET_TOKEN (or username/app password) for AuthStatus.
func ConformityBitbucketExec() cliexec.Exec {
	return &FakeExec{Responses: map[string][]byte{
		"repositories/acme/app/refs/branches": []byte(`{"values":[
			{"name":"main","target":{"date":"2026-01-01T00:00:00Z"}},
			{"name":"feat/b","target":{"date":"2026-01-02T00:00:00Z"}}
		]}`),
		"repositories/acme/app/pullrequests?state=OPEN":   []byte(`{"values":[]}`),
		"repositories/acme/app/pullrequests?state=MERGED": []byte(`{"values":[]}`),
		"repositories/acme/app/pipelines/":                []byte(`{"values":[]}`),
		"repositories/acme/app":                           []byte(`{"mainbranch":{"name":"main"}}`),
	}}
}
