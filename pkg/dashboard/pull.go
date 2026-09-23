package dashboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
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
	OK              bool   `json:"ok"`
	Branch          string `json:"branch"`
	Path            string `json:"path"`
	FetchMs         int64  `json:"fetch_ms,omitempty"`
	SubmodulePreMs  int64  `json:"submodule_pre_ms,omitempty"`
	MergeMs         int64  `json:"merge_ms,omitempty"`
	SubmodulePostMs int64  `json:"submodule_post_ms,omitempty"`
}

// PullFF validates the mapped checkout, then fast-forwards the branch from origin.
// Ahead/behind is decided inside PullFFOnly after a fresh fetch.
// Already up to date is success: the board may still show a stale ↓N until refresh.
// Validation uses a path allowlist (cached scan) instead of a full attachLocal forge refresh.
func (c *Commands) PullFF(ctx context.Context, doc config.File, req PullFFRequest) (PullFFResult, error) {
	var zero PullFFResult
	s, err := c.requireLocal()
	if err != nil {
		return zero, err
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	repoPath := strings.TrimSpace(req.RepoPath)
	if projectID == "" || branch == "" || repoPath == "" {
		return zero, badRequest("project_id, branch, and repo_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return zero, badRequestCause("", err)
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return zero, badRequest("unknown project")
	}

	abs, err := localgit.ExpandPath(repoPath)
	if err != nil {
		return zero, badRequestCause(fmt.Sprintf("repo path: %v", err), err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	allowed, err := s.mappedPathAllowlist(ctx, doc, p, abs)
	if err != nil {
		return zero, err
	}
	if !allowed {
		return zero, badRequest("repo_path is not a mapped checkout for this project")
	}

	common, err := localgit.ResolveCommonDir(ctx, s.Local.CommonGitDir, abs)
	if err != nil {
		return zero, badRequestCause("resolve common git dir", err)
	}
	lease, err := s.Mutations.Acquire(ctx, common)
	if err != nil {
		if errors.Is(err, localgit.ErrMutationCanceled) {
			return zero, err
		}
		return zero, fmt.Errorf("mutation lock: %w", err)
	}
	defer lease.Release()

	if err := c.ensureWritableIndex(ctx, abs, req.ClearIndexLock); err != nil {
		return zero, err
	}

	var phases localgit.PullPhases
	err = s.pullFFOnlyTimed(ctx, abs, branch, &phases)
	if err != nil && localgit.IsIndexLockError(err) {
		if lockErr := c.ensureWritableIndex(ctx, abs, req.ClearIndexLock); lockErr != nil {
			return zero, lockErr
		}
		if req.ClearIndexLock {
			var retryPhases localgit.PullPhases
			retryErr := s.pullFFOnlyTimed(ctx, abs, branch, &retryPhases)
			phases.FetchMs += retryPhases.FetchMs
			phases.SubmodulePreMs += retryPhases.SubmodulePreMs
			phases.MergeMs += retryPhases.MergeMs
			phases.SubmodulePostMs += retryPhases.SubmodulePostMs
			err = retryErr
		}
	}
	if err != nil {
		// Idempotent with ensureBranchFFFromOrigin: a fresh fetch can show the
		// checkout is already caught up while the board still painted ↓N.
		if errors.Is(err, localgit.ErrUpToDate) {
			lease.BumpEpoch()
			s.InvalidateOrigin(common)
			s.InvalidateProject(string(p.Host), p.Path)
			return PullFFResult{OK: true, Branch: branch, Path: abs, FetchMs: phases.FetchMs, SubmodulePreMs: phases.SubmodulePreMs, MergeMs: phases.MergeMs, SubmodulePostMs: phases.SubmodulePostMs}, nil
		}
		if errors.Is(err, localgit.ErrInvalidBranch) ||
			errors.Is(err, localgit.ErrMissingBranch) ||
			errors.Is(err, localgit.ErrDirtyTree) ||
			errors.Is(err, localgit.ErrDiverged) {
			return zero, badRequestCause("", err)
		}
		return zero, fmt.Errorf("pull ff-only: %w", err)
	}
	lease.BumpEpoch()
	s.InvalidateOrigin(common)
	s.InvalidateProject(string(p.Host), p.Path)
	return PullFFResult{OK: true, Branch: branch, Path: abs, FetchMs: phases.FetchMs, SubmodulePreMs: phases.SubmodulePreMs, MergeMs: phases.MergeMs, SubmodulePostMs: phases.SubmodulePostMs}, nil
}

// pullFFOnlyTimed runs the timed pull when the local backend supports phases,
// otherwise falls back to PullFFOnly without timings.
func (s *Service) pullFFOnlyTimed(ctx context.Context, abs, branch string, phases *localgit.PullPhases) error {
	if s == nil || s.Local == nil {
		return ErrLocalInspectorMissing
	}
	if timed, ok := s.Local.(interface {
		PullFFOnlyWithPhases(context.Context, string, string) (localgit.PullPhases, error)
	}); ok {
		got, err := timed.PullFFOnlyWithPhases(ctx, abs, branch)
		if phases != nil {
			*phases = got
		}
		return err
	}
	return s.Local.PullFFOnly(ctx, abs, branch)
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
	s, err := c.requireLocal()
	if err != nil {
		return "", err
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
		return "", badRequestCause(fmt.Sprintf("path: %v", err), err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	allowed, err := s.mappedPathAllowlist(ctx, doc, p, abs)
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", badRequest("path is not a mapped checkout for this project")
	}
	return abs, nil
}
