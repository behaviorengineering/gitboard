package dashboard

import (
	"context"
	"fmt"
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

func TestConfirmMergedForCandidatesHit(t *testing.T) {
	fx := &fakeExec{
		responses: map[string][]byte{
			"--head feat/gone": []byte(`[{"number":25,"headRefName":"feat/gone","url":"https://example/pr/25","mergedAt":"2026-09-12T02:44:00Z"}]`),
		},
	}
	s := newTestService(fx, nil)
	p := boardconfig.Project{ID: "app", Host: boardconfig.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		OpenOK:        true,
		MergedOK:      true,
		Branches: []board.BranchRef{
			{Name: "main", Default: true},
		},
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/gone"},
			},
		},
	}
	s.confirmMergedForCandidates(context.Background(), p, row, remotegit.SummaryOpts{
		Cache:    s.cache,
		HeadsTTL: 0,
		Fresh:    true,
	})
	if !row.MergedChecked["feat/gone"] {
		t.Fatal("expected MergedChecked")
	}
	if len(row.Merged) != 1 || row.Merged[0].ID != 25 {
		t.Fatalf("merged: %+v", row.Merged)
	}
	enrichPruneHints(row)
	if row.Local.Worktrees[0].PruneHint != board.PruneSafe {
		t.Fatalf("hint=%q", row.Local.Worktrees[0].PruneHint)
	}
}

func TestConfirmMergedForCandidatesMissIsLikely(t *testing.T) {
	fx := &fakeExec{
		responses: map[string][]byte{
			"--head feat/gone": []byte(`[]`),
		},
	}
	s := newTestService(fx, nil)
	p := boardconfig.Project{ID: "app", Host: boardconfig.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		OpenOK:        true,
		MergedOK:      true,
		Branches: []board.BranchRef{
			{Name: "main", Default: true},
		},
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/gone"},
			},
		},
	}
	s.confirmMergedForCandidates(context.Background(), p, row, remotegit.SummaryOpts{
		Cache:    s.cache,
		HeadsTTL: 0,
		Fresh:    true,
	})
	if !row.MergedChecked["feat/gone"] {
		t.Fatal("miss must still be checked")
	}
	enrichPruneHints(row)
	if row.Local.Worktrees[0].PruneHint != board.PruneLikely {
		t.Fatalf("hint=%q", row.Local.Worktrees[0].PruneHint)
	}
}

func TestConfirmMergedForCandidatesLookupErrorNoLikely(t *testing.T) {
	fx := &errExec{err: fmt.Errorf("glab failed")}
	s := newTestService(fx, nil)
	p := boardconfig.Project{ID: "app", Host: boardconfig.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		OpenOK:        true,
		MergedOK:      true,
		Local: &board.LocalStatus{
			Mapped: true,
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/gone"},
			},
		},
	}
	s.confirmMergedForCandidates(context.Background(), p, row, remotegit.SummaryOpts{
		Cache: s.cache,
		Fresh: true,
	})
	if row.MergedChecked["feat/gone"] {
		t.Fatal("error must not mark checked")
	}
	enrichPruneHints(row)
	if row.Local.Worktrees[0].PruneHint != "" {
		t.Fatalf("hint=%q", row.Local.Worktrees[0].PruneHint)
	}
}

func TestConfirmMergedForCandidatesSkipsBulkHit(t *testing.T) {
	fx := &errExec{err: fmt.Errorf("should not lookup")}
	s := newTestService(fx, nil)
	p := boardconfig.Project{ID: "app", Host: boardconfig.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		OpenOK:        true,
		MergedOK:      true,
		Merged: []board.MergedReview{
			{Branch: "feat/gone", ID: 9},
		},
		Local: &board.LocalStatus{
			Mapped:        true,
			DefaultBranch: "main",
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/gone"},
			},
		},
	}
	s.confirmMergedForCandidates(context.Background(), p, row, remotegit.SummaryOpts{
		Cache: s.cache,
		Fresh: true,
	})
	enrichPruneHints(row)
	if row.Local.Worktrees[0].PruneHint != board.PruneSafe {
		t.Fatalf("bulk hit should stay safe without lookup, hint=%q", row.Local.Worktrees[0].PruneHint)
	}
}

type errExec struct {
	err error
}

func (f *errExec) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *errExec) Run(context.Context, string, ...string) ([]byte, error) {
	return nil, f.err
}

func (f *errExec) RunJSON(context.Context, string, ...string) ([]byte, error) {
	return nil, f.err
}
