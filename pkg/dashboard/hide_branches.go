package dashboard

import (
	"path"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
)

// filterHiddenBranches drops branches whose names match any path.Match pattern.
// Empty or whitespace-only patterns are ignored. Invalid patterns never match.
// The input slice is not modified; nil in with no removals returns nil.
func filterHiddenBranches(branches []board.BranchRef, patterns []string) []board.BranchRef {
	if len(branches) == 0 || len(patterns) == 0 {
		return branches
	}
	compiled := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		compiled = append(compiled, p)
	}
	if len(compiled) == 0 {
		return branches
	}

	out := make([]board.BranchRef, 0, len(branches))
	for _, b := range branches {
		if branchHidden(b.Name, compiled) {
			continue
		}
		out = append(out, b)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func branchHidden(name string, patterns []string) bool {
	for _, pat := range patterns {
		ok, err := path.Match(pat, name)
		if err != nil {
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

// syncOpenItemsToVisibleBranches resets open_items to the count of still-visible
// branches marked open_review. Call after filterHiddenBranches so board filters
// and attention ranking do not treat hidden PR/MR heads as open review work.
func syncOpenItemsToVisibleBranches(row *board.ProjectSummary) {
	if row == nil {
		return
	}
	n := 0
	for _, b := range row.Branches {
		if b.OpenReview {
			n++
		}
	}
	row.OpenItems = board.OpenItems{}
	if strings.EqualFold(strings.TrimSpace(row.Host), "gitlab") {
		row.OpenItems.MergeRequests = n
		return
	}
	row.OpenItems.PullRequests = n
}
