package dashboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

func TestFindWorktree(t *testing.T) {
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
			name: "hint is left to pruneVerdict",
			local: &board.LocalStatus{
				Worktrees: []board.LocalWorktree{{
					Path:      abs,
					Branch:    "feature",
					PruneHint: board.PruneLikely,
				}},
			},
			branch: "feature",
			path:   abs,
			wantOK: true,
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
			wt, ok := findWorktree(tt.local, tt.branch, tt.path, nil)
			if ok != tt.wantOK {
				t.Fatalf("ok=%v want %v", ok, tt.wantOK)
			}
			if tt.wantOK && wt.Path != abs {
				t.Fatalf("path=%q want %q", wt.Path, abs)
			}
		})
	}
}

func TestFindWorktreeGitDirVsCheckout(t *testing.T) {
	checkout := filepath.Clean("/tmp/providers/pkg")
	gitdir := filepath.Clean("/tmp/parent/.git/modules/providers/pkg")
	canon := func(p string) string {
		p = filepath.Clean(p)
		if p == gitdir {
			return checkout
		}
		return p
	}
	local := &board.LocalStatus{
		Worktrees: []board.LocalWorktree{{
			Path:      gitdir,
			Branch:    "feat/ai-copilots",
			PruneHint: board.PruneSafe,
		}},
	}
	wt, ok := findWorktree(local, "feat/ai-copilots", checkout, canon)
	if !ok || wt.Path != checkout {
		t.Fatalf("gitdir must match checkout: ok=%v path=%q", ok, wt.Path)
	}
}

func TestInvalidCauseKeepsChain(t *testing.T) {
	cause := ErrLocalInspectorMissing
	err := invalidCause("dashboard.test", "bad input", cause)
	if !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("want invalid, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("want unwrap to %v, got %v", cause, err)
	}
	if e, ok := gberr.As(err); !ok || e.Msg != "bad input" {
		t.Fatalf("client message: %+v", e)
	}
	if gberr.HasCode(nil, gberr.CodeInvalid) {
		t.Fatal("nil has no code")
	}
}

func TestGitMutationErrCodes(t *testing.T) {
	cases := []struct {
		err  error
		want gberr.Code
	}{
		{fmt.Errorf("%w: %q", localgit.ErrInvalidBranch, "x"), gberr.CodeInvalid},
		{fmt.Errorf("%w: %q", localgit.ErrMissingBranch, "x"), gberr.CodeInvalid},
		{fmt.Errorf("%w at /home/me/repo", localgit.ErrDirtyTree), gberr.CodeConflict},
		{fmt.Errorf("%w: %q", localgit.ErrDiverged, "x"), gberr.CodeConflict},
		{fmt.Errorf("%w: %w", localgit.ErrMutationCanceled, context.Canceled), gberr.CodeCanceled},
		{errors.New("git merge: fatal: /home/me/repo stderr"), gberr.CodeUnavailable},
	}
	for _, tc := range cases {
		err := gitMutationErr("dashboard.test", tc.err)
		if got := gberr.CodeOf(err); got != tc.want {
			t.Errorf("gitMutationErr(%v) code = %q, want %q", tc.err, got, tc.want)
		}
		if !errors.Is(err, tc.err) {
			t.Errorf("gitMutationErr(%v) lost the cause", tc.err)
		}
		if e, _ := gberr.As(err); strings.Contains(e.Msg, "/home/me") {
			t.Errorf("client message leaks the cause: %q", e.Msg)
		}
	}
}

func TestClientForNilInterface(t *testing.T) {
	s := newTestService(&fakeExec{}, nil)
	if ClientFor(s, boardconfig.Project{Host: boardconfig.Host("gitea")}) != nil {
		t.Fatal("unsupported host should be nil interface")
	}
	if ClientFor(s, boardconfig.Project{Host: boardconfig.HostGitHub}) == nil {
		t.Fatal("GitHub client should be wired")
	}
	if ClientFor(nil, boardconfig.Project{Host: boardconfig.HostGitHub}) != nil {
		t.Fatal("nil service should return nil")
	}
}

func TestPruneSafeValidation(t *testing.T) {
	fk := &fakeExecForPrune{}
	s := newTestService(fk, nil)
	cmds := NewCommands(s)
	err := cmds.PruneSafe(t.Context(), boardconfig.File{}, PruneSafeRequest{
		ProjectID: "p", Branch: "b", WorktreePath: "/tmp/x",
	})
	if !errors.Is(err, ErrLocalInspectorMissing) {
		t.Fatalf("got %v", err)
	}

	s = newTestService(fk, NewLocalGit(fk))
	cmds = NewCommands(s)
	err = cmds.PruneSafe(t.Context(), boardconfig.File{}, PruneSafeRequest{})
	if !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("empty request want bad request, got %v", err)
	}
	err = cmds.PruneSafe(t.Context(), boardconfig.File{Projects: []boardconfig.Project{{ID: "p"}}}, PruneSafeRequest{
		ProjectID: "missing", Branch: "feature", WorktreePath: "/tmp/x",
	})
	if !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("unknown project want bad request, got %v", err)
	}
	err = cmds.PruneSafe(t.Context(), boardconfig.File{Projects: []boardconfig.Project{{ID: "p"}}}, PruneSafeRequest{
		ProjectID: "p", Branch: "evil:ref", WorktreePath: "/tmp/x",
	})
	if !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("invalid branch want bad request, got %v", err)
	}
}

func TestPullFFValidation(t *testing.T) {
	fk := &fakeExecForPrune{}
	s := newTestService(fk, nil)
	cmds := NewCommands(s)
	_, err := cmds.PullFF(t.Context(), boardconfig.File{}, PullFFRequest{
		ProjectID: "p", Branch: "main", RepoPath: "/tmp/x",
	})
	if !errors.Is(err, ErrLocalInspectorMissing) {
		t.Fatalf("got %v", err)
	}

	s = newTestService(fk, NewLocalGit(fk))
	cmds = NewCommands(s)
	_, err = cmds.PullFF(t.Context(), boardconfig.File{}, PullFFRequest{})
	if !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("empty request want bad request, got %v", err)
	}
	_, err = cmds.PullFF(t.Context(), boardconfig.File{Projects: []boardconfig.Project{{ID: "p"}}}, PullFFRequest{
		ProjectID: "missing", Branch: "main", RepoPath: "/tmp/x",
	})
	if !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("unknown project want bad request, got %v", err)
	}
	_, err = cmds.PullFF(t.Context(), boardconfig.File{Projects: []boardconfig.Project{{ID: "p"}}}, PullFFRequest{
		ProjectID: "p", Branch: "evil:ref", RepoPath: "/tmp/x",
	})
	if !gberr.HasCode(err, gberr.CodeInvalid) {
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
