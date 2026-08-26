package localgit

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// OriginFetchCache TTL-gates git fetch origin by common git dir.
// Successful fetches are recorded; failures are not, so the next call retries.
type OriginFetchCache struct {
	mu  sync.Mutex
	now func() time.Time
	at  map[string]time.Time
}

// NewOriginFetchCache returns an empty fetch TTL cache.
func NewOriginFetchCache() *OriginFetchCache {
	return &OriginFetchCache{
		now: time.Now,
		at:  map[string]time.Time{},
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
	c.at = map[string]time.Time{}
}

func (c *OriginFetchCache) clock() time.Time {
	if c == nil || c.now == nil {
		return time.Now()
	}
	return c.now()
}

// NeedsFetch reports whether origin should be fetched for commonDir.
// ttl <= 0 or fresh always needs a fetch. A nil cache always needs a fetch.
func (c *OriginFetchCache) NeedsFetch(commonDir string, ttl time.Duration, fresh bool) bool {
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" {
		return true
	}
	if c == nil || ttl <= 0 || fresh {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	at, ok := c.at[commonDir]
	if !ok {
		return true
	}
	return c.clock().Sub(at) >= ttl
}

// MarkSuccess records a successful fetch for commonDir.
func (c *OriginFetchCache) MarkSuccess(commonDir string) {
	if c == nil {
		return
	}
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at == nil {
		c.at = map[string]time.Time{}
	}
	c.at[commonDir] = c.clock()
}

// FetchOrigin runs git fetch --prune origin to refresh refs/remotes/origin/*.
func (in *Inspector) FetchOrigin(ctx context.Context, repoPath string) error {
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	if _, err := in.git(ctx, abs, "fetch", "--prune", "origin"); err != nil {
		return fmt.Errorf("fetch --prune origin: %w", err)
	}
	return nil
}

// FetchOriginCached fetches origin when the TTL cache says the common git dir is stale.
// On success it records the fetch time. On failure it returns the error and does not
// update the cache (so the next call retries).
func (in *Inspector) FetchOriginCached(ctx context.Context, repoPath string, ttl time.Duration, fresh bool, cache *OriginFetchCache) error {
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	common, err := in.CommonGitDir(ctx, abs)
	if err != nil || common == "" {
		common = abs
	}
	common = filepath.Clean(common)
	if !cache.NeedsFetch(common, ttl, fresh) {
		return nil
	}
	if err := in.FetchOrigin(ctx, abs); err != nil {
		return err
	}
	cache.MarkSuccess(common)
	return nil
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
