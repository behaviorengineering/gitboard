package dashboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

// PullFFRequest identifies a local branch to fast-forward from origin.
type PullFFRequest struct {
	ProjectID      string `json:"project_id"`
	Branch         string `json:"branch"`
	RepoPath       string `json:"repo_path"`
	ClearIndexLock bool   `json:"clear_index_lock,omitempty"`
}

// PullFFResult is returned after a successful fast-forward.
type PullFFResult struct {
	OK     bool   `json:"ok"`
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

// PullFF validates the mapped checkout, then fast-forwards the branch from origin.
// Ahead/behind is decided inside PullFFOnly after a fresh fetch.
// Already up to date is success: the board may still show a stale ↓N until refresh.
func (c *Commands) PullFF(ctx context.Context, doc config.File, req PullFFRequest) (PullFFResult, error) {
	var zero PullFFResult
	s := c.Service
	if s == nil || s.Local == nil {
		return zero, fmt.Errorf("local git inspector missing")
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	repoPath := strings.TrimSpace(req.RepoPath)
	if projectID == "" || branch == "" || repoPath == "" {
		return zero, badRequest("project_id, branch, and repo_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return zero, badRequest(err.Error())
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return zero, badRequest("unknown project")
	}

	abs, err := localgit.ExpandPath(repoPath)
	if err != nil {
		return zero, badRequest(fmt.Sprintf("repo path: %v", err))
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	local := s.attachLocal(ctx, p, disc, projectLabelsByKey(doc.Projects), false, time.Duration(doc.EffectiveFetchSeconds())*time.Second)
	if local == nil || !local.Mapped {
		return zero, badRequest("project has no mapped local checkout")
	}
	if !repoPathAllowed(local, abs) {
		return zero, badRequest("repo_path is not a mapped checkout for this project")
	}

	if err := c.ensureWritableIndex(ctx, abs, req.ClearIndexLock); err != nil {
		return zero, err
	}

	err = s.Local.PullFFOnly(ctx, abs, branch)
	if err != nil && localgit.IsIndexLockError(err) {
		if lockErr := c.ensureWritableIndex(ctx, abs, req.ClearIndexLock); lockErr != nil {
			return zero, lockErr
		}
		if req.ClearIndexLock {
			err = s.Local.PullFFOnly(ctx, abs, branch)
		}
	}
	if err != nil {
		// Idempotent with ensureBranchFFFromOrigin: a fresh fetch can show the
		// checkout is already caught up while the board still painted ↓N.
		if errors.Is(err, localgit.ErrUpToDate) {
			return PullFFResult{OK: true, Branch: branch, Path: abs}, nil
		}
		if errors.Is(err, localgit.ErrInvalidBranch) ||
			errors.Is(err, localgit.ErrMissingBranch) ||
			errors.Is(err, localgit.ErrDirtyTree) ||
			errors.Is(err, localgit.ErrDiverged) {
			return zero, badRequest(err.Error())
		}
		return zero, fmt.Errorf("pull ff-only: %w", err)
	}
	return PullFFResult{OK: true, Branch: branch, Path: abs}, nil
}

func repoPathAllowed(local *board.LocalStatus, abs string) bool {
	return repoPathAllowedCanon(local, abs, nil)
}

func repoPathAllowedCanon(local *board.LocalStatus, abs string, canon func(string) string) bool {
	if local == nil {
		return false
	}
	check := func(c string) bool {
		return sameCheckoutPath(c, abs, canon)
	}
	if check(local.Path) {
		return true
	}
	for _, wt := range local.Worktrees {
		if check(wt.Path) || check(wt.AppearancePath) {
			return true
		}
	}
	for _, app := range local.Appearances {
		if check(app.Path) {
			return true
		}
		for _, wt := range app.Worktrees {
			if check(wt.Path) || check(wt.AppearancePath) {
				return true
			}
		}
	}
	return false
}

// RequireMappedPath expands path and ensures it belongs to the project's mapped checkouts.
func (c *Commands) RequireMappedPath(ctx context.Context, doc config.File, projectID, repoPath string) (string, error) {
	s := c.Service
	if s == nil || s.Local == nil {
		return "", fmt.Errorf("local git inspector missing")
	}
	projectID = strings.TrimSpace(projectID)
	repoPath = strings.TrimSpace(repoPath)
	if projectID == "" || repoPath == "" {
		return "", badRequest("project_id and path are required")
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return "", badRequest("unknown project")
	}
	abs, err := localgit.ExpandPath(repoPath)
	if err != nil {
		return "", badRequest(fmt.Sprintf("path: %v", err))
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	local := s.attachLocal(ctx, p, disc, projectLabelsByKey(doc.Projects), false, time.Duration(doc.EffectiveFetchSeconds())*time.Second)
	if local == nil || !local.Mapped {
		return "", badRequest("project has no mapped local checkout")
	}
	if !repoPathAllowed(local, abs) {
		return "", badRequest("path is not a mapped checkout for this project")
	}
	return abs, nil
}
