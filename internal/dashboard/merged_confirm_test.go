package dashboard

import (
	"context"
	"fmt"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/remotegit"
)

func TestConfirmMergedForCandidatesHit(t *testing.T) {
	fx := &fakeExec{
		responses: map[string][]byte{
			"--head feat/gone": []byte(`[{"number":25,"headRefName":"feat/gone","url":"https://example/pr/25","mergedAt":"2026-09-12T02:44:00Z"}]`),
		},
	}
	s := New(remotegit.NewGitHub(fx), nil, nil)
	p := config.Project{ID: "app", Host: config.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
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
		Cache:    s.Cache,
		HeadsTTL: 0,
		Fresh:    true,
	})
	if !row.MergedChecked["feat/gone"] {
		t.Fatal("expected MergedChecked")
	}
	if len(row.Merged) != 1 || row.Merged[0].ID != 25 {
		t.Fatalf("merged: %+v", row.Merged)
	}
	remotegit.EnrichPruneHints(row)
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
	s := New(remotegit.NewGitHub(fx), nil, nil)
	p := config.Project{ID: "app", Host: config.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
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
		Cache:    s.Cache,
		HeadsTTL: 0,
		Fresh:    true,
	})
	if !row.MergedChecked["feat/gone"] {
		t.Fatal("miss must still be checked")
	}
	remotegit.EnrichPruneHints(row)
	if row.Local.Worktrees[0].PruneHint != board.PruneLikely {
		t.Fatalf("hint=%q", row.Local.Worktrees[0].PruneHint)
	}
}

func TestConfirmMergedForCandidatesLookupErrorNoLikely(t *testing.T) {
	fx := &errExec{err: fmt.Errorf("glab failed")}
	s := New(remotegit.NewGitHub(fx), nil, nil)
	p := config.Project{ID: "app", Host: config.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		MergedOK:      true,
		Local: &board.LocalStatus{
			Mapped: true,
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/gone"},
			},
		},
	}
	s.confirmMergedForCandidates(context.Background(), p, row, remotegit.SummaryOpts{
		Cache: s.Cache,
		Fresh: true,
	})
	if row.MergedChecked["feat/gone"] {
		t.Fatal("error must not mark checked")
	}
	remotegit.EnrichPruneHints(row)
	if row.Local.Worktrees[0].PruneHint != "" {
		t.Fatalf("hint=%q", row.Local.Worktrees[0].PruneHint)
	}
}

func TestConfirmMergedForCandidatesSkipsBulkHit(t *testing.T) {
	fx := &errExec{err: fmt.Errorf("should not lookup")}
	s := New(remotegit.NewGitHub(fx), nil, nil)
	p := config.Project{ID: "app", Host: config.HostGitHub, Path: "acme/app"}
	row := &board.ProjectSummary{
		RemoteNames:   []string{"main"},
		RemoteNamesOK: true,
		MergedOK:      true,
		Merged: []board.MergedReview{
			{Branch: "feat/gone", ID: 9},
		},
		Local: &board.LocalStatus{
			Mapped: true,
			Worktrees: []board.LocalWorktree{
				{Path: "/wt", Branch: "feat/gone"},
			},
		},
	}
	s.confirmMergedForCandidates(context.Background(), p, row, remotegit.SummaryOpts{
		Cache: s.Cache,
		Fresh: true,
	})
	remotegit.EnrichPruneHints(row)
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
