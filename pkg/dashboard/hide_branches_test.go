package dashboard

import (
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/board"
)

func TestFilterHiddenBranches(t *testing.T) {
	branches := []board.BranchRef{
		{Name: "main"},
		{Name: "feat/x"},
		{Name: "majordomo-context/polypus"},
		{Name: "majordomo-context/polypus-update"},
	}

	t.Run("empty patterns keep all", func(t *testing.T) {
		got := filterHiddenBranches(branches, nil)
		if len(got) != len(branches) {
			t.Fatalf("len=%d want %d", len(got), len(branches))
		}
		got = filterHiddenBranches(branches, []string{})
		if len(got) != len(branches) {
			t.Fatalf("empty slice len=%d want %d", len(got), len(branches))
		}
	})

	t.Run("prefix star", func(t *testing.T) {
		got := filterHiddenBranches(branches, []string{"majordomo-context/*"})
		if len(got) != 2 {
			t.Fatalf("len=%d want 2: %+v", len(got), got)
		}
		if got[0].Name != "main" || got[1].Name != "feat/x" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("exact name", func(t *testing.T) {
		got := filterHiddenBranches(branches, []string{"feat/x"})
		if len(got) != 3 {
			t.Fatalf("len=%d want 3", len(got))
		}
		for _, b := range got {
			if b.Name == "feat/x" {
				t.Fatal("feat/x should be hidden")
			}
		}
	})

	t.Run("whitespace and invalid patterns skipped", func(t *testing.T) {
		got := filterHiddenBranches(branches, []string{"", "  ", "["})
		if len(got) != len(branches) {
			t.Fatalf("len=%d want %d", len(got), len(branches))
		}
	})

	t.Run("all hidden returns nil", func(t *testing.T) {
		got := filterHiddenBranches([]board.BranchRef{{Name: "tmp"}}, []string{"tmp"})
		if got != nil {
			t.Fatalf("want nil, got %+v", got)
		}
	})
}

func TestSyncOpenItemsToVisibleBranches(t *testing.T) {
	t.Run("github recounts pull requests", func(t *testing.T) {
		row := board.ProjectSummary{
			Host:      "github",
			OpenItems: board.OpenItems{PullRequests: 3},
			Branches: []board.BranchRef{
				{Name: "main"},
				{Name: "feat/a", OpenReview: true},
			},
		}
		syncOpenItemsToVisibleBranches(&row)
		if row.OpenItems.PullRequests != 1 || row.OpenItems.MergeRequests != 0 {
			t.Fatalf("open_items: %+v", row.OpenItems)
		}
	})

	t.Run("gitlab recounts merge requests", func(t *testing.T) {
		row := board.ProjectSummary{
			Host:      "gitlab",
			OpenItems: board.OpenItems{MergeRequests: 2},
			Branches: []board.BranchRef{
				{Name: "main"},
			},
		}
		syncOpenItemsToVisibleBranches(&row)
		if row.OpenItems.PullRequests != 0 || row.OpenItems.MergeRequests != 0 {
			t.Fatalf("open_items: %+v", row.OpenItems)
		}
	})

	t.Run("nil row is no-op", func(t *testing.T) {
		syncOpenItemsToVisibleBranches(nil)
	})
}
