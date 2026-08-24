package forge

import "strings"

// EnrichPruneHints marks local worktrees that are candidates for removal.
// Safe: remote head gone, clean tree, forge has a merged PR/MR for the branch.
// Likely: remote head gone, clean tree, merged lookup succeeded, no merged match.
// Main is the primary worktree (first git worktree list entry), not "never prune";
// only the default branch name is excluded.
func EnrichPruneHints(summary *ProjectSummary) {
	if summary == nil || summary.Local == nil {
		return
	}

	remote := make(map[string]struct{})
	open := make(map[string]struct{})
	defaultName := strings.TrimSpace(summary.Local.DefaultBranch)

	for _, name := range summary.RemoteNames {
		name = trimBranch(name)
		if name != "" {
			remote[name] = struct{}{}
		}
	}
	for _, b := range summary.Branches {
		name := trimBranch(b.Name)
		if name == "" {
			continue
		}
		if len(summary.RemoteNames) == 0 {
			remote[name] = struct{}{}
		}
		if b.OpenReview {
			open[name] = struct{}{}
		}
		if b.Default && defaultName == "" {
			defaultName = name
		}
	}

	mergedByBranch := newestMergedByBranch(summary.Merged)
	for i := range summary.Local.Worktrees {
		applyPruneHint(&summary.Local.Worktrees[i], defaultName, remote, open, mergedByBranch, summary.MergedOK)
	}
}

func newestMergedByBranch(merged []MergedReview) map[string]MergedReview {
	out := make(map[string]MergedReview, len(merged))
	for _, m := range merged {
		name := trimBranch(m.Branch)
		if name == "" {
			continue
		}
		prev, ok := out[name]
		if !ok {
			out[name] = m
			continue
		}
		tNew, okNew := parseTime(m.MergedAt)
		tOld, okOld := parseTime(prev.MergedAt)
		if okNew && (!okOld || tNew.After(tOld)) {
			out[name] = m
		}
	}
	return out
}

func applyPruneHint(
	wt *LocalWorktree,
	defaultName string,
	remote, open map[string]struct{},
	mergedByBranch map[string]MergedReview,
	mergedOK bool,
) {
	if wt == nil {
		return
	}
	branch := trimBranch(wt.Branch)
	if branch == "" || wt.Detached || wt.Bare {
		return
	}
	if defaultName != "" && branch == defaultName {
		return
	}
	if _, onRemote := remote[branch]; onRemote {
		return
	}
	if wt.Dirty {
		return
	}
	if m, ok := mergedByBranch[branch]; ok {
		wt.PruneHint = PruneSafe
		wt.MergedID = m.ID
		wt.MergedURL = m.URL
		wt.MergedAt = m.MergedAt
		return
	}
	if !mergedOK {
		return
	}
	if _, stillOpen := open[branch]; stillOpen {
		return
	}
	wt.PruneHint = PruneLikely
}
