package dashboard

import (
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
)

// Reasons a branch may not be deleted. They are client-safe.
const (
	reasonNoBranch       = "worktree is not on a branch"
	reasonDefaultUnknown = "default branch is unknown"
	reasonDefault        = "branch is the default branch"
	reasonRemoteUnknown  = "remote branches did not load"
	reasonOnRemote       = "branch still exists on the remote"
	reasonDirty          = "working tree is dirty"
	reasonOpenUnknown    = "open reviews did not load"
	reasonOpenReview     = "branch has an open review"
	reasonMergedUnknown  = "merged review lookup did not run"
	reasonUnproven       = "no merged review and content is not on the default branch"
)

// PruneInputs is everything one prune decision reads, as plain data.
type PruneInputs struct {
	// Branch is the worktree's branch; NotBranch is true when the worktree
	// is detached or bare.
	Branch    string
	NotBranch bool
	Dirty     bool
	// DefaultBranch is the forge default, else the local origin/HEAD target.
	DefaultBranch string
	// RemoteNamesOK is false when remote heads did not load; OnRemote is
	// true when Branch still has a remote head.
	RemoteNamesOK bool
	OnRemote      bool
	// OpenOK is false when open reviews did not load; OpenReview is true
	// when Branch has one.
	OpenOK     bool
	OpenReview bool
	// Merged is true when a merged review exists for Branch. MergedOK is
	// true when the bulk merged list loaded; MergedChecked when a
	// per-branch lookup ran and succeeded.
	Merged        bool
	MergedOK      bool
	MergedChecked bool
	// ContentOnDefault is true when the branch tip adds nothing to default.
	ContentOnDefault bool
}

// PruneVerdict says whether one branch may be deleted, and why not.
type PruneVerdict struct {
	// Allow is true only for a safe prune: the board's "safe" hint and the
	// only verdict PruneSafe acts on.
	Allow bool
	// Hint is the board hint: safe, likely, or empty.
	Hint board.PruneHint
	// Candidate is true when every gate passed and only merged or content
	// evidence decides; callers look that evidence up for candidates only.
	Candidate bool
	// Reason says why Allow is false; empty when it is true.
	Reason string
}

// pruneVerdict holds every prune rule. enrichPruneHints, pruneLookupBranches,
// annotateContentOnDefault and PruneSafe all call it, so the board and the
// delete agree. Every unknown input refuses (fail closed).
func pruneVerdict(in PruneInputs) PruneVerdict {
	if reason := pruneGate(in); reason != "" {
		return PruneVerdict{Reason: reason}
	}
	switch {
	case in.Merged, in.ContentOnDefault:
		return PruneVerdict{Allow: true, Hint: board.PruneSafe, Candidate: true}
	case !in.MergedOK || !in.MergedChecked:
		return PruneVerdict{Candidate: true, Reason: reasonMergedUnknown}
	default:
		return PruneVerdict{Hint: board.PruneLikely, Candidate: true, Reason: reasonUnproven}
	}
}

// pruneGate returns the first rule that refuses before any merged or content
// evidence is considered, or "" when every gate passes.
func pruneGate(in PruneInputs) string {
	branch := strings.TrimSpace(in.Branch)
	def := strings.TrimSpace(in.DefaultBranch)
	switch {
	case branch == "" || in.NotBranch:
		return reasonNoBranch
	case def == "":
		return reasonDefaultUnknown
	case branch == def:
		return reasonDefault
	case !in.RemoteNamesOK:
		return reasonRemoteUnknown
	case in.OnRemote:
		return reasonOnRemote
	case in.Dirty:
		return reasonDirty
	case !in.OpenOK:
		return reasonOpenUnknown
	case in.OpenReview:
		return reasonOpenReview
	}
	return ""
}

// pruneScope is the per-row data every worktree's PruneInputs is built from.
type pruneScope struct {
	defaultBranch string
	remote        map[string]struct{}
	open          map[string]struct{}
	merged        map[string]board.MergedReview
	row           *board.ProjectSummary
}

// newPruneScope reads row once. The remote set is RemoteNames plus every
// forge branch row, so a branch still shown as a remote row is never pruned
// even when RemoteNames lags.
func newPruneScope(row *board.ProjectSummary) pruneScope {
	sc := pruneScope{
		defaultBranch: resolveDefaultBranch(row),
		remote:        map[string]struct{}{},
		open:          map[string]struct{}{},
		merged:        newestMergedByBranch(row.Merged),
		row:           row,
	}
	for _, name := range row.RemoteNames {
		if name = strings.TrimSpace(name); name != "" {
			sc.remote[name] = struct{}{}
		}
	}
	for _, b := range row.Branches {
		name := strings.TrimSpace(b.Name)
		if name == "" {
			continue
		}
		sc.remote[name] = struct{}{}
		if b.OpenReview {
			sc.open[name] = struct{}{}
		}
	}
	return sc
}

// inputs builds wt's PruneInputs.
func (sc pruneScope) inputs(wt board.LocalWorktree) PruneInputs {
	branch := strings.TrimSpace(wt.Branch)
	_, onRemote := sc.remote[branch]
	_, open := sc.open[branch]
	_, merged := sc.merged[branch]
	return PruneInputs{
		Branch:           branch,
		NotBranch:        wt.Detached || wt.Bare,
		Dirty:            wt.Dirty,
		DefaultBranch:    sc.defaultBranch,
		RemoteNamesOK:    sc.row.RemoteNamesOK,
		OnRemote:         onRemote,
		OpenOK:           sc.row.OpenOK,
		OpenReview:       open,
		Merged:           merged,
		MergedOK:         sc.row.MergedOK,
		MergedChecked:    sc.row.MergedChecked[branch],
		ContentOnDefault: wt.ContentOnDefault,
	}
}

// enrichPruneHints sets PruneHint on each local worktree from pruneVerdict:
// safe when the remote head is gone, the tree is clean, nothing is open and
// either a merged review exists or the content is already on default;
// likely when a per-branch merged lookup confirmed no match. The primary
// worktree is not exempt; only the default branch is.
//
// The default prefers the forge-marked default over local origin/HEAD, which
// can still point at a feature branch after a non-default clone. When the
// forge default is known, Local.DefaultBranch is aligned to it for the
// switch in a later prune.
func enrichPruneHints(row *board.ProjectSummary) {
	if row == nil || row.Local == nil {
		return
	}
	sc := newPruneScope(row)
	if sc.defaultBranch != "" {
		row.Local.DefaultBranch = sc.defaultBranch
	}
	for i := range row.Local.Worktrees {
		wt := &row.Local.Worktrees[i]
		v := pruneVerdict(sc.inputs(*wt))
		wt.PruneHint = v.Hint
		if m, ok := sc.merged[strings.TrimSpace(wt.Branch)]; ok && v.Allow {
			wt.MergedID = m.ID
			wt.MergedURL = m.URL
			wt.MergedAt = m.MergedAt
		}
	}
}

// resolveDefaultBranch prefers the forge default branch over local origin/HEAD.
func resolveDefaultBranch(row *board.ProjectSummary) string {
	for _, b := range row.Branches {
		if !b.Default {
			continue
		}
		if name := strings.TrimSpace(b.Name); name != "" {
			return name
		}
	}
	if row.Local == nil {
		return ""
	}
	return strings.TrimSpace(row.Local.DefaultBranch)
}

// newestMergedByBranch keeps the most recently merged review per branch.
func newestMergedByBranch(merged []board.MergedReview) map[string]board.MergedReview {
	out := make(map[string]board.MergedReview, len(merged))
	for _, m := range merged {
		name := strings.TrimSpace(m.Branch)
		if name == "" {
			continue
		}
		prev, ok := out[name]
		if !ok {
			out[name] = m
			continue
		}
		tNew, okNew := parseMergedAt(m.MergedAt)
		tOld, okOld := parseMergedAt(prev.MergedAt)
		if okNew && (!okOld || tNew.After(tOld)) {
			out[name] = m
		}
	}
	return out
}

// parseMergedAt reads the RFC 3339 stamps forge loaders write.
func parseMergedAt(raw string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(raw))
	return t, err == nil
}
