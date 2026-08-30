package remotegit

import (
	"context"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
)

// summaryLoader is the interface that GitHub and GitLab implement to share
// the ProjectSummary loading flow. Unexported methods are package-internal.
type summaryLoader interface {
	AuthStatus(ctx context.Context) (installed, authed bool, detail string)
	// unauthMsg formats the not-authenticated error message for this forge.
	unauthMsg(detail string) string
	// seedHeads loads the default branch and a bounded first page of branch heads.
	// This replaces the old paginated census (max one page, ~100 branches).
	seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error)
	// loadOpenReviews populates open PR/MR counts in summary and seeds PR branches
	// into branches so they appear even if outside the first page of remote heads.
	loadOpenReviews(ctx context.Context, repo string, summary *board.ProjectSummary, branches *branchAccum) error
	// loadCI populates the latest CI status in summary and attaches run info to branches.
	loadCI(ctx context.Context, repo string, summary *board.ProjectSummary, branches *branchAccum) error
	// loadMerged returns recently merged PRs/MRs for prune hint matching.
	loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error)
}

// projectSummaryShared is the shared ProjectSummary flow used by both GitHub and GitLab.
// Loading order:
//  1. auth check
//  2. seed heads (default branch + first page; no 50-page census)
//  3. loadOpenReviews (seeds PR/MR branches into branchAccum)
//  4. loadCI (attaches CI to seeded branches)
//  5. loadMerged (via cache)
func projectSummaryShared(ctx context.Context, p config.Project, opts SummaryOpts, loader summaryLoader) (board.ProjectSummary, error) {
	summary := baseSummary(p)

	// Step 1: auth.
	installed, authed, detail := loader.AuthStatus(ctx)
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
	if err := loader.loadOpenReviews(ctx, repo, &summary, branches); err != nil && summary.Error == "" {
		summary.Error = err.Error()
	}

	// Step 4: CI (attaches to branches seeded by heads or reviews).
	if err := loader.loadCI(ctx, repo, &summary, branches); err != nil && summary.Error == "" {
		summary.Error = err.Error()
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
