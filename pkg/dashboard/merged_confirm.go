package dashboard

import (
	"context"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// confirmMergedForCandidates runs a per-branch merged lookup for prune candidates
// missing from the bulk merged list. Hits are appended to row.Merged. Confirmed
// misses are recorded on row.MergedChecked so EnrichPruneHints can set likely.
func (s *Service) confirmMergedForCandidates(ctx context.Context, p config.Project, row *board.ProjectSummary, opts remotegit.SummaryOpts) {
	if s == nil || row == nil || row.Local == nil || !row.Local.Mapped || !row.RemoteNamesOK {
		return
	}
	client := ClientFor(s, p)
	if client == nil {
		return
	}

	cache := opts.Cache
	if cache == nil {
		cache = s.Cache
	}
	key := remotegit.CacheKey(string(p.Host), p.Path)

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
	known := map[string]struct{}{}
	for _, m := range row.Merged {
		if n := strings.TrimSpace(m.Branch); n != "" {
			known[n] = struct{}{}
		}
	}

	for _, branch := range candidates {
		if ctx.Err() != nil {
			return
		}
		if _, ok := known[branch]; ok {
			continue
		}
		if err := localgit.ValidateBranchName(branch); err != nil {
			continue
		}
		list, ok := cache.GetOrLoadMergedBranch(key, branch, opts.HeadsTTL, opts.Fresh, func() ([]board.MergedReview, error) {
			return client.MergedForBranch(ctx, p.Path, branch)
		})
		if !ok {
			continue
		}
		row.MergedChecked[branch] = true
		if len(list) == 0 {
			continue
		}
		row.Merged = append(row.Merged, list...)
		known[branch] = struct{}{}
	}
}

func pruneLookupBranches(row *board.ProjectSummary) []string {
	if row == nil || row.Local == nil {
		return nil
	}
	remote := map[string]struct{}{}
	open := map[string]struct{}{}
	for _, name := range row.RemoteNames {
		name = strings.TrimSpace(name)
		if name != "" {
			remote[name] = struct{}{}
		}
	}
	for _, b := range row.Branches {
		name := strings.TrimSpace(b.Name)
		if name == "" {
			continue
		}
		remote[name] = struct{}{}
		if b.OpenReview {
			open[name] = struct{}{}
		}
	}
	defaultName := strings.TrimSpace(row.Local.DefaultBranch)
	for _, b := range row.Branches {
		if b.Default {
			if n := strings.TrimSpace(b.Name); n != "" {
				defaultName = n
			}
			break
		}
	}

	seen := map[string]struct{}{}
	var out []string
	for _, wt := range row.Local.Worktrees {
		branch := strings.TrimSpace(wt.Branch)
		if branch == "" || wt.Detached || wt.Bare || wt.Dirty {
			continue
		}
		if defaultName != "" && branch == defaultName {
			continue
		}
		if _, onRemote := remote[branch]; onRemote {
			continue
		}
		if _, stillOpen := open[branch]; stillOpen {
			continue
		}
		if _, dup := seen[branch]; dup {
			continue
		}
		seen[branch] = struct{}{}
		out = append(out, branch)
	}
	return out
}
