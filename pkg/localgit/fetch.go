package localgit

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// OriginFetchCache TTL-gates full origin fetches by common git dir and records
// which branches were refreshed by a targeted fetch while the full TTL is warm.
type OriginFetchCache struct {
	mu       sync.Mutex
	now      func() time.Time
	fullAt   map[string]time.Time
	branchAt map[string]map[string]time.Time
	group    singleflight.Group
}

// NewOriginFetchCache returns an empty fetch TTL cache.
func NewOriginFetchCache() *OriginFetchCache {
	return &OriginFetchCache{
		now:      time.Now,
		fullAt:   map[string]time.Time{},
		branchAt: map[string]map[string]time.Time{},
	}
}

// SetNow injects a clock for tests.
func (c *OriginFetchCache) SetNow(now func() time.Time) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if now == nil {
		c.now = time.Now
		return
	}
	c.now = now
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
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.fullAt, commonDir)
	delete(c.branchAt, commonDir)
}

func (c *OriginFetchCache) clock() time.Time {
	if c == nil || c.now == nil {
		return time.Now()
	}
	return c.now()
}

// NeedsFullFetch reports whether a full `git fetch --prune origin` is required.
// ttl <= 0 or fresh always needs a full fetch. A nil cache always needs a full fetch.
func (c *OriginFetchCache) NeedsFullFetch(commonDir string, ttl time.Duration, fresh bool) bool {
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" {
		return true
	}
	if c == nil || ttl <= 0 || fresh {
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

// NeedsFetch is an alias for NeedsFullFetch for callers that only care about full prune TTL.
func (c *OriginFetchCache) NeedsFetch(commonDir string, ttl time.Duration, fresh bool) bool {
	return c.NeedsFullFetch(commonDir, ttl, fresh)
}

// StaleBranches returns branches that still need a targeted refresh while the full TTL is warm.
func (c *OriginFetchCache) StaleBranches(commonDir string, ttl time.Duration, branches []string) []string {
	commonDir = filepath.Clean(commonDir)
	cleaned := sanitizeBranchList(branches)
	if commonDir == "" || len(cleaned) == 0 {
		return nil
	}
	if c == nil || ttl <= 0 {
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
	if c == nil {
		return
	}
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fullAt == nil {
		c.fullAt = map[string]time.Time{}
	}
	now := c.clock()
	c.fullAt[commonDir] = now
	// A full prune refreshes every remote head we care about next.
	c.branchAt[commonDir] = map[string]time.Time{}
}

// MarkSuccess is an alias for MarkFullSuccess.
func (c *OriginFetchCache) MarkSuccess(commonDir string) {
	c.MarkFullSuccess(commonDir)
}

// MarkBranchesSuccess records a successful targeted fetch for the named branches.
func (c *OriginFetchCache) MarkBranchesSuccess(commonDir string, branches []string) {
	if c == nil {
		return
	}
	commonDir = filepath.Clean(commonDir)
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

// FetchOrigin runs git fetch --prune origin to refresh refs/remotes/origin/*.
func (in *Inspector) FetchOrigin(ctx context.Context, repoPath string) error {
	if in == nil {
		return ErrInspectorMissing
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	if _, err := in.git(ctx, abs, "fetch", "--prune", "origin"); err != nil {
		return fmt.Errorf("fetch --prune origin: %w", err)
	}
	return nil
}

// FetchOriginBranches refreshes only the named origin branches (no --prune).
// An empty or invalid branch list is a no-op (never falls back to full prune).
func (in *Inspector) FetchOriginBranches(ctx context.Context, repoPath string, branches []string) error {
	if in == nil {
		return ErrInspectorMissing
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	cleaned := sanitizeBranchList(branches)
	if len(cleaned) == 0 {
		return nil
	}
	args := append([]string{"fetch", "origin"}, cleaned...)
	if _, err := in.git(ctx, abs, args...); err != nil {
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
	return in.listRefShortNames(ctx, abs, "refs/heads/")
}

// FetchOriginCached fetches origin when the full-fetch TTL says the common git dir is stale.
// Concurrent misses for the same common dir coalesce to one full fetch.
func (in *Inspector) FetchOriginCached(ctx context.Context, repoPath string, ttl time.Duration, fresh bool, cache *OriginFetchCache) error {
	if in == nil {
		return ErrInspectorMissing
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	common, err := in.CommonGitDir(ctx, abs)
	if err != nil || common == "" {
		common = abs
	}
	common = filepath.Clean(common)
	if !cache.NeedsFullFetch(common, ttl, fresh) {
		return nil
	}
	if cache == nil {
		return in.FetchOrigin(ctx, abs)
	}
	_, err, _ = cache.group.Do("full:"+common, func() (any, error) {
		if !cache.NeedsFullFetch(common, ttl, fresh) {
			return nil, nil
		}
		if err := in.FetchOrigin(ctx, abs); err != nil {
			return nil, err
		}
		cache.MarkFullSuccess(common)
		return nil, nil
	})
	return err
}

// FetchOriginSmart chooses full prune vs targeted branch fetch.
// Fresh or expired full TTL → git fetch --prune origin.
// Warm full TTL → targeted fetch of stale branches only; empty branch list skips network.
func (in *Inspector) FetchOriginSmart(ctx context.Context, repoPath string, ttl time.Duration, fresh bool, cache *OriginFetchCache, branches []string) error {
	if in == nil {
		return ErrInspectorMissing
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	common, err := in.CommonGitDir(ctx, abs)
	if err != nil || common == "" {
		common = abs
	}
	common = filepath.Clean(common)

	if cache.NeedsFullFetch(common, ttl, fresh) {
		if err := in.FetchOriginCached(ctx, abs, ttl, fresh, cache); err != nil {
			return err
		}
		// Full prune refreshed remote heads; mark requested branches so this
		// call does not immediately re-fetch them via the targeted path.
		if cache != nil {
			cache.MarkBranchesSuccess(common, branches)
		}
		return nil
	}

	stale := cache.StaleBranches(common, ttl, branches)
	if len(stale) == 0 {
		return nil
	}
	if cache == nil {
		return in.FetchOriginBranches(ctx, abs, stale)
	}

	key := "branches:" + common + ":" + strings.Join(stale, ",")
	_, err, _ = cache.group.Do(key, func() (any, error) {
		need := cache.StaleBranches(common, ttl, stale)
		if len(need) == 0 {
			return nil, nil
		}
		if err := in.FetchOriginBranches(ctx, abs, need); err != nil {
			return nil, err
		}
		cache.MarkBranchesSuccess(common, need)
		return nil, nil
	})
	if err != nil {
		return err
	}

	// A concurrent singleflight may have covered a different subset; fetch any leftovers.
	leftover := cache.StaleBranches(common, ttl, branches)
	if len(leftover) == 0 {
		return nil
	}
	leftoverKey := "branches:" + common + ":" + strings.Join(leftover, ",")
	_, err, _ = cache.group.Do(leftoverKey, func() (any, error) {
		need := cache.StaleBranches(common, ttl, leftover)
		if len(need) == 0 {
			return nil, nil
		}
		if err := in.FetchOriginBranches(ctx, abs, need); err != nil {
			return nil, err
		}
		cache.MarkBranchesSuccess(common, need)
		return nil, nil
	})
	return err
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
