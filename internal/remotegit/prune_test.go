package remotegit

import (
	"testing"

	"github.com/behaviorengineering/gitboard/internal/board"
)

func TestEnrichPruneHintsSafeAndLikely(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNames:   []string{"main", "feat/open", "feat/still-remote"},
		RemoteNamesOK: true,
		Branches: []board.BranchRef{
			{Name: "main", Default: true},
			{Name: "feat/open", OpenReview: true},
		},
		Merged: []board.MergedReview{
			{Branch: "feat/merged", ID: 42, URL: "https://example/pr/42", MergedAt: "2026-08-20T10:00:00Z"},
			{Branch: "feat/merged", ID: 41, URL: "https://example/pr/41", MergedAt: "2026-08-10T10:00:00Z"},
		},
		MergedOK: true,
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/repo", Branch: "main", Main: true},
				{Path: "/repo-merged", Branch: "feat/merged"},
				{Path: "/repo-likely", Branch: "feat/gone"},
				{Path: "/repo-dirty", Branch: "feat/dirty", Dirty: true},
				{Path: "/repo-remote", Branch: "feat/still-remote"},
				{Path: "/repo-open", Branch: "feat/open"},
			},
		},
	}

	EnrichPruneHints(summary)

	byBranch := map[string]board.LocalWorktree{}
	for _, wt := range summary.Local.Worktrees {
		byBranch[wt.Branch] = wt
	}

	if got := byBranch["feat/merged"]; got.PruneHint != board.PruneSafe || got.MergedID != 42 {
		t.Fatalf("merged: %+v", got)
	}
	if got := byBranch["feat/gone"]; got.PruneHint != board.PruneLikely {
		t.Fatalf("likely: %+v", got)
	}
	for _, name := range []string{"main", "feat/dirty", "feat/still-remote", "feat/open"} {
		if got := byBranch[name]; got.PruneHint != "" {
			t.Fatalf("%s should have no prune hint: %+v", name, got)
		}
	}
}

func TestEnrichPruneHintsUsesFullRemoteNames(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNames:   []string{"main", "feat/hidden-remote"},
		RemoteNamesOK: true,
		Branches: []board.BranchRef{
			{Name: "main", Default: true},
		},
		MergedOK: true,
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/hidden-remote"},
			},
		},
	}
	EnrichPruneHints(summary)
	if got := summary.Local.Worktrees[0].PruneHint; got != "" {
		t.Fatalf("should not prune live remote outside UI list, got %q", got)
	}
}

func TestEnrichPruneHintsRemoteNamesNotOKSuppressesAll(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNamesOK: false,
		Merged: []board.MergedReview{
			{Branch: "feat/merged", ID: 1, MergedAt: "2026-08-20T10:00:00Z"},
		},
		MergedOK: true,
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/merged"},
			},
		},
	}
	EnrichPruneHints(summary)
	if got := summary.Local.Worktrees[0].PruneHint; got != "" {
		t.Fatalf("heads unknown must suppress prune, got %q", got)
	}
}

func TestEnrichPruneHintsEmptyRemoteNamesOKStillAllowsGone(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNames:   nil,
		RemoteNamesOK: true,
		Merged: []board.MergedReview{
			{Branch: "feat/merged", ID: 1, MergedAt: "2026-08-20T10:00:00Z"},
		},
		MergedOK: true,
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/merged"},
			},
		},
	}
	EnrichPruneHints(summary)
	if got := summary.Local.Worktrees[0].PruneHint; got != board.PruneSafe {
		t.Fatalf("want safe when heads OK and empty, got %q", got)
	}
}

func TestEnrichPruneHintsMergedNotOKSuppressesLikely(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		MergedOK:      false,
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/gone"},
			},
		},
	}
	EnrichPruneHints(summary)
	if got := summary.Local.Worktrees[0].PruneHint; got != "" {
		t.Fatalf("likely suppressed when mergedOK=false, got %q", got)
	}
}

func TestEnrichPruneHintsPrimaryCheckout(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		Branches: []board.BranchRef{
			{Name: "main", Default: true},
		},
		Merged: []board.MergedReview{
			{Branch: "feat/primary-merged", ID: 7, URL: "https://example/pr/7", MergedAt: "2026-08-24T01:00:00Z"},
		},
		MergedOK: true,
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/repo", Branch: "feat/primary-merged", Main: true},
				{Path: "/repo-gone", Branch: "feat/primary-gone", Main: true},
			},
		},
	}

	EnrichPruneHints(summary)

	byBranch := map[string]board.LocalWorktree{}
	for _, wt := range summary.Local.Worktrees {
		byBranch[wt.Branch] = wt
	}

	if got := byBranch["feat/primary-merged"]; got.PruneHint != board.PruneSafe || got.MergedID != 7 {
		t.Fatalf("primary checkout merged: %+v", got)
	}
	if got := byBranch["feat/primary-gone"]; got.PruneHint != board.PruneLikely {
		t.Fatalf("primary checkout likely: %+v", got)
	}
}

func TestEnrichPruneHintsMultiAppearanceWorktrees(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		Branches: []board.BranchRef{
			{Name: "main", Default: true},
		},
		Merged: []board.MergedReview{
			{Branch: "feat/a", ID: 1, URL: "https://example/pr/1", MergedAt: "2026-08-20T10:00:00Z"},
		},
		MergedOK: true,
		Local: &board.LocalStatus{
			Mapped:        true,
			Path:          "/standalone",
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/standalone", Branch: "feat/a", Main: true, AppearancePath: "/standalone", AppearanceLabel: "~/standalone"},
				{Path: "/parent/providers/repo", Branch: "feat/a", Main: true, AppearancePath: "/parent/providers/repo", AppearanceLabel: "parent → providers/repo"},
				{Path: "/parent/providers/repo", Branch: "feat/gone", Main: true, AppearancePath: "/parent/providers/repo", AppearanceLabel: "parent → providers/repo"},
			},
		},
	}
	EnrichPruneHints(summary)

	var safe, likely int
	for _, wt := range summary.Local.Worktrees {
		switch {
		case wt.Branch == "feat/a" && wt.PruneHint == board.PruneSafe:
			safe++
		case wt.Branch == "feat/gone" && wt.PruneHint == board.PruneLikely:
			likely++
		}
	}
	if safe != 2 {
		t.Fatalf("safe prune on both appearances: got %d", safe)
	}
	if likely != 1 {
		t.Fatalf("likely prune: got %d", likely)
	}
}

func TestEnrichPruneHintsNilSafe(t *testing.T) {
	EnrichPruneHints(nil)
	EnrichPruneHints(&board.ProjectSummary{})
}

func TestEnrichPruneHintsPrefersForgeDefaultOverStaleOriginHEAD(t *testing.T) {
	summary := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		Branches: []board.BranchRef{
			{Name: "main", Default: true},
		},
		Merged: []board.MergedReview{
			{Branch: "feat/config-sync-dashboard", ID: 4, URL: "https://example/pr/4", MergedAt: "2026-08-26T05:12:17Z"},
		},
		MergedOK: true,
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "feat/config-sync-dashboard",
			Worktrees: []board.LocalWorktree{
				{Path: "/repo", Branch: "feat/config-sync-dashboard", Main: true},
			},
		},
	}

	EnrichPruneHints(summary)

	if got := summary.Local.DefaultBranch; got != "main" {
		t.Fatalf("aligned default_branch: got %q want main", got)
	}
	if got := summary.Local.Worktrees[0].PruneHint; got != board.PruneSafe {
		t.Fatalf("stale origin/HEAD must not block safe prune, got %q", got)
	}
}
