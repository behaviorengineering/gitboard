//go:build integration

package localgit_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

func TestFetchOriginCachedRefreshesStaleTracking(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupFetchRepos(t)

	runGit(t, origin, "commit", "--allow-empty", "-m", "origin-ahead")

	in := localgit.NewInspector(cliexec.New())
	ctx := context.Background()

	st := in.InspectPath(ctx, local)
	in.EnrichOriginSync(ctx, &st)
	if st.DefaultBehind != 0 || st.DefaultAhead != 0 {
		t.Fatalf("before fetch want in-sync vs stale origin, ahead=%d behind=%d", st.DefaultAhead, st.DefaultBehind)
	}

	cache := localgit.NewOriginFetchCache()
	if err := in.FetchOriginCached(ctx, local, time.Minute, false, cache); err != nil {
		t.Fatalf("FetchOriginCached: %v", err)
	}

	st = in.InspectPath(ctx, local)
	in.EnrichOriginSync(ctx, &st)
	if st.Error != "" {
		t.Fatalf("status: %+v", st)
	}
	if st.DefaultBehind != 1 || st.DefaultAhead != 0 {
		t.Fatalf("after fetch ahead=%d behind=%d", st.DefaultAhead, st.DefaultBehind)
	}
}

func TestFetchOriginCachedRespectsTTL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	_, local := setupFetchRepos(t)

	var fetches atomic.Int32
	run := &countingExec{inner: cliexec.New(), fetches: &fetches}
	in := localgit.NewInspector(run)
	ctx := context.Background()
	cache := localgit.NewOriginFetchCache()
	now := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	cache.SetNow(func() time.Time { return now })

	if err := in.FetchOriginCached(ctx, local, time.Minute, false, cache); err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	if got := fetches.Load(); got != 1 {
		t.Fatalf("fetches after first=%d", got)
	}

	if err := in.FetchOriginCached(ctx, local, time.Minute, false, cache); err != nil {
		t.Fatalf("second fetch within TTL: %v", err)
	}
	if got := fetches.Load(); got != 1 {
		t.Fatalf("want TTL skip, fetches=%d", got)
	}

	now = now.Add(2 * time.Minute)
	if err := in.FetchOriginCached(ctx, local, time.Minute, false, cache); err != nil {
		t.Fatalf("third fetch after TTL: %v", err)
	}
	if got := fetches.Load(); got != 2 {
		t.Fatalf("want fetch after TTL, fetches=%d", got)
	}

	if err := in.FetchOriginCached(ctx, local, time.Minute, true, cache); err != nil {
		t.Fatalf("fresh fetch: %v", err)
	}
	if got := fetches.Load(); got != 3 {
		t.Fatalf("want fresh bypass, fetches=%d", got)
	}
}

func TestOriginFetchCacheClear(t *testing.T) {
	cache := localgit.NewOriginFetchCache()
	cache.MarkSuccess("/tmp/repo.git")
	if cache.NeedsFetch("/tmp/repo.git", time.Hour, false) {
		t.Fatal("want cache hit")
	}
	cache.Clear()
	if !cache.NeedsFetch("/tmp/repo.git", time.Hour, false) {
		t.Fatal("want needs fetch after clear")
	}
}

func TestInvalidateOriginSync(t *testing.T) {
	st := localgit.Status{
		DefaultAhead:  2,
		DefaultBehind: 3,
		OriginSync: []localgit.BranchSync{
			{Name: "main", Ahead: 2, Behind: 3},
		},
	}
	localgit.InvalidateOriginSync(&st)
	if st.DefaultAhead != 0 || st.DefaultBehind != 0 || st.OriginSync != nil {
		t.Fatalf("want sync cleared, got %+v", st)
	}
	localgit.InvalidateOriginSync(nil) // must not panic
}

type countingExec struct {
	inner   cliexec.Exec
	fetches *atomic.Int32
}

func (c *countingExec) LookPath(name string) (string, error) {
	return c.inner.LookPath(name)
}

func (c *countingExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	if name == "git" && strings.Contains(joined, "fetch --prune origin") {
		c.fetches.Add(1)
	}
	return c.inner.Run(ctx, name, args...)
}

func (c *countingExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	return c.inner.RunJSON(ctx, name, args...)
}

func setupFetchRepos(t *testing.T) (origin, local string) {
	t.Helper()
	root := t.TempDir()
	origin = filepath.Join(root, "origin")
	local = filepath.Join(root, "local")
	if err := os.MkdirAll(origin, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, origin, "init", "-b", "main")
	runGit(t, origin, "config", "user.email", "t@example.com")
	runGit(t, origin, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(origin, "README"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, origin, "add", "README")
	runGit(t, origin, "commit", "-m", "init")
	runGit(t, root, "clone", origin, local)
	return origin, local
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v (dir=%s): %v\n%s", args, dir, err, out)
	}
}
