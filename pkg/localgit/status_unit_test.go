package localgit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInspectPathWithFakeExec(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --git-dir": []byte(".git\n"),
			"worktree list --porcelain": []byte(
				"worktree " + dir + "\nHEAD abcdef\nbranch refs/heads/main\n\n",
			),
			"rev-parse --abbrev-ref HEAD":        []byte("main\n"),
			"status --porcelain":                 []byte(""),
			"rev-parse --abbrev-ref @{upstream}": []byte("origin/main\n"),
			"rev-list --left-right --count":      []byte("0\t0\n"),
		},
	}
	in := NewInspector(fk)
	st := in.InspectPath(context.Background(), dir)
	if !st.Mapped || st.Error != "" {
		t.Fatalf("status: %+v", st)
	}
	if st.Branch != "main" {
		t.Fatalf("branch=%q", st.Branch)
	}
	if st.Dirty {
		t.Fatalf("want clean tree")
	}
	if len(st.Worktrees) != 1 {
		t.Fatalf("worktrees=%d", len(st.Worktrees))
	}
}

func TestLooksLikeGitDir(t *testing.T) {
	if !looksLikeGitDir("/repo/.git/modules/providers/pkg") {
		t.Fatal("modules git dir")
	}
	if !looksLikeGitDir("/repo/.git") {
		t.Fatal(".git")
	}
	if looksLikeGitDir("/repo/providers/pkg") {
		t.Fatal("checkout is not a git dir")
	}
}

func TestInspectPathResolvesSubmoduleGitDir(t *testing.T) {
	checkout := t.TempDir()
	gitdir := filepath.Join(checkout, ".git", "modules", "prov")
	if err := os.MkdirAll(gitdir, 0o750); err != nil {
		t.Fatal(err)
	}
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --git-dir": []byte(".git\n"),
			"worktree list --porcelain": []byte(
				"worktree " + gitdir + "\nHEAD abcdef\nbranch refs/heads/main\n\n",
			),
			"--show-toplevel":                    []byte(checkout + "\n"),
			"rev-parse --abbrev-ref HEAD":        []byte("main\n"),
			"status --porcelain":                 []byte(""),
			"rev-parse --abbrev-ref @{upstream}": []byte("origin/main\n"),
			"rev-list --left-right --count":      []byte("0\t0\n"),
		},
	}
	in := NewInspector(fk)
	st := in.InspectPath(context.Background(), checkout)
	if !st.Mapped || st.Error != "" {
		t.Fatalf("status: %+v", st)
	}
	if len(st.Worktrees) != 1 {
		t.Fatalf("worktrees=%d", len(st.Worktrees))
	}
	want := checkout
	if resolved, err := filepath.EvalSymlinks(checkout); err == nil {
		want = resolved
	}
	if st.Worktrees[0].Path != want {
		t.Fatalf("path=%q want %q", st.Worktrees[0].Path, want)
	}
}

func TestInspectPathNotARepoWithFakeExec(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		errors: map[string]error{
			"rev-parse --git-dir": fmtError("not a git repository"),
		},
	}
	in := NewInspector(fk)
	st := in.InspectPath(context.Background(), dir)
	if !st.Mapped || st.Error != "not a git repository" {
		t.Fatalf("status: %+v", st)
	}
}

func TestEnrichOriginSyncWithFakeExec(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"for-each-ref --format=" + refFormat + " refs/heads/ refs/remotes/origin/": []byte(refLines(
				"refs/heads/main\x00\x00\x00",
				"refs/heads/feature\x00refs/remotes/origin/feature\x00ahead 2\x00",
				"refs/remotes/origin/HEAD\x00\x00\x00refs/remotes/origin/main",
				"refs/remotes/origin/main\x00\x00\x00",
				"refs/remotes/origin/feature\x00\x00\x00",
			)),
			"rev-list --left-right --count refs/remotes/origin/main...refs/heads/main": []byte("1\t0\n"),
		},
	}

	in := NewInspector(fk)
	st := Status{Mapped: true, Path: dir}
	in.EnrichOriginSync(context.Background(), &st)
	if st.Error != "" {
		t.Fatalf("error: %s calls=%v", st.Error, fk.calls)
	}
	if st.DefaultBranch != "main" {
		t.Fatalf("default=%q calls=%v", st.DefaultBranch, fk.calls)
	}
	if st.DefaultBehind != 1 || st.DefaultAhead != 0 {
		t.Fatalf("default ahead=%d behind=%d sync=%+v", st.DefaultAhead, st.DefaultBehind, st.OriginSync)
	}
	byName := map[string]BranchSync{}
	for _, s := range st.OriginSync {
		byName[s.Name] = s
	}
	if byName["feature"].Ahead != 2 || byName["feature"].Behind != 0 {
		t.Fatalf("feature sync=%+v", byName["feature"])
	}
}

func TestNilMutatorRefusesWrites(t *testing.T) {
	var m *Mutator
	_, err := m.FetchOriginSmart(context.Background(), filepath.Join(t.TempDir(), "repo"), FetchPolicy{TTL: time.Minute, Fresh: true}, nil)
	if !errors.Is(err, ErrLeaseNotHeld) {
		t.Fatalf("want ErrLeaseNotHeld, got %v", err)
	}
}

type fmtError string

func (e fmtError) Error() string { return string(e) }
