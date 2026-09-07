package dashboard

import (
	"context"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/board"
)

// annotateContentOnDefault sets LocalWorktree.ContentOnDefault for remote-gone,
// clean, non-default worktrees before EnrichPruneHints runs.
func (s *Service) annotateContentOnDefault(ctx context.Context, row *board.ProjectSummary) {
	if s == nil || s.Local == nil || row == nil || row.Local == nil || !row.Local.Mapped {
		return
	}
	if !row.RemoteNamesOK {
		return
	}

	remote := make(map[string]struct{}, len(row.RemoteNames))
	for _, name := range row.RemoteNames {
		name = strings.TrimSpace(name)
		if name != "" {
			remote[name] = struct{}{}
		}
	}

	defaultName := strings.TrimSpace(row.Local.DefaultBranch)
	for _, b := range row.Branches {
		if b.Default {
			if n := strings.TrimSpace(b.Name); n != "" {
				defaultName = n
			}
			break
		}
	}
	if defaultName == "" {
		defaultName = "main"
	}

	repoPath := strings.TrimSpace(row.Local.Path)
	if repoPath == "" {
		return
	}

	for i := range row.Local.Worktrees {
		wt := &row.Local.Worktrees[i]
		branch := strings.TrimSpace(wt.Branch)
		if branch == "" || wt.Detached || wt.Bare || wt.Dirty {
			continue
		}
		if branch == defaultName {
			continue
		}
		if _, onRemote := remote[branch]; onRemote {
			continue
		}
		checkPath := strings.TrimSpace(wt.Path)
		if checkPath == "" {
			checkPath = repoPath
		}
		ok, _, err := s.Local.ContentOnDefault(ctx, checkPath, branch, defaultName)
		if err != nil || !ok {
			continue
		}
		wt.ContentOnDefault = true
	}
}
