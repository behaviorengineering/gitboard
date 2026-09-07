package remotegit

import (
	"strings"

	"github.com/behaviorengineering/gitboard/internal/board"
)

// EnrichPruneHints marks local worktrees that are candidates for removal.
// Safe: remote head gone, clean tree, no open PR/MR, and either forge has a
// merged PR/MR for the branch or LocalWorktree.ContentOnDefault is true
// (caller annotates via localgit).
// Likely: remote head gone, clean tree, no open PR/MR, merged lookup succeeded,
// no merged/content match.
// Main is the primary worktree (first git worktree list entry), not "never prune";
// only the default branch name is excluded.
//
// The excluded default name prefers the forge-marked default over local origin/HEAD,
// which can still point at a feature branch after a non-default clone. When the forge
// default is known, Local.DefaultBranch is aligned to it for switch/prune callers.
//
// Fail closed: when RemoteNamesOK is false (heads unknown or incomplete), no hints
// are set. A branch still listed in Branches (UI remote rows) or with OpenReview
// never gets a prune hint, even if ContentOnDefault or a merged record would
// otherwise say safe. Remote membership for "still exists" uses RemoteNames plus
// Branches names; it never uses Branches alone as the full remote set.
func EnrichPruneHints(summary *board.ProjectSummary) {
	if summary == nil || summary.Local == nil {
		return
	}
	if !summary.RemoteNamesOK {
		return
	}

	remote := make(map[string]struct{})
	open := make(map[string]struct{})
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
		// Fail closed: a branch still shown as a remote row must not be prune-hinted,
		// even if RemoteNames briefly lagged behind the Branches payload.
		remote[name] = struct{}{}
		if b.OpenReview {
			open[name] = struct{}{}
		}
	}

	defaultName := resolveDefaultBranch(summary)
	if defaultName != "" {
		summary.Local.DefaultBranch = defaultName
	}

	mergedByBranch := newestMergedByBranch(summary.Merged)
	for i := range summary.Local.Worktrees {
		applyPruneHint(&summary.Local.Worktrees[i], defaultName, remote, open, mergedByBranch, summary.MergedOK)
	}
}

// resolveDefaultBranch prefers the forge default branch over local origin/HEAD.
func resolveDefaultBranch(summary *board.ProjectSummary) string {
	if summary == nil {
		return ""
	}
	for _, b := range summary.Branches {
		if !b.Default {
			continue
		}
		if name := trimBranch(b.Name); name != "" {
			return name
		}
	}
	if summary.Local == nil {
		return ""
	}
	return strings.TrimSpace(summary.Local.DefaultBranch)
}

func newestMergedByBranch(merged []board.MergedReview) map[string]board.MergedReview {
	out := make(map[string]board.MergedReview, len(merged))
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
	wt *board.LocalWorktree,
	defaultName string,
	remote, open map[string]struct{},
	mergedByBranch map[string]board.MergedReview,
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
	// Open review blocks every prune hint (including content-on-default / merged).
	if _, stillOpen := open[branch]; stillOpen {
		return
	}
	if m, ok := mergedByBranch[branch]; ok {
		wt.PruneHint = board.PruneSafe
		wt.MergedID = m.ID
		wt.MergedURL = m.URL
		wt.MergedAt = m.MergedAt
		return
	}
	if wt.ContentOnDefault {
		wt.PruneHint = board.PruneSafe
		return
	}
	if !mergedOK {
		return
	}
	wt.PruneHint = board.PruneLikely
}
