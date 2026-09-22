package remotegit

import (
	"context"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
)

// DefaultOpenTTL caches open reviews when SummaryOpts.OpenTTL is zero and Cache is set.
const DefaultOpenTTL = 30 * time.Second

// DefaultCITTL caches CI when SummaryOpts.CITTL is zero and Cache is set.
const DefaultCITTL = 30 * time.Second

// SummaryOpts controls which forge slices use the TTL cache.
type SummaryOpts struct {
	Fresh     bool
	Cache     *TTLCache
	Auth      *AuthCache    // optional request-scoped auth reuse
	HeadsTTL  time.Duration // 0 = always refetch heads
	MergedTTL time.Duration // 0 = always refetch merged
	OpenTTL   time.Duration // 0 = DefaultOpenTTL when Cache set; negative = always refetch
	CITTL     time.Duration // 0 = DefaultCITTL when Cache set; negative = always refetch
}

// AuthCache reuses AuthStatus results within one dashboard collect (or similar request).
type AuthCache struct {
	mu   sync.Mutex
	byID map[string]authResult
}

type authResult struct {
	installed bool
	authed    bool
	detail    string
}

// NewAuthCache returns an empty request-scoped auth cache.
func NewAuthCache() *AuthCache {
	return &AuthCache{byID: map[string]authResult{}}
}

// GetOrCheck returns a cached auth result or calls check once per id.
func (a *AuthCache) GetOrCheck(id string, check func() (installed, authed bool, detail string)) (installed, authed bool, detail string) {
	if a == nil {
		return check()
	}
	a.mu.Lock()
	if got, ok := a.byID[id]; ok {
		a.mu.Unlock()
		return got.installed, got.authed, got.detail
	}
	a.mu.Unlock()

	installed, authed, detail = check()
	a.mu.Lock()
	if a.byID == nil {
		a.byID = map[string]authResult{}
	}
	if got, ok := a.byID[id]; ok {
		a.mu.Unlock()
		return got.installed, got.authed, got.detail
	}
	a.byID[id] = authResult{installed: installed, authed: authed, detail: detail}
	a.mu.Unlock()
	return installed, authed, detail
}

// summaryLoader is the interface that forge adapters implement to share
// the ProjectSummary loading flow. Unexported methods are package-internal.
type summaryLoader interface {
	AuthStatus(ctx context.Context) (installed, authed bool, detail string)
	// authCacheID is a stable key for request-scoped AuthCache (for example "github").
	authCacheID() string
	// unauthMsg formats the not-authenticated error message for this forge.
	unauthMsg(detail string) string
	// seedHeads loads the default branch and a bounded first page of branch heads.
	seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error)
	// loadOpenReviews populates open PR/MR counts and review rows for caching.
	loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error)
	// loadCI populates the latest CI status and per-branch runs for caching.
	loadCI(ctx context.Context, repo string) (CISnapshot, error)
	// loadMerged returns a bounded recent merged PR/MR list (warm index for prune).
	loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error)
}

// projectSummaryShared is the shared ProjectSummary flow used by forge adapters.
// Loading order:
//  1. auth check (request-scoped reuse when Auth is set)
//  2. seed heads (default branch + first page; no 50-page census)
//  3. loadOpenReviews (short TTL)
//  4. loadCI (short TTL)
//  5. loadMerged (via cache)
func projectSummaryShared(ctx context.Context, p config.Project, opts SummaryOpts, loader summaryLoader) (board.ProjectSummary, error) {
	summary := baseSummary(p)

	// Step 1: auth.
	installed, authed, detail := opts.Auth.GetOrCheck(loader.authCacheID(), func() (bool, bool, string) {
		return loader.AuthStatus(ctx)
	})
	if !installed {
		summary.Error = detail
		return summary, nil
	}
	if !authed {
		summary.Error = loader.unauthMsg(detail)
		return summary, nil
	}

	repo := p.Path
	key := cacheKey(string(p.Host), repo)
	branches := newBranchAccum()

	openTTL := opts.OpenTTL
	if openTTL == 0 {
		openTTL = DefaultOpenTTL
	} else if openTTL < 0 {
		openTTL = 0
	}
	ciTTL := opts.CITTL
	if ciTTL == 0 {
		ciTTL = DefaultCITTL
	} else if ciTTL < 0 {
		ciTTL = 0
	}

	// Step 2: seed heads (default branch + first page only).
	heads, headsErr := opts.Cache.GetOrLoadHeads(key, opts.HeadsTTL, opts.Fresh, func() (HeadsSnapshot, error) {
		return loader.seedHeads(ctx, repo)
	})
	if headsErr != nil {
		if summary.Error == "" {
			summary.Error = headsErr.Error()
		}
		summary.RemoteNamesOK = false
	} else {
		summary.RemoteNamesOK = true
		if heads.DefaultBranch != "" {
			branches.setDefault(heads.DefaultBranch)
		}
		names := make([]string, 0, len(heads.Heads))
		for _, h := range heads.Heads {
			branches.addRemote(h.Name, h.UpdatedAt, h.WebURL)
			if n := trimBranch(h.Name); n != "" {
				names = append(names, n)
			}
		}
		summary.RemoteNames = names
	}

	// Step 3: open reviews (seeds PR/MR branches into branchAccum).
	open, openErr := opts.Cache.GetOrLoadOpen(key, openTTL, opts.Fresh, func() (OpenReviewsSnapshot, error) {
		return loader.loadOpenReviews(ctx, repo)
	})
	if openErr != nil && summary.Error == "" {
		summary.Error = openErr.Error()
	} else if openErr == nil {
		applyOpenReviews(&summary, branches, open)
	}

	// Step 4: CI (attaches to branches seeded by heads or reviews).
	ci, ciErr := opts.Cache.GetOrLoadCI(key, ciTTL, opts.Fresh, func() (CISnapshot, error) {
		return loader.loadCI(ctx, repo)
	})
	if ciErr != nil && summary.Error == "" {
		summary.Error = ciErr.Error()
	} else if ciErr == nil {
		applyCI(&summary, branches, ci)
	}

	// Step 5: merged reviews (for prune hints).
	merged, mergedOK := opts.Cache.GetOrLoadMerged(key, opts.MergedTTL, opts.Fresh, func() ([]board.MergedReview, error) {
		return loader.loadMerged(ctx, repo)
	})
	summary.Merged = merged
	summary.MergedOK = mergedOK
	summary.Branches = branches.list()
	return summary, nil
}

func applyOpenReviews(summary *board.ProjectSummary, branches *branchAccum, open OpenReviewsSnapshot) {
	if summary == nil {
		return
	}
	summary.OpenItems.PullRequests = open.PullRequests
	summary.OpenItems.MergeRequests = open.MergeRequests
	for _, r := range open.Reviews {
		branches.setOpenReview(r.Branch, reviewInfo{
			ID:        r.ID,
			URL:       r.URL,
			Conflict:  r.Conflict,
			UpdatedAt: r.UpdatedAt,
			Draft:     r.Draft,
		})
	}
}

func applyCI(summary *board.ProjectSummary, branches *branchAccum, ci CISnapshot) {
	if summary == nil {
		return
	}
	if ci.Latest != nil {
		cp := *ci.Latest
		summary.CI = &cp
	}
	for _, r := range ci.Runs {
		branches.setCI(r.Branch, r.Status, r.URL, r.UpdatedAt, r.RunID)
	}
}
