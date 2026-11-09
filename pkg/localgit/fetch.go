package localgit

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/behaviorengineering/gitboard/pkg/clock"
)

// OriginFetchCache TTL-gates full origin fetches by common git dir and records
// which branches were refreshed by a targeted fetch while the full TTL is warm.
type OriginFetchCache struct {
	mu       sync.Mutex
	now      clock.Clock
	fullAt   map[string]time.Time
	branchAt map[string]map[string]time.Time
	group    singleflight.Group
}

// NewOriginFetchCache returns an empty fetch TTL cache that reads now for
// every freshness stamp. Panics if now is nil (required dependency at wire-up).
func NewOriginFetchCache(now clock.Clock) *OriginFetchCache {
	if now == nil {
		panic("localgit.NewOriginFetchCache: Clock is required")
	}
	return &OriginFetchCache{
		now:      now,
		fullAt:   map[string]time.Time{},
		branchAt: map[string]map[string]time.Time{},
	}
}

// Clear drops all fetch timestamps.
func (c *OriginFetchCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fullAt = map[string]time.Time{}
	c.branchAt = map[string]map[string]time.Time{}
}

// Invalidate drops full and branch fetch timestamps for one common git dir.
func (c *OriginFetchCache) Invalidate(commonDir string) {
	if c == nil {
		return
	}
	commonDir = normalizeCommonDir(commonDir)
	if commonDir == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.fullAt, commonDir)
	delete(c.branchAt, commonDir)
}

// clock reads the injected time source; callers hold c.mu and have checked usable.
func (c *OriginFetchCache) clock() time.Time {
	return c.now()
}

// usable reports whether c can stamp entries. A nil cache or a zero value
// built without NewOriginFetchCache (no clock) behaves like no cache: every
// fetch is needed and nothing is recorded.
func (c *OriginFetchCache) usable() bool {
	return c != nil && c.now != nil
}

// NeedsFullFetch reports whether a full `git fetch --prune origin` is required.
// ttl <= 0 or fresh always needs a full fetch. A nil cache always needs a full fetch.
func (c *OriginFetchCache) NeedsFullFetch(commonDir string, ttl time.Duration, fresh bool) bool {
	commonDir = normalizeCommonDir(commonDir)
	if commonDir == "" {
		return true
	}
	if !c.usable() || ttl <= 0 || fresh {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	at, ok := c.fullAt[commonDir]
	if !ok {
		return true
	}
	return c.clock().Sub(at) >= ttl
}

// StaleBranches returns branches that still need a targeted refresh while the full TTL is warm.
func (c *OriginFetchCache) StaleBranches(commonDir string, ttl time.Duration, branches []string) []string {
	commonDir = normalizeCommonDir(commonDir)
	cleaned := sanitizeBranchList(branches)
	if commonDir == "" || len(cleaned) == 0 {
		return nil
	}
	if !c.usable() || ttl <= 0 {
		return cleaned
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	byBranch := c.branchAt[commonDir]
	var stale []string
	for _, b := range cleaned {
		at, ok := byBranch[b]
		if !ok || now.Sub(at) >= ttl {
			stale = append(stale, b)
		}
	}
	return stale
}

// MarkFullSuccess records a successful full prune fetch for commonDir.
func (c *OriginFetchCache) MarkFullSuccess(commonDir string) {
	if !c.usable() {
		return
	}
	commonDir = normalizeCommonDir(commonDir)
	if commonDir == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fullAt == nil {
		c.fullAt = map[string]time.Time{}
	}
	if c.branchAt == nil {
		c.branchAt = map[string]map[string]time.Time{}
	}
	now := c.clock()
	c.fullAt[commonDir] = now
	// A full prune refreshes every remote head we care about next.
	c.branchAt[commonDir] = map[string]time.Time{}
}

// MarkBranchesSuccess records a successful targeted fetch for the named branches.
func (c *OriginFetchCache) MarkBranchesSuccess(commonDir string, branches []string) {
	if !c.usable() {
		return
	}
	commonDir = normalizeCommonDir(commonDir)
	cleaned := sanitizeBranchList(branches)
	if commonDir == "" || len(cleaned) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.branchAt == nil {
		c.branchAt = map[string]map[string]time.Time{}
	}
	byBranch := c.branchAt[commonDir]
	if byBranch == nil {
		byBranch = map[string]time.Time{}
		c.branchAt[commonDir] = byBranch
	}
	now := c.clock()
	for _, b := range cleaned {
		byBranch[b] = now
	}
}

// fetchOrigin runs git fetch --prune origin to refresh refs/remotes/origin/*.
// abs must already be bound to the lease.
func (m *Mutator) fetchOrigin(ctx context.Context, abs string) error {
	if _, err := m.in.git(ctx, abs, "fetch", "--prune", "origin"); err != nil {
		return fmt.Errorf("fetch --prune origin: %w", err)
	}
	return nil
}

// fetchOriginBranches refreshes only the named origin branches (no --prune).
// An empty or invalid branch list is a no-op (never falls back to full prune).
func (m *Mutator) fetchOriginBranches(ctx context.Context, abs string, branches []string) error {
	cleaned := sanitizeBranchList(branches)
	if len(cleaned) == 0 {
		return nil
	}
	args := append([]string{"fetch", "origin"}, cleaned...)
	if _, err := m.in.git(ctx, abs, args...); err != nil {
		return fmt.Errorf("fetch origin branches: %w", err)
	}
	return nil
}

// ListLocalHeads returns short names under refs/heads/ for repoPath.
func (in *Inspector) ListLocalHeads(ctx context.Context, repoPath string) ([]string, error) {
	if in == nil {
		return nil, ErrInspectorMissing
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return nil, fmt.Errorf("expand path: %w", err)
	}
	return in.listRefShortNames(ctx, abs, headsPrefix)
}

// fetchOriginCached fetches origin when the full-fetch TTL says common is stale.
// Concurrent misses for the same common dir coalesce to one full fetch.
// updated is true when this call performed (or shared) a real full fetch.
func (m *Mutator) fetchOriginCached(ctx context.Context, abs, common string, ttl time.Duration, fresh bool, cache *OriginFetchCache) (updated bool, err error) {
	if cache == nil {
		if err := m.fetchOrigin(ctx, abs); err != nil {
			return false, err
		}
		return true, nil
	}
	if !cache.NeedsFullFetch(common, ttl, fresh) {
		return false, nil
	}
	return doFlight(&cache.group, "full:"+common, func() (bool, error) {
		if !cache.NeedsFullFetch(common, ttl, fresh) {
			return false, nil
		}
		if err := m.fetchOrigin(ctx, abs); err != nil {
			return false, err
		}
		cache.MarkFullSuccess(common)
		return true, nil
	})
}

// doFlight runs fn through g with a statically typed result, so a shared
// result of the wrong type is an error rather than a silent zero value.
func doFlight[T any](g *singleflight.Group, key string, fn func() (T, error)) (T, error) {
	var zero T
	v, err, _ := g.Do(key, func() (any, error) { return fn() })
	if err != nil {
		return zero, err
	}
	out, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("singleflight %q: unexpected result %T", key, v)
	}
	return out, nil
}

// FetchPolicy says when FetchOriginSmart may skip the network.
type FetchPolicy struct {
	// TTL is how long a full fetch stays warm; zero or less always fetches.
	TTL time.Duration
	// Fresh forces a full fetch whatever the TTL says.
	Fresh bool
	// Branches are refreshed by a targeted fetch while the full TTL is warm.
	Branches []string
}

// FetchOriginSmart chooses full prune vs targeted branch fetch.
// Fresh or expired full TTL → git fetch --prune origin.
// Warm full TTL → targeted fetch of stale branches only; empty branch list skips network.
// updated is true when this call performed (or shared) a real network fetch.
// A nil cache always runs a full fetch. repoPath must resolve to the lease's
// common git directory.
func (m *Mutator) FetchOriginSmart(ctx context.Context, repoPath string, policy FetchPolicy, cache *OriginFetchCache) (updated bool, err error) {
	abs, common, err := m.bind(ctx, repoPath)
	if err != nil {
		return false, err
	}
	ttl, fresh, branches := policy.TTL, policy.Fresh, policy.Branches

	if cache == nil || cache.NeedsFullFetch(common, ttl, fresh) {
		updated, err := m.fetchOriginCached(ctx, abs, common, ttl, fresh, cache)
		if err != nil {
			return false, err
		}
		// Full prune refreshed remote heads; mark requested branches so this
		// call does not immediately re-fetch them via the targeted path.
		if cache != nil {
			cache.MarkBranchesSuccess(common, branches)
		}
		return updated, nil
	}

	stale := cache.StaleBranches(common, ttl, branches)
	if len(stale) == 0 {
		return false, nil
	}

	key := "branches:" + common + ":" + strings.Join(stale, ",")
	did, err := doFlight(&cache.group, key, func() (bool, error) {
		need := cache.StaleBranches(common, ttl, stale)
		if len(need) == 0 {
			return false, nil
		}
		if err := m.fetchOriginBranches(ctx, abs, need); err != nil {
			return false, err
		}
		cache.MarkBranchesSuccess(common, need)
		return true, nil
	})
	if err != nil {
		return false, err
	}

	// A concurrent singleflight may have covered a different subset; fetch any leftovers.
	leftover := cache.StaleBranches(common, ttl, branches)
	if len(leftover) == 0 {
		return did, nil
	}
	leftoverKey := "branches:" + common + ":" + strings.Join(leftover, ",")
	did2, err := doFlight(&cache.group, leftoverKey, func() (bool, error) {
		need := cache.StaleBranches(common, ttl, leftover)
		if len(need) == 0 {
			return false, nil
		}
		if err := m.fetchOriginBranches(ctx, abs, need); err != nil {
			return false, err
		}
		cache.MarkBranchesSuccess(common, need)
		return true, nil
	})
	if err != nil {
		return false, err
	}
	return did || did2, nil
}

func sanitizeBranchList(branches []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, b := range branches {
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		if err := ValidateBranchName(b); err != nil {
			continue
		}
		if _, ok := seen[b]; ok {
			continue
		}
		seen[b] = struct{}{}
		out = append(out, b)
	}
	sort.Strings(out)
	return out
}

// InvalidateOriginSync clears ahead/behind data when origin freshness is unknown
// (for example after a failed fetch). Callers still set Status.Error separately.
func InvalidateOriginSync(st *Status) {
	if st == nil {
		return
	}
	st.OriginSync = nil
	st.DefaultAhead = 0
	st.DefaultBehind = 0
}
