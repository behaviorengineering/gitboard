package dashboard

import (
	"path"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/board"
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
