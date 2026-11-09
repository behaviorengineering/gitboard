package dashboard

import (
	"context"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// confirmMergedForCandidates runs a per-branch merged lookup for prune candidates
// missing from the bulk merged list. Hits are appended to row.Merged. Confirmed
// misses are recorded on row.MergedChecked so EnrichPruneHints can set likely.
func (s *Service) confirmMergedForCandidates(ctx context.Context, p boardconfig.Project, row *board.ProjectSummary, opts remotegit.SummaryOpts) {
	if s == nil || row == nil || row.Local == nil || !row.Local.Mapped || !row.RemoteNamesOK {
		return
	}
	client := ClientFor(s, p)
	if client == nil {
		return
	}
	cache := opts.Cache
	if cache == nil {
		cache = s.cache
	}
	key := remotegit.CacheKey(p.Host, p.Path)
	// A branch back on the remote must not keep a stale per-branch answer.
	for _, name := range row.RemoteNames {
		cache.ForgetMergedBranch(key, name)
	}
	candidates := pruneLookupBranches(row)
	if len(candidates) == 0 {
		return
	}
	if row.MergedChecked == nil {
		row.MergedChecked = map[string]bool{}
	}
	known := mergedBranches(row.Merged)
	for _, branch := range candidates {
		if ctx.Err() != nil {
			return
		}
		if _, ok := known[branch]; ok || localgit.ValidateBranchName(branch) != nil {
			continue
		}
		list, err := cache.GetOrLoadMergedBranch(remotegit.MergedBranchQuery{Key: key, Branch: branch, NegativeTTL: opts.HeadsTTL, Fresh: opts.Fresh}, func() ([]board.MergedReview, error) {
			return client.MergedForBranch(ctx, p.Path, branch)
		})
		if err != nil {
			// Fail closed: the branch stays out of MergedChecked, so no
			// merged prune hint is drawn from a lookup that did not run.
			continue
		}
		row.MergedChecked[branch] = true
		row.Merged = append(row.Merged, list...)
	}
}

// mergedBranches is the set of branch names with a merged review.
func mergedBranches(merged []board.MergedReview) map[string]struct{} {
	out := map[string]struct{}{}
	for _, m := range merged {
		if n := strings.TrimSpace(m.Branch); n != "" {
			out[n] = struct{}{}
		}
	}
	return out
}

// pruneLookupBranches lists the distinct branches of prune candidates: the
// worktrees whose prune verdict passed every gate, so only merged or content
// evidence is left to decide.
func pruneLookupBranches(row *board.ProjectSummary) []string {
	if row == nil || row.Local == nil {
		return nil
	}
	sc := newPruneScope(row)
	seen := map[string]struct{}{}
	var out []string
	for _, wt := range row.Local.Worktrees {
		if !pruneVerdict(sc.inputs(wt)).Candidate {
			continue
		}
		branch := strings.TrimSpace(wt.Branch)
		if _, dup := seen[branch]; dup {
			continue
		}
		seen[branch] = struct{}{}
		out = append(out, branch)
	}
	return out
}
