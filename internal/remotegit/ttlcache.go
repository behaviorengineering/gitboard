package remotegit

import (
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/board"
)

// TTLCache stores slow-changing forge slices in memory.
type TTLCache struct {
	mu    sync.Mutex
	now   func() time.Time
	byKey map[string]*cacheEntry
}

type cacheEntry struct {
	heads     HeadsSnapshot
	headsAt   time.Time
	hasHeads  bool
	merged    []board.MergedReview
	mergedOK  bool
	mergedAt  time.Time
	hasMerged bool
}

// NewTTLCache returns an empty upstream cache.
func NewTTLCache() *TTLCache {
	return &TTLCache{
		now:   time.Now,
		byKey: map[string]*cacheEntry{},
	}
}

// SetNow injects a clock for tests.
func (c *TTLCache) SetNow(now func() time.Time) {
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

// Clear drops all cached heads and merged slices.
// Used when tracked projects change (for example after gitboard sync).
func (c *TTLCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byKey = map[string]*cacheEntry{}
}

func (c *TTLCache) clock() time.Time {
	if c == nil || c.now == nil {
		return time.Now()
	}
	return c.now()
}

func cacheKey(host, path string) string {
	return host + "/" + path
}

func (c *TTLCache) entry(key string) *cacheEntry {
	if c.byKey == nil {
		c.byKey = map[string]*cacheEntry{}
	}
	e, ok := c.byKey[key]
	if !ok {
		e = &cacheEntry{}
		c.byKey[key] = e
	}
	return e
}

// GetOrLoadHeads returns cached heads or calls load.
func (c *TTLCache) GetOrLoadHeads(key string, ttl time.Duration, fresh bool, load func() (HeadsSnapshot, error)) (HeadsSnapshot, error) {
	if c == nil || ttl <= 0 || fresh {
		return load()
	}
	c.mu.Lock()
	e := c.entry(key)
	if e.hasHeads && c.clock().Sub(e.headsAt) < ttl {
		out := cloneHeads(e.heads)
		c.mu.Unlock()
		return out, nil
	}
	c.mu.Unlock()

	snap, err := load()
	if err != nil {
		return HeadsSnapshot{}, err
	}

	c.mu.Lock()
	e = c.entry(key)
	e.heads = cloneHeads(snap)
	e.headsAt = c.clock()
	e.hasHeads = true
	c.mu.Unlock()
	return snap, nil
}

// GetOrLoadMerged returns cached merged reviews or calls load.
// ok is false when load fails and there is no prior successful cache.
func (c *TTLCache) GetOrLoadMerged(key string, ttl time.Duration, fresh bool, load func() ([]board.MergedReview, error)) (merged []board.MergedReview, ok bool) {
	if c == nil || ttl <= 0 || fresh {
		list, err := load()
		if err != nil {
			return nil, false
		}
		return cloneMerged(list), true
	}
	c.mu.Lock()
	e := c.entry(key)
	if e.hasMerged && e.mergedOK && c.clock().Sub(e.mergedAt) < ttl {
		out := cloneMerged(e.merged)
		c.mu.Unlock()
		return out, true
	}
	c.mu.Unlock()

	list, err := load()
	if err != nil {
		// Fail closed for prune: expired cache must not keep MergedOK after a failed refresh.
		return nil, false
	}

	c.mu.Lock()
	e = c.entry(key)
	e.merged = cloneMerged(list)
	e.mergedAt = c.clock()
	e.hasMerged = true
	e.mergedOK = true
	c.mu.Unlock()
	return cloneMerged(list), true
}

func cloneHeads(in HeadsSnapshot) HeadsSnapshot {
	out := HeadsSnapshot{DefaultBranch: in.DefaultBranch}
	if len(in.Heads) > 0 {
		out.Heads = append([]RemoteHead(nil), in.Heads...)
	}
	return out
}

func cloneMerged(in []board.MergedReview) []board.MergedReview {
	if len(in) == 0 {
		return nil
	}
	return append([]board.MergedReview(nil), in...)
}
