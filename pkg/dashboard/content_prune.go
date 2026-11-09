package dashboard

import (
	"context"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
)

// annotateContentOnDefault sets LocalWorktree.ContentOnDefault for every
// prune candidate (pruneVerdict passed its gates) before enrichPruneHints
// runs. A failed check leaves it false (fail closed).
func (s *Service) annotateContentOnDefault(ctx context.Context, row *board.ProjectSummary) {
	if s == nil || s.local == nil || row == nil || row.Local == nil || !row.Local.Mapped {
		return
	}
	repoPath := strings.TrimSpace(row.Local.Path)
	if repoPath == "" {
		return
	}
	sc := newPruneScope(row)
	for i := range row.Local.Worktrees {
		wt := &row.Local.Worktrees[i]
		if !pruneVerdict(sc.inputs(*wt)).Candidate {
			continue
		}
		checkPath := strings.TrimSpace(wt.Path)
		if checkPath == "" {
			checkPath = repoPath
		}
		ok, _, err := s.local.ContentOnDefault(ctx, checkPath, strings.TrimSpace(wt.Branch), sc.defaultBranch)
		wt.ContentOnDefault = err == nil && ok
	}
}
