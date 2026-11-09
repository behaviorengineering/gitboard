package dashboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

type pullLocalFake struct {
	path     string
	pullErr  error
	indexErr error
}

func (f *pullLocalFake) ScanRoots(context.Context, []string) localgit.Discovery {
	return localgit.Discovery{}
}

func (f *pullLocalFake) InspectPath(_ context.Context, path string) localgit.Status {
	return localgit.Status{Mapped: true, Path: path, Branch: "main"}
}

func (f *pullLocalFake) OriginSync(context.Context, string) (localgit.RepoSync, error) {
	return localgit.RepoSync{}, nil
}

func (f *pullLocalFake) CommonGitDir(_ context.Context, path string) (string, error) {
	return path, nil
}

func (f *pullLocalFake) ListLocalHeads(context.Context, string) ([]string, error) {
	return nil, nil
}

func (f *pullLocalFake) FetchOriginSmart(context.Context, string, localgit.FetchPolicy, *localgit.OriginFetchCache) (bool, error) {
	return false, nil
}

func (f *pullLocalFake) OriginRemote(context.Context, string) (string, error) {
	return "https://example.test/acme/app.git", nil
}

func (f *pullLocalFake) PullFFOnly(context.Context, string, string) (localgit.PullPhases, error) {
	return localgit.PullPhases{}, f.pullErr
}

func (f *pullLocalFake) RemoveSafeCheckout(context.Context, string, string, string) error {
	return nil
}

func (f *pullLocalFake) EnsureWritableIndex(context.Context, string, bool) error {
	return f.indexErr
}

func (f *pullLocalFake) ContentOnDefault(context.Context, string, string, string) (bool, string, error) {
	return false, "", nil
}

func (f *pullLocalFake) CompareSync(context.Context, string, string) (localgit.SyncInspection, error) {
	return localgit.SyncInspection{}, fmt.Errorf("not used")
}

func (f *pullLocalFake) CanonicalCheckoutPath(_ context.Context, path string) string {
	return path
}

func (f *pullLocalFake) Mutator(*localgit.MutationLease) (LocalMutator, error) {
	return f, nil
}

func TestPullFFTreatsUpToDateAsSuccess(t *testing.T) {
	path := filepath.Clean(t.TempDir())
	fake := &pullLocalFake{
		path:    path,
		pullErr: fmt.Errorf("%w: %q", localgit.ErrUpToDate, "main"),
	}
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: path,
		}},
	}

	got, err := commands.PullFF(context.Background(), doc, PullFFRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  path,
	})
	if err != nil {
		t.Fatalf("up-to-date want success, got %v", err)
	}
	if !got.OK || got.Branch != "main" {
		t.Fatalf("result: %+v", got)
	}
	if got.Path == "" {
		t.Fatal("expected resolved path")
	}
}

func TestPullFFStillRejectsDiverged(t *testing.T) {
	path := filepath.Clean(t.TempDir())
	fake := &pullLocalFake{
		path:    path,
		pullErr: fmt.Errorf("%w: %q", localgit.ErrDiverged, "main"),
	}
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: path,
		}},
	}

	_, err := commands.PullFF(context.Background(), doc, PullFFRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  path,
	})
	if err == nil || !gberr.HasCode(err, gberr.CodeConflict) {
		t.Fatalf("diverged want conflict, got %v", err)
	}
}

func TestPullFFStaleIndexLockNeedsConfirm(t *testing.T) {
	path := filepath.Clean(t.TempDir())
	fake := &pullLocalFake{
		path: path,
		indexErr: &localgit.StaleIndexLockError{
			RepoPath: path,
			LockPath: filepath.Join(path, ".git", "index.lock"),
			Age:      2 * time.Minute,
		},
	}
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: path,
		}},
	}

	_, err := commands.PullFF(context.Background(), doc, PullFFRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  path,
	})
	if !IsConfirmRequired(err) {
		t.Fatalf("want confirm required, got %v", err)
	}
	var cre ConfirmRequiredError
	if !errors.As(err, &cre) || cre.Code != ConfirmStaleIndexLock {
		t.Fatalf("confirm payload: %+v", err)
	}
}

func TestPullFFClearsStaleIndexLockWhenApproved(t *testing.T) {
	path := filepath.Clean(t.TempDir())
	fake := &pullLocalFake{
		path:    path,
		pullErr: nil,
	}
	var cleared bool
	fake.indexErr = nil
	// First call without clear is not used; with clear EnsureWritableIndex succeeds.
	service := newTestService(&fakeExec{}, &pullClearFake{pullLocalFake: *fake, cleared: &cleared})
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: path,
		}},
	}

	got, err := commands.PullFF(context.Background(), doc, PullFFRequest{
		ProjectID:      "gh-app",
		Branch:         "main",
		RepoPath:       path,
		ClearIndexLock: true,
	})
	if err != nil {
		t.Fatalf("clear+pull: %v", err)
	}
	if !got.OK || !cleared {
		t.Fatalf("got=%+v cleared=%v", got, cleared)
	}
}

type pullClearFake struct {
	pullLocalFake
	cleared *bool
}

func (f *pullClearFake) Mutator(*localgit.MutationLease) (LocalMutator, error) {
	return f, nil
}

func (f *pullClearFake) EnsureWritableIndex(_ context.Context, _ string, clearStale bool) error {
	if clearStale && f.cleared != nil {
		*f.cleared = true
	}
	return nil
}

type pullTimedFake struct {
	pullLocalFake
	phases localgit.PullPhases
	err    error
}

func (f *pullTimedFake) PullFFOnly(context.Context, string, string) (localgit.PullPhases, error) {
	return f.phases, f.err
}

func (f *pullTimedFake) Mutator(*localgit.MutationLease) (LocalMutator, error) {
	return f, nil
}

func TestPullFFReturnsPhaseTimings(t *testing.T) {
	path := filepath.Clean(t.TempDir())
	fake := &pullTimedFake{
		phases: localgit.PullPhases{FetchMs: 12, SubmodulePreMs: 3, MergeMs: 4, SubmodulePostMs: 5},
	}
	fake.pullErr = nil
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: path,
		}},
	}

	got, err := commands.PullFF(context.Background(), doc, PullFFRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  path,
	})
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if !got.OK {
		t.Fatalf("result: %+v", got)
	}
	if got.FetchMs != 12 || got.SubmodulePreMs != 3 || got.MergeMs != 4 || got.SubmodulePostMs != 5 {
		t.Fatalf("phases: %+v", got)
	}
}

func TestPullFFUpToDateKeepsPhaseTimings(t *testing.T) {
	path := filepath.Clean(t.TempDir())
	fake := &pullTimedFake{
		phases: localgit.PullPhases{FetchMs: 7},
		err:    fmt.Errorf("%w: %q", localgit.ErrUpToDate, "main"),
	}
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: path,
		}},
	}

	got, err := commands.PullFF(context.Background(), doc, PullFFRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  path,
	})
	if err != nil {
		t.Fatalf("up-to-date want success, got %v", err)
	}
	if !got.OK || got.FetchMs != 7 {
		t.Fatalf("result: %+v", got)
	}
}
