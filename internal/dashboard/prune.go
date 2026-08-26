package dashboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/forge"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

// BadRequestError is a client/validation failure for prune APIs.
type BadRequestError struct {
	Msg string
}

func (e BadRequestError) Error() string {
	if e.Msg == "" {
		return "bad request"
	}
	return e.Msg
}

// IsBadRequest reports whether err is a prune validation / not-safe failure.
func IsBadRequest(err error) bool {
	var e BadRequestError
	return errors.As(err, &e)
}

func badRequest(msg string) error {
	return BadRequestError{Msg: msg}
}

// PruneSafeRequest identifies a local checkout marked safe to remove.
type PruneSafeRequest struct {
	ProjectID    string `json:"project_id"`
	Branch       string `json:"branch"`
	WorktreePath string `json:"worktree_path"`
}

// PruneSafe re-checks forge prune hints, then removes the local checkout when still safe.
func (s *Service) PruneSafe(ctx context.Context, doc config.File, req PruneSafeRequest) error {
	if s == nil || s.Local == nil {
		return fmt.Errorf("local git inspector missing")
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	worktreePath := strings.TrimSpace(req.WorktreePath)
	if projectID == "" || branch == "" || worktreePath == "" {
		return badRequest("project_id, branch, and worktree_path are required")
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return badRequest("unknown project")
	}

	abs, err := localgit.ExpandPath(worktreePath)
	if err != nil {
		return badRequest(fmt.Sprintf("worktree path: %v", err))
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	opts := forge.SummaryOpts{
		Fresh:     true,
		Cache:     s.Cache,
		HeadsTTL:  0,
		MergedTTL: 0,
	}
	row := s.summarize(ctx, p, opts)
	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	row.Local = s.attachLocal(ctx, p, disc, projectLabelsByKey(doc.Projects))
	forge.EnrichPruneHints(&row)

	wt, ok := findSafeWorktree(row.Local, branch, abs)
	if !ok {
		return badRequest(fmt.Sprintf("branch %q is not safe to remove (re-check prune hints)", branch))
	}
	defaultBranch := strings.TrimSpace(row.Local.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	if err := s.Local.RemoveSafeCheckout(ctx, wt.Path, branch, defaultBranch); err != nil {
		return fmt.Errorf("remove checkout: %w", err)
	}
	return nil
}

func findSafeWorktree(local *forge.LocalStatus, branch, absPath string) (forge.LocalWorktree, bool) {
	if local == nil {
		return forge.LocalWorktree{}, false
	}
	branch = strings.TrimSpace(branch)
	for _, wt := range local.Worktrees {
		if strings.TrimSpace(wt.Branch) != branch {
			continue
		}
		wtPath := filepath.Clean(wt.Path)
		if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
			wtPath = resolved
		}
		if wtPath != absPath {
			continue
		}
		if wt.PruneHint != forge.PruneSafe {
			continue
		}
		wt.Path = wtPath
		return wt, true
	}
	return forge.LocalWorktree{}, false
}
