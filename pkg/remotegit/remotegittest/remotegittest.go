// Package remotegittest is test support for pkg/remotegit: a canned-output
// cliexec.Exec, a canned-response HTTP handler and deterministic forge
// fixtures for conformity and benchmark tests. It is not part of the
// remotegit runtime API.
package remotegittest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
)

// FakeExec returns canned CLI output matched by argv substrings (longest match wins).
// An argv with no match fails, so tests notice unexpected forge calls.
type FakeExec struct {
	Responses map[string][]byte
	Errors    map[string]error
}

// LookPath reports every tool as installed under /fake.
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

// Run returns the canned output (or error) for the longest matching argv substring.
func (f *FakeExec) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	return f.match(name, args...)
}

// RunJSON is Run with empty output mapped to an empty JSON array.
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

// FakeHTTP serves canned forge API bodies matched by request URI substring
// (longest match wins). An unmatched request gets 404, so tests notice
// unexpected calls. SetStatus makes a matched key answer with an HTTP status
// instead of its body. Safe for concurrent use.
type FakeHTTP struct {
	mu        sync.Mutex
	responses map[string][]byte
	statuses  map[string]int
	requests  []string
}

// NewFakeHTTP returns a fake serving responses keyed by URI substring.
func NewFakeHTTP(responses map[string][]byte) *FakeHTTP {
	return &FakeHTTP{responses: responses, statuses: map[string]int{}}
}

// SetStatus makes requests matching key answer with status and no body.
func (f *FakeHTTP) SetStatus(key string, status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statuses[key] = status
}

// ClearStatus restores the canned body for key.
func (f *FakeHTTP) ClearStatus(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.statuses, key)
}

// Requests returns the request URIs served so far, in order.
func (f *FakeHTTP) Requests() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.requests...)
}

// ServeHTTP answers with the canned body or status for the longest key that
// is a substring of the request URI.
func (f *FakeHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()
	f.mu.Lock()
	f.requests = append(f.requests, uri)
	var best string
	for key := range f.responses {
		if strings.Contains(uri, key) && len(key) >= len(best) {
			best = key
		}
	}
	for key := range f.statuses {
		if strings.Contains(uri, key) && len(key) >= len(best) {
			best = key
		}
	}
	status, failed := f.statuses[best]
	body, ok := f.responses[best]
	f.mu.Unlock()
	switch {
	case best == "":
		http.Error(w, "unexpected request: "+uri, http.StatusNotFound)
	case failed:
		w.WriteHeader(status)
	case ok:
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}
}

// ConformityBitbucketHTTP returns a deterministic Bitbucket API fake for
// conformity tests. Serve it with httptest and point
// remotegit.BitbucketConfig.BaseURL at the server.
func ConformityBitbucketHTTP() *FakeHTTP {
	return NewFakeHTTP(map[string][]byte{
		"repositories/acme/app/refs/branches": []byte(`{"values":[
			{"name":"main","target":{"date":"2026-01-01T00:00:00Z"}},
			{"name":"feat/b","target":{"date":"2026-01-02T00:00:00Z"}}
		]}`),
		"repositories/acme/app/pullrequests?state=OPEN":   []byte(`{"values":[]}`),
		"repositories/acme/app/pullrequests?state=MERGED": []byte(`{"values":[]}`),
		"repositories/acme/app/pipelines/":                []byte(`{"values":[]}`),
		"repositories/acme/app":                           []byte(`{"mainbranch":{"name":"main"}}`),
	})
}
