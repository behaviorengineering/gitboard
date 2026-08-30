package dashboard

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

func TestFindSafeWorktree(t *testing.T) {
	abs := filepath.Clean("/tmp/wt-feature")
	tests := []struct {
		name   string
		local  *board.LocalStatus
		branch string
		path   string
		wantOK bool
	}{
		{
			name:   "nil local",
			branch: "feature",
			path:   abs,
			wantOK: false,
		},
		{
			name: "safe match",
			local: &board.LocalStatus{
				Worktrees: []board.LocalWorktree{{
					Path:      abs,
					Branch:    "feature",
					PruneHint: board.PruneSafe,
				}},
			},
			branch: "feature",
			path:   abs,
			wantOK: true,
		},
		{
			name: "likely not enough",
			local: &board.LocalStatus{
				Worktrees: []board.LocalWorktree{{
					Path:      abs,
					Branch:    "feature",
					PruneHint: board.PruneLikely,
				}},
			},
			branch: "feature",
			path:   abs,
			wantOK: false,
		},
		{
			name: "wrong branch",
			local: &board.LocalStatus{
				Worktrees: []board.LocalWorktree{{
					Path:      abs,
					Branch:    "other",
					PruneHint: board.PruneSafe,
				}},
			},
			branch: "feature",
			path:   abs,
			wantOK: false,
		},
		{
			name: "wrong path",
			local: &board.LocalStatus{
				Worktrees: []board.LocalWorktree{{
					Path:      filepath.Clean("/tmp/other"),
					Branch:    "feature",
					PruneHint: board.PruneSafe,
				}},
			},
			branch: "feature",
			path:   abs,
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wt, ok := findSafeWorktree(tt.local, tt.branch, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("ok=%v want %v", ok, tt.wantOK)
			}
			if tt.wantOK && wt.Path != abs {
				t.Fatalf("path=%q want %q", wt.Path, abs)
			}
		})
	}
}

func TestIsBadRequest(t *testing.T) {
	if !IsBadRequest(badRequest("nope")) {
		t.Fatal("expected bad request")
	}
	if IsBadRequest(nil) {
		t.Fatal("nil should not be bad request")
	}
}

func TestClientForNilInterface(t *testing.T) {
	s := &Service{}
	client := ClientFor(s, config.Project{Host: config.HostGitHub})
	if client != nil {
		t.Fatal("missing GitHub client should be nil interface")
	}
	client = ClientFor(s, config.Project{Host: config.HostGitLab})
	if client != nil {
		t.Fatal("missing GitLab client should be nil interface")
	}
	if ClientFor(nil, config.Project{Host: config.HostGitHub}) != nil {
		t.Fatal("nil service should return nil")
	}
}

func TestPruneSafeValidation(t *testing.T) {
	fk := &fakeExecForPrune{}
	s := &Service{}
	cmds := NewCommands(s)
	err := cmds.PruneSafe(t.Context(), config.File{}, PruneSafeRequest{
		ProjectID: "p", Branch: "b", WorktreePath: "/tmp/x",
	})
	if err == nil || err.Error() != "local git inspector missing" {
		t.Fatalf("got %v", err)
	}

	s = &Service{Local: localgit.NewInspector(fk)}
	cmds = NewCommands(s)
	err = cmds.PruneSafe(t.Context(), config.File{}, PruneSafeRequest{})
	if !IsBadRequest(err) {
		t.Fatalf("empty request want bad request, got %v", err)
	}
	err = cmds.PruneSafe(t.Context(), config.File{Projects: []config.Project{{ID: "p"}}}, PruneSafeRequest{
		ProjectID: "missing", Branch: "feature", WorktreePath: "/tmp/x",
	})
	if !IsBadRequest(err) {
		t.Fatalf("unknown project want bad request, got %v", err)
	}
	err = cmds.PruneSafe(t.Context(), config.File{Projects: []config.Project{{ID: "p"}}}, PruneSafeRequest{
		ProjectID: "p", Branch: "evil:ref", WorktreePath: "/tmp/x",
	})
	if !IsBadRequest(err) {
		t.Fatalf("invalid branch want bad request, got %v", err)
	}
}

func TestPullFFValidation(t *testing.T) {
	fk := &fakeExecForPrune{}
	s := &Service{}
	cmds := NewCommands(s)
	_, err := cmds.PullFF(t.Context(), config.File{}, PullFFRequest{
		ProjectID: "p", Branch: "main", RepoPath: "/tmp/x",
	})
	if err == nil || err.Error() != "local git inspector missing" {
		t.Fatalf("got %v", err)
	}

	s = &Service{Local: localgit.NewInspector(fk)}
	cmds = NewCommands(s)
	_, err = cmds.PullFF(t.Context(), config.File{}, PullFFRequest{})
	if !IsBadRequest(err) {
		t.Fatalf("empty request want bad request, got %v", err)
	}
	_, err = cmds.PullFF(t.Context(), config.File{Projects: []config.Project{{ID: "p"}}}, PullFFRequest{
		ProjectID: "missing", Branch: "main", RepoPath: "/tmp/x",
	})
	if !IsBadRequest(err) {
		t.Fatalf("unknown project want bad request, got %v", err)
	}
	_, err = cmds.PullFF(t.Context(), config.File{Projects: []config.Project{{ID: "p"}}}, PullFFRequest{
		ProjectID: "p", Branch: "evil:ref", RepoPath: "/tmp/x",
	})
	if !IsBadRequest(err) {
		t.Fatalf("invalid branch want bad request, got %v", err)
	}
}

func TestRepoPathAllowed(t *testing.T) {
	abs := filepath.Clean("/tmp/repo")
	local := &board.LocalStatus{
		Mapped: true,
		Path:   abs,
		Appearances: []board.LocalAppearance{{
			Path: filepath.Clean("/tmp/other"),
		}},
	}
	if !repoPathAllowed(local, abs) {
		t.Fatal("primary path should be allowed")
	}
	if !repoPathAllowed(local, filepath.Clean("/tmp/other")) {
		t.Fatal("appearance path should be allowed")
	}
	if repoPathAllowed(local, filepath.Clean("/tmp/evil")) {
		t.Fatal("foreign path must be rejected")
	}
}

// fakeExecForPrune satisfies cliexec.Exec for validation-only tests.
type fakeExecForPrune struct{}

func (f *fakeExecForPrune) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *fakeExecForPrune) Run(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeExecForPrune) RunJSON(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
