package dashboard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// argvCounter is fakeExec that records every argv, so tests can count git
// processes.
type argvCounter struct {
	fakeExec
	mu    sync.Mutex
	calls []string
}

func (f *argvCounter) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	f.mu.Unlock()
	return f.fakeExec.Run(ctx, name, args...)
}

func (f *argvCounter) count(substr string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if strings.Contains(c, substr) {
			n++
		}
	}
	return n
}

// repoFake answers git for one repository with a main worktree, two linked
// worktrees and branches b00..b(n-1), each tracking origin/<same name>.
func repoFake(t *testing.T, branches int) (*argvCounter, string, []string) {
	t.Helper()
	root := t.TempDir()
	common := filepath.Join(root, "main", ".git")
	paths := []string{filepath.Join(root, "main"), filepath.Join(root, "wt1"), filepath.Join(root, "wt2")}
	for i := range paths {
		paths[i] = mustDir(t, paths[i])
	}
	var heads, remotes, refs, list strings.Builder
	remotes.WriteString("origin\n")
	refs.WriteString("refs/remotes/origin/HEAD\x00\x00\x00refs/remotes/origin/b00\n")
	for i := 0; i < branches; i++ {
		name := fmt.Sprintf("b%02d", i)
		fmt.Fprintf(&heads, "%s\n", name)
		fmt.Fprintf(&remotes, "origin/%s\n", name)
		fmt.Fprintf(&refs, "refs/heads/%s\x00refs/remotes/origin/%s\x00behind 1\x00\x001 0\n", name, name)
		fmt.Fprintf(&refs, "refs/remotes/origin/%s\x00\x00\x00\x00\n", name)
	}
	for i, p := range paths {
		fmt.Fprintf(&list, "worktree %s\nHEAD abc\nbranch refs/heads/b%02d\n\n", p, i)
	}
	fk := &argvCounter{fakeExec: fakeExec{responses: map[string][]byte{
		"auth status":                                        []byte(""),
		"--git-common-dir":                                   []byte(common + "\n"),
		"rev-parse --git-dir":                                []byte(".git\n"),
		"worktree list --porcelain":                          []byte(list.String()),
		"rev-parse --abbrev-ref HEAD":                        []byte("b00\n"),
		"status --porcelain":                                 []byte(""),
		"@{upstream}":                                        []byte("origin/b00\n"),
		"rev-list --left-right":                              []byte("1\t0\n"),
		"fetch --prune origin":                               []byte(""),
		"symbolic-ref --quiet refs/remotes/origin/HEAD":      []byte("refs/remotes/origin/b00\n"),
		"for-each-ref --format=%(refname:short) refs/heads/": []byte(heads.String()),
		"for-each-ref --format=%(refname:short) refs/remotes/origin/": []byte(remotes.String()),
		"for-each-ref --format=%(refname)%00":                         []byte(refs.String()),
		"-C . version":                                                []byte("git version 2.46.2\n"),
	}}}
	return fk, common, paths
}

func mustDir(t *testing.T, p string) string {
	t.Helper()
	if err := os.MkdirAll(p, 0o750); err != nil {
		t.Fatal(err)
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}

// TestAttachLocalSharesOriginSyncPerRepo maps one repository twice (the
// scanned main worktree and an explicit local_path on a linked worktree).
// Both checkouts share one fetch and one origin sync: a single for-each-ref
// for 20 tracked branches, and no per-branch rev-list.
func TestAttachLocalSharesOriginSyncPerRepo(t *testing.T) {
	fk, common, paths := repoFake(t, 20)
	s := newTestService(fk, NewLocalGit(fk))
	p := boardconfig.Project{ID: "gh-app", Host: boardconfig.HostGitHub, Path: "acme/app", LocalPath: paths[1]}
	disc := localgit.Discovery{ByKey: map[string][]localgit.Checkout{
		localgit.TrackKey(p.Host, p.Path): {{Path: paths[0], CommonGitDir: common}},
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	local := s.attachLocal(ctx, p, localInputs{disc: disc, fetchTTL: time.Minute})
	if local == nil || len(local.Appearances) != 2 {
		t.Fatalf("local=%+v", local)
	}
	for _, app := range local.Appearances {
		if app.Error != "" || len(app.OriginSync) != 20 || app.DefaultBranch != "b00" || app.DefaultBehind != 1 {
			t.Fatalf("appearance %s: err=%q sync=%d default=%q behind=%d", app.Path, app.Error, len(app.OriginSync), app.DefaultBranch, app.DefaultBehind)
		}
	}
	if n := fk.count("for-each-ref --format=%(refname)%00"); n != 1 {
		t.Fatalf("origin sync for-each-ref ran %d times, want once per repository", n)
	}
	if n := fk.count("fetch --prune origin"); n != 1 {
		t.Fatalf("fetch ran %d times, want once per repository", n)
	}
	if n := fk.count("--git-common-dir"); n != 2 {
		// One for the explicit checkout's key, one for the lease bind.
		t.Fatalf("common dir resolved %d times, want 2", n)
	}
	// Only the per-worktree upstream counts (3 worktrees, inspected from
	// each of the 2 checkouts) still use rev-list.
	if n := fk.count("rev-list"); n != 6 {
		t.Fatalf("rev-list ran %d times, want 6", n)
	}
	t.Logf("git processes: %d (81 before sharing origin sync per repository)", len(fk.calls))
}
