package dashboard

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

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

func (f *syncLocalFake) FetchOriginCached(context.Context, string, time.Duration, bool, *localgit.OriginFetchCache) error {
	return nil
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

func (f *syncLocalFake) ContentOnDefault(context.Context, string, string, string) (bool, string, error) {
	return false, "", nil
}

func (f *syncLocalFake) InspectSync(context.Context, string, string) (localgit.SyncInspection, error) {
	return f.result, nil
}

func TestInvestigateSyncValidatesMappedPathAndReturnsEvidence(t *testing.T) {
	path := filepath.Clean(t.TempDir())
	fake := &syncLocalFake{
		result: localgit.SyncInspection{
			Path:        path,
			Branch:      "main",
			Relation:    "diverged",
			AheadCount:  1,
			BehindCount: 2,
		},
	}
	service := New(nil, nil, fake)
	commands := NewCommands(service)
	doc := config.File{
		Projects: []config.Project{{
			ID:        "gh-app",
			Host:      config.HostGitHub,
			Path:      "acme/app",
			LocalPath: path,
		}},
	}

	got, err := commands.InvestigateSync(context.Background(), doc, SyncInvestigationRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  path,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ProjectID != "gh-app" || got.Result.Relation != "diverged" {
		t.Fatalf("result: %+v", got)
	}

	_, err = commands.InvestigateSync(context.Background(), doc, SyncInvestigationRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  filepath.Join(path, "other"),
	})
	if err == nil || !IsBadRequest(err) {
		t.Fatalf("want mapped-path validation error, got %v", err)
	}
}
