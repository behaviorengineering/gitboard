package remotegit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/clock"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// TTLCache stores slow-changing forge slices in memory.
type TTLCache struct {
	mu    sync.Mutex
	now   clock.Clock
	byKey map[string]*cacheEntry
	group singleflight.Group
}

type cacheEntry struct {
	heads          slot[HeadsSnapshot]
	merged         slot[[]board.MergedReview]
	open           slot[OpenReviewsSnapshot]
	ci             slot[CISnapshot]
	mergedByBranch map[string]mergedBranchEntry
}

// slot is one cached forge slice and when it was stored.
type slot[T any] struct {
	val T
	at  time.Time
	has bool
}

// slotKind describes one cached slice: its singleflight key prefix, where it
// lives in a cacheEntry, and how to copy it so callers never share storage.
type slotKind[T any] struct {
	name  string
	pick  func(*cacheEntry) *slot[T]
	clone func(T) T
}

var (
	headsKind  = slotKind[HeadsSnapshot]{name: "heads", pick: func(e *cacheEntry) *slot[HeadsSnapshot] { return &e.heads }, clone: cloneHeads}
	mergedKind = slotKind[[]board.MergedReview]{name: "merged", pick: func(e *cacheEntry) *slot[[]board.MergedReview] { return &e.merged }, clone: cloneMerged}
	openKind   = slotKind[OpenReviewsSnapshot]{name: "open", pick: func(e *cacheEntry) *slot[OpenReviewsSnapshot] { return &e.open }, clone: cloneOpen}
	ciKind     = slotKind[CISnapshot]{name: "ci", pick: func(e *cacheEntry) *slot[CISnapshot] { return &e.ci }, clone: cloneCI}
)

// OpenReviewsSnapshot is a cacheable open PR/MR slice.
type OpenReviewsSnapshot struct {
	PullRequests  int
	MergeRequests int
	Reviews       []CachedOpenReview
}

// CachedOpenReview is one open review used to re-seed branchAccum.
type CachedOpenReview struct {
	Branch    string
	ID        int
	URL       string
	Conflict  bool
	UpdatedAt string
	Draft     bool
}

// CISnapshot is a cacheable CI slice for one project.
type CISnapshot struct {
	Latest *board.CIStatus
	Runs   []CachedCIRun
}

// CachedCIRun attaches CI to a branch name.
type CachedCIRun struct {
	Branch    string
	Status    string
	URL       string
	UpdatedAt string
	RunID     string
}

// mergedBranchEntry is a per-branch merged lookup result.
// Positives do not expire; negatives expire with the heads TTL.
type mergedBranchEntry struct {
	reviews []board.MergedReview
	miss    bool
	at      time.Time
	has     bool
}

// NewTTLCache returns an empty upstream cache that reads now for every TTL
// check. Panics if now is nil (required dependency at wire-up).
func NewTTLCache(now clock.Clock) *TTLCache {
	if now == nil {
		panic("remotegit.NewTTLCache: Clock is required")
	}
	return &TTLCache{
		now:   now,
		byKey: map[string]*cacheEntry{},
	}
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

// ForgetProject drops cached slices for one forge project key.
func (c *TTLCache) ForgetProject(key string) {
	if c == nil || key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.byKey, key)
}

// clock reads the injected time source; callers have checked c for nil.
func (c *TTLCache) clock() time.Time {
	return c.now()
}

// CacheKey is the TTL map key for one forge project. The path is trimmed of
// spaces and slashes, so the loader and InvalidateProject derive one key.
func CacheKey(host boardconfig.Host, path string) string {
	return string(host) + "/" + strings.Trim(strings.TrimSpace(path), "/")
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
// Concurrent misses for the same key coalesce to one load.
// Successful live loads (including Fresh and TTL 0) write the cache so the next
// poll cannot disagree with prune or ?fresh=1.
func (c *TTLCache) GetOrLoadHeads(key string, ttl time.Duration, fresh bool, load func() (HeadsSnapshot, error)) (HeadsSnapshot, error) {
	return getOrLoad(c, headsKind, key, ttl, fresh, load)
}

// GetOrLoadMerged returns cached merged reviews or calls load.
// A failed load returns its error and is never cached, so an expired entry
// cannot keep MergedOK. Successful live loads write the cache, including
// Fresh and TTL 0.
func (c *TTLCache) GetOrLoadMerged(key string, ttl time.Duration, fresh bool, load func() ([]board.MergedReview, error)) ([]board.MergedReview, error) {
	return getOrLoad(c, mergedKind, key, ttl, fresh, load)
}

// GetOrLoadOpen returns cached open reviews or calls load.
// A failed load returns its error and is never cached as an empty snapshot.
func (c *TTLCache) GetOrLoadOpen(key string, ttl time.Duration, fresh bool, load func() (OpenReviewsSnapshot, error)) (OpenReviewsSnapshot, error) {
	return getOrLoad(c, openKind, key, ttl, fresh, load)
}

// GetOrLoadCI returns cached CI or calls load.
// A failed load returns its error and is never cached as an empty snapshot.
func (c *TTLCache) GetOrLoadCI(key string, ttl time.Duration, fresh bool, load func() (CISnapshot, error)) (CISnapshot, error) {
	return getOrLoad(c, ciKind, key, ttl, fresh, load)
}

// getOrLoad is the shared TTL read for every cached slice. A hit younger
// than ttl is returned unless fresh is set (ttl <= 0 never hits). Misses for
// the same key coalesce to one load; a successful load is stored even when
// fresh or ttl skipped the read, and a failed load is returned and never
// stored. A nil cache calls load directly. Every caller gets its own copy.
func getOrLoad[T any](c *TTLCache, kind slotKind[T], key string, ttl time.Duration, fresh bool, load func() (T, error)) (T, error) {
	var zero T
	if load == nil {
		return zero, ErrNilLoader
	}
	if c == nil {
		v, err := load()
		if err != nil {
			return zero, err
		}
		return kind.clone(v), nil
	}
	useCache := ttl > 0 && !fresh
	if useCache {
		if v, ok := peekSlot(c, kind, key, ttl); ok {
			return v, nil
		}
	}
	v, err := doFlight(&c.group, kind.name+":"+key, func() (T, error) {
		if useCache {
			if v, ok := peekSlot(c, kind, key, ttl); ok {
				return v, nil
			}
		}
		v, err := load()
		if err != nil {
			return zero, err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		s := kind.pick(c.entry(key))
		s.val, s.at, s.has = kind.clone(v), c.clock(), true
		return v, nil
	})
	if err != nil {
		return zero, err
	}
	return kind.clone(v), nil
}

// peekSlot returns a copy of the cached slice when it is younger than ttl.
func peekSlot[T any](c *TTLCache, kind slotKind[T], key string, ttl time.Duration) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := kind.pick(c.entry(key))
	if !s.has || c.clock().Sub(s.at) >= ttl {
		var zero T
		return zero, false
	}
	return kind.clone(s.val), true
}

// MergedBranchQuery names one per-branch merged lookup.
type MergedBranchQuery struct {
	// Key is the project cache key (CacheKey).
	Key string
	// Branch is the source branch name.
	Branch string
	// NegativeTTL is how long a "no merged review" answer is reused.
	NegativeTTL time.Duration
	// Fresh skips the cached answer and loads again.
	Fresh bool
}

// GetOrLoadMergedBranch returns a targeted merged lookup for one source branch.
// Positives stick until ForgetMergedBranch. Negatives expire with q.NegativeTTL.
// A lookup error with no sticky positive is returned (do not treat it as a miss).
func (c *TTLCache) GetOrLoadMergedBranch(q MergedBranchQuery, load func() ([]board.MergedReview, error)) ([]board.MergedReview, error) {
	if load == nil {
		return nil, ErrNilLoader
	}
	key, negativeTTL, fresh := q.Key, q.NegativeTTL, q.Fresh
	branch := trimBranch(q.Branch)
	if branch == "" {
		return nil, gberr.New(gberr.CodeInvalid, "remotegit.TTLCache.GetOrLoadMergedBranch", "branch is required")
	}
	if c == nil {
		list, err := load()
		if err != nil {
			return nil, err
		}
		return cloneMerged(list), nil
	}
	if !fresh {
		if got, hit := c.peekBranch(key, branch, negativeTTL); hit {
			return got, nil
		}
	}
	out, err := doFlight(&c.group, "mergedbranch:"+key+":"+branch, func() ([]board.MergedReview, error) {
		if !fresh {
			if got, hit := c.peekBranch(key, branch, negativeTTL); hit {
				return got, nil
			}
		}
		return c.loadBranch(key, branch, load)
	})
	if err != nil {
		return nil, err
	}
	return cloneMerged(out), nil
}

// peekBranch returns a copy of a sticky positive or an unexpired negative.
func (c *TTLCache) peekBranch(key, branch string, negativeTTL time.Duration) ([]board.MergedReview, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	got, hit := c.entry(key).branchHit(branch, negativeTTL, c.clock())
	if !hit {
		return nil, false
	}
	return cloneMerged(got), true
}

// loadBranch runs load and stores the result. A failed load falls back to a
// sticky positive when there is one, and is otherwise returned.
func (c *TTLCache) loadBranch(key, branch string, load func() ([]board.MergedReview, error)) ([]board.MergedReview, error) {
	list, err := load()
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.entry(key)
	if err != nil {
		if prev, ok := e.mergedByBranch[branch]; ok && prev.has && !prev.miss {
			return cloneMerged(prev.reviews), nil
		}
		return nil, err
	}
	e.storeBranch(branch, list, c.clock())
	return cloneMerged(list), nil
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

func cloneOpen(in OpenReviewsSnapshot) OpenReviewsSnapshot {
	out := OpenReviewsSnapshot{
		PullRequests:  in.PullRequests,
		MergeRequests: in.MergeRequests,
	}
	if len(in.Reviews) > 0 {
		out.Reviews = append([]CachedOpenReview(nil), in.Reviews...)
	}
	return out
}

func cloneCI(in CISnapshot) CISnapshot {
	out := CISnapshot{}
	if in.Latest != nil {
		cp := *in.Latest
		out.Latest = &cp
	}
	if len(in.Runs) > 0 {
		out.Runs = append([]CachedCIRun(nil), in.Runs...)
	}
	return out
}
