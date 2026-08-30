package localgit

import (
	"context"
	"path/filepath"
	"testing"
)

func TestInspectPathWithFakeExec(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --git-dir": []byte(".git\n"),
			"worktree list --porcelain": []byte(
				"worktree " + dir + "\nHEAD abcdef\nbranch refs/heads/main\n\n",
			),
			"rev-parse --abbrev-ref HEAD": []byte("main\n"),
			"status --porcelain":         []byte(""),
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
			"symbolic-ref --quiet refs/remotes/origin/HEAD":              []byte("refs/remotes/origin/main\n"),
			"for-each-ref --format=%(refname:short) refs/heads/":         []byte("main\nfeature\n"),
			"for-each-ref --format=%(refname:short) refs/remotes/origin/": []byte("origin/main\norigin/feature\n"),
			"rev-list --left-right --count refs/remotes/origin/main...refs/heads/main":       []byte("1\t0\n"),
			"rev-list --left-right --count refs/remotes/origin/feature...refs/heads/feature": []byte("0\t2\n"),
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

func TestFetchOriginNilInspector(t *testing.T) {
	var in *Inspector
	if err := in.FetchOrigin(context.Background(), filepath.Join(t.TempDir(), "repo")); err == nil {
		t.Fatal("want inspector missing error")
	}
}

type fmtError string

func (e fmtError) Error() string { return string(e) }
