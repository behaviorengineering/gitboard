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
	heads          HeadsSnapshot
	headsAt        time.Time
	hasHeads       bool
	merged         []board.MergedReview
	mergedOK       bool
	mergedAt       time.Time
	hasMerged      bool
	mergedByBranch map[string]mergedBranchEntry
}

// mergedBranchEntry is a per-branch merged lookup result.
// Positives do not expire; negatives expire with the heads TTL.
type mergedBranchEntry struct {
	reviews []board.MergedReview
	miss    bool
	at      time.Time
	has     bool
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

// CacheKey is the TTL map key for one forge project.
func CacheKey(host, path string) string {
	return cacheKey(host, path)
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
// Successful live loads (including Fresh and TTL 0) write the cache so the next
// poll cannot disagree with prune or ?fresh=1.
func (c *TTLCache) GetOrLoadHeads(key string, ttl time.Duration, fresh bool, load func() (HeadsSnapshot, error)) (HeadsSnapshot, error) {
	if c != nil && ttl > 0 && !fresh {
		c.mu.Lock()
		e := c.entry(key)
		if e.hasHeads && c.clock().Sub(e.headsAt) < ttl {
			out := cloneHeads(e.heads)
			c.mu.Unlock()
			return out, nil
		}
		c.mu.Unlock()
	}

	snap, err := load()
	if err != nil {
		return HeadsSnapshot{}, err
	}
	if c != nil {
		c.mu.Lock()
		e := c.entry(key)
		e.heads = cloneHeads(snap)
		e.headsAt = c.clock()
		e.hasHeads = true
		c.mu.Unlock()
	}
	return snap, nil
}

// GetOrLoadMerged returns cached merged reviews or calls load.
// ok is false when load fails (expired cache must not keep MergedOK).
// Successful live loads write the cache, including Fresh and TTL 0.
func (c *TTLCache) GetOrLoadMerged(key string, ttl time.Duration, fresh bool, load func() ([]board.MergedReview, error)) (merged []board.MergedReview, ok bool) {
	if c != nil && ttl > 0 && !fresh {
		c.mu.Lock()
		e := c.entry(key)
		if e.hasMerged && e.mergedOK && c.clock().Sub(e.mergedAt) < ttl {
			out := cloneMerged(e.merged)
			c.mu.Unlock()
			return out, true
		}
		c.mu.Unlock()
	}

	list, err := load()
	if err != nil {
		// Fail closed for prune: expired cache must not keep MergedOK after a failed refresh.
		return nil, false
	}
	if c != nil {
		c.mu.Lock()
		e := c.entry(key)
		e.merged = cloneMerged(list)
		e.mergedAt = c.clock()
		e.hasMerged = true
		e.mergedOK = true
		c.mu.Unlock()
	}
	return cloneMerged(list), true
}

// GetOrLoadMergedBranch returns a targeted merged lookup for one source branch.
// Positives stick until ForgetMergedBranch. Negatives expire with negativeTTL.
// ok is false on lookup error with no sticky positive (do not treat as a miss).
func (c *TTLCache) GetOrLoadMergedBranch(key, branch string, negativeTTL time.Duration, fresh bool, load func() ([]board.MergedReview, error)) (merged []board.MergedReview, ok bool) {
	branch = trimBranch(branch)
	if branch == "" || load == nil {
		return nil, false
	}

	if c != nil && !fresh {
		c.mu.Lock()
		e := c.entry(key)
		if got, hit := e.branchHit(branch, negativeTTL, c.clock()); hit {
			out := cloneMerged(got)
			c.mu.Unlock()
			return out, true
		}
		c.mu.Unlock()
	}

	list, err := load()
	if err != nil {
		if c == nil {
			return nil, false
		}
		c.mu.Lock()
		e := c.entry(key)
		prev, okPrev := e.mergedByBranch[branch]
		c.mu.Unlock()
		if okPrev && prev.has && !prev.miss {
			return cloneMerged(prev.reviews), true
		}
		return nil, false
	}

	if c != nil {
		c.mu.Lock()
		e := c.entry(key)
		e.storeBranch(branch, list, c.clock())
		c.mu.Unlock()
	}
	return cloneMerged(list), true
}

// ForgetMergedBranch drops a sticky per-branch merged result (remote head returned).
func (c *TTLCache) ForgetMergedBranch(key, branch string) {
	branch = trimBranch(branch)
	if c == nil || branch == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.byKey[key]
	if !ok || e == nil || e.mergedByBranch == nil {
		return
	}
	delete(e.mergedByBranch, branch)
}

func (e *cacheEntry) branchHit(branch string, negativeTTL time.Duration, now time.Time) ([]board.MergedReview, bool) {
	if e == nil || e.mergedByBranch == nil {
		return nil, false
	}
	got, ok := e.mergedByBranch[branch]
	if !ok || !got.has {
		return nil, false
	}
	if !got.miss {
		return got.reviews, true
	}
	if negativeTTL <= 0 {
		return nil, false
	}
	if now.Sub(got.at) >= negativeTTL {
		return nil, false
	}
	return nil, true
}

func (e *cacheEntry) storeBranch(branch string, list []board.MergedReview, now time.Time) {
	if e == nil {
		return
	}
	if e.mergedByBranch == nil {
		e.mergedByBranch = map[string]mergedBranchEntry{}
	}
	if len(list) == 0 {
		e.mergedByBranch[branch] = mergedBranchEntry{miss: true, at: now, has: true}
		return
	}
	e.mergedByBranch[branch] = mergedBranchEntry{
		reviews: cloneMerged(list),
		at:      now,
		has:     true,
	}
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
