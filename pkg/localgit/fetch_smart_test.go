package localgit

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testRepoPaths(t *testing.T) (repo, common string) {
	t.Helper()
	dir := t.TempDir()
	repo = filepath.Join(dir, "work")
	common = filepath.Join(dir, "repo.git")
	return repo, common
}

func TestFetchOriginBranchesEmptyIsNoop(t *testing.T) {
	repo, common := testRepoPaths(t)
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir": []byte(common + "\n"),
	}}
	m := leasedMutator(t, fx, common)
	ctx := context.Background()
	if err := m.fetchOriginBranches(ctx, repo, nil); err != nil {
		t.Fatalf("empty: %v", err)
	}
	if err := m.fetchOriginBranches(ctx, repo, []string{"", "  "}); err != nil {
		t.Fatalf("whitespace: %v", err)
	}
	for _, call := range fx.calls {
		if strings.Contains(call, "fetch") {
			t.Fatalf("empty target must not fetch: %v", fx.calls)
		}
	}
}

func TestFetchOriginSmartFullThenTargeted(t *testing.T) {
	repo, common := testRepoPaths(t)
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir":     []byte(common + "\n"),
		"fetch --prune origin": []byte(""),
		"fetch origin":         []byte(""),
	}}
	m := leasedMutator(t, fx, common)
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	cache := NewOriginFetchCache(func() time.Time { return now })
	ctx := context.Background()
	ttl := time.Minute

	if _, err := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: ttl, Branches: []string{"main", "feature/x"}}, cache); err != nil {
		t.Fatalf("cold smart: %v", err)
	}
	if !hasCallSubstr(fx.calls, "fetch --prune origin") {
		t.Fatalf("cold want full prune, calls=%v", fx.calls)
	}
	if hasCallSubstr(fx.calls, "fetch origin main") || hasCallSubstr(fx.calls, "fetch origin feature") {
		t.Fatalf("cold must not also targeted-fetch, calls=%v", fx.calls)
	}

	fx.calls = nil
	updated, err := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: ttl, Branches: []string{"main", "feature/x"}}, cache)
	if err != nil {
		t.Fatalf("warm within TTL: %v", err)
	}
	if updated {
		t.Fatal("warm within branch TTL must report updated=false")
	}
	if hasCallSubstr(fx.calls, "fetch") {
		t.Fatalf("warm within branch TTL should skip, calls=%v", fx.calls)
	}

	now = now.Add(2 * time.Minute)
	fx.calls = nil
	if _, err := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: ttl, Branches: []string{"main", "feature/x"}}, cache); err != nil {
		t.Fatalf("after full TTL: %v", err)
	}
	if !hasCallSubstr(fx.calls, "fetch --prune origin") {
		t.Fatalf("expired full TTL want prune, calls=%v", fx.calls)
	}
}

func TestFetchOriginSmartTargetedWhenFullWarm(t *testing.T) {
	repo, common := testRepoPaths(t)
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir":     []byte(common + "\n"),
		"fetch --prune origin": []byte(""),
		"fetch origin":         []byte(""),
	}}
	m := leasedMutator(t, fx, common)
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	cache := NewOriginFetchCache(func() time.Time { return now })
	ctx := context.Background()
	ttl := time.Minute

	if _, err := m.fetchOriginCached(ctx, repo, filepath.Clean(common), ttl, false, cache); err != nil {
		t.Fatalf("full: %v", err)
	}
	fx.calls = nil

	cache.MarkBranchesSuccess(filepath.Clean(common), []string{"main"})
	if _, err := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: ttl, Branches: []string{"main", "feature/x"}}, cache); err != nil {
		t.Fatalf("targeted: %v", err)
	}
	if hasCallSubstr(fx.calls, "fetch --prune origin") {
		t.Fatalf("warm full must not prune, calls=%v", fx.calls)
	}
	joined := strings.Join(fx.calls, "\n")
	if !strings.Contains(joined, "fetch origin") || !strings.Contains(joined, "feature/x") {
		t.Fatalf("want targeted feature/x, calls=%v", fx.calls)
	}
}

func TestFetchOriginSmartFreshForcesPrune(t *testing.T) {
	repo, common := testRepoPaths(t)
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir":     []byte(common + "\n"),
		"fetch --prune origin": []byte(""),
	}}
	m := leasedMutator(t, fx, common)
	cache := NewOriginFetchCache(time.Now)
	cache.MarkFullSuccess(filepath.Clean(common))
	cache.MarkBranchesSuccess(filepath.Clean(common), []string{"main"})
	ctx := context.Background()
	if _, err := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: time.Hour, Fresh: true, Branches: []string{"main"}}, cache); err != nil {
		t.Fatalf("fresh: %v", err)
	}
	if !hasCallSubstr(fx.calls, "fetch --prune origin") {
		t.Fatalf("fresh want prune, calls=%v", fx.calls)
	}
}

func TestFetchOriginSmartEmptyBranchesSkipsWhenWarm(t *testing.T) {
	repo, common := testRepoPaths(t)
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir":     []byte(common + "\n"),
		"fetch --prune origin": []byte(""),
	}}
	m := leasedMutator(t, fx, common)
	cache := NewOriginFetchCache(time.Now)
	cache.MarkFullSuccess(filepath.Clean(common))
	ctx := context.Background()
	fx.calls = nil
	if _, err := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: time.Hour}, cache); err != nil {
		t.Fatalf("empty warm: %v", err)
	}
	if hasCallSubstr(fx.calls, "fetch") {
		t.Fatalf("empty targets must not fall back to prune, calls=%v", fx.calls)
	}
}

func TestFetchOriginSmartCoalescesAndCoversLeftoverBranches(t *testing.T) {
	repo, common := testRepoPaths(t)
	var fetches atomic.Int32
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir":     []byte(common + "\n"),
		"fetch --prune origin": []byte(""),
		"fetch origin":         []byte(""),
	}}
	counting := &countingFetchExec{inner: fx, fetches: &fetches}
	m := leasedMutator(t, counting, common)
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	cache := NewOriginFetchCache(func() time.Time { return now })
	ctx := context.Background()
	ttl := time.Minute

	if _, err := m.fetchOriginCached(ctx, repo, filepath.Clean(common), ttl, false, cache); err != nil {
		t.Fatalf("seed full: %v", err)
	}
	fetches.Store(0)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, e1 := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: ttl, Branches: []string{"main"}}, cache)
		errCh <- e1
	}()
	go func() {
		defer wg.Done()
		_, e2 := m.FetchOriginSmart(ctx, repo, FetchPolicy{TTL: ttl, Branches: []string{"feature/x"}}, cache)
		errCh <- e2
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent: %v", err)
		}
	}
	if got := fetches.Load(); got < 1 {
		t.Fatalf("want at least one targeted fetch, got %d", got)
	}
	if stale := cache.StaleBranches(filepath.Clean(common), ttl, []string{"main", "feature/x"}); len(stale) != 0 {
		t.Fatalf("leftover branches still stale: %v (fetches=%d calls=%v)", stale, fetches.Load(), fx.calls)
	}
}

func TestOriginFetchCachePerBranchTTLIndependent(t *testing.T) {
	common := filepath.Clean(filepath.Join(t.TempDir(), "repo.git"))
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	cache := NewOriginFetchCache(func() time.Time { return now })
	cache.MarkFullSuccess(common)
	cache.MarkBranchesSuccess(common, []string{"main"})
	now = now.Add(30 * time.Second)
	stale := cache.StaleBranches(common, time.Minute, []string{"main", "feature/x"})
	if len(stale) != 1 || stale[0] != "feature/x" {
		t.Fatalf("want only feature/x stale, got %v", stale)
	}
}

type countingFetchExec struct {
	inner   *fakeExec
	fetches *atomic.Int32
}

func (c *countingFetchExec) LookPath(name string) (string, error) {
	return c.inner.LookPath(name)
}

func (c *countingFetchExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	if strings.Contains(key, "fetch origin") && !strings.Contains(key, "--prune") {
		c.fetches.Add(1)
	}
	return c.inner.Run(ctx, name, args...)
}

func (c *countingFetchExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	return c.Run(ctx, name, args...)
}

func hasCallSubstr(calls []string, substr string) bool {
	for _, c := range calls {
		if strings.Contains(c, substr) {
			return true
		}
	}
	return false
}

// A zero-value cache (no clock, nil maps) must behave like no cache rather
// than panic inside a dashboard collect goroutine.
func TestOriginFetchCacheZeroValueIsSafe(t *testing.T) {
	var c OriginFetchCache
	c.MarkFullSuccess("/repo/.git")
	c.MarkBranchesSuccess("/repo/.git", []string{"main"})
	if !c.NeedsFullFetch("/repo/.git", time.Minute, false) {
		t.Fatal("zero value must always need a full fetch")
	}
	if got := c.StaleBranches("/repo/.git", time.Minute, []string{"main"}); len(got) != 1 {
		t.Fatalf("zero value must report every branch stale: %v", got)
	}
}

// filepath.Clean("") is ".", so an empty common dir used to share one TTL
// slot across unrelated repositories.
func TestOriginFetchCacheEmptyCommonDirIsNotKeyed(t *testing.T) {
	c := NewOriginFetchCache(time.Now)
	c.MarkFullSuccess("")
	if !c.NeedsFullFetch(".", time.Minute, false) {
		t.Fatal(`an empty common dir must not be stored as "."`)
	}
}
