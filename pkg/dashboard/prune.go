package dashboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// BadRequestError is a client/validation failure for command APIs.
type BadRequestError struct {
	Msg   string
	Cause error `json:"-"`
}

func (e BadRequestError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "bad request"
}

// Unwrap returns the underlying cause when present.
func (e BadRequestError) Unwrap() error {
	return e.Cause
}

// IsBadRequest reports whether err is a prune validation / not-safe failure.
func IsBadRequest(err error) bool {
	var e BadRequestError
	return errors.As(err, &e)
}

func badRequest(msg string) error {
	return BadRequestError{Msg: msg}
}

func badRequestCause(msg string, cause error) error {
	if cause == nil {
		return badRequest(msg)
	}
	if msg == "" {
		return BadRequestError{Msg: cause.Error(), Cause: cause}
	}
	return BadRequestError{Msg: msg, Cause: cause}
}

// PruneSafeRequest identifies a local checkout marked safe to remove.
type PruneSafeRequest struct {
	ProjectID      string `json:"project_id"`
	Branch         string `json:"branch"`
	WorktreePath   string `json:"worktree_path"`
	ClearIndexLock bool   `json:"clear_index_lock,omitempty"`
}

// PruneSafe re-checks forge prune hints, then removes the local checkout when still safe.
func (c *Commands) PruneSafe(ctx context.Context, doc config.File, req PruneSafeRequest) error {
	s, err := c.requireLocal()
	if err != nil {
		return err
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	worktreePath := strings.TrimSpace(req.WorktreePath)
	if projectID == "" || branch == "" || worktreePath == "" {
		return badRequest("project_id, branch, and worktree_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return badRequestCause("", err)
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return badRequest("unknown project")
	}

	abs, err := localgit.ExpandPath(worktreePath)
	if err != nil {
		return badRequestCause(fmt.Sprintf("worktree path: %v", err), err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	abs = canonicalCheckoutPath(ctx, s.Local, abs)
	canon := func(p string) string { return canonicalCheckoutPath(ctx, s.Local, p) }

	opts := remotegit.SummaryOpts{
		Fresh:     false,
		Cache:     s.Cache,
		HeadsTTL:  time.Duration(doc.EffectiveHeadsSeconds()) * time.Second,
		MergedTTL: time.Duration(doc.EffectiveMergedSeconds()) * time.Second,
	}
	row := s.summarize(ctx, p, opts)
	if !row.RemoteNamesOK {
		return badRequest("cannot re-validate remote heads; refuse prune")
	}
	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	forgeBranches := filterHiddenBranches(row.Branches, doc.UI.HideBranches)
	row.Local = s.attachLocal(ctx, p, disc, projectLabelsByKey(doc.Projects), false, time.Duration(doc.EffectiveFetchSeconds())*time.Second, forgeBranchNames(forgeBranches))
	row.Branches = forgeBranches
	if row.Local == nil || !row.Local.Mapped {
		return badRequest("project has no mapped local checkout")
	}
	if !repoPathAllowedCanon(row.Local, abs, canon) {
		return badRequest("worktree_path is not a mapped checkout for this project")
	}
	s.annotateContentOnDefault(ctx, &row)
	s.confirmMergedForCandidates(ctx, p, &row, opts)
	remotegit.EnrichPruneHints(&row)

	wt, ok := findSafeWorktree(row.Local, branch, abs, canon)
	if !ok {
		return badRequest(fmt.Sprintf("branch %q is not safe to remove (re-check prune hints)", branch))
	}
	defaultBranch := strings.TrimSpace(row.Local.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	if err := c.ensureWritableIndex(ctx, wt.Path, req.ClearIndexLock); err != nil {
		return err
	}
	err = s.Local.RemoveSafeCheckout(ctx, wt.Path, branch, defaultBranch)
	if err != nil && localgit.IsIndexLockError(err) {
		if lockErr := c.ensureWritableIndex(ctx, wt.Path, req.ClearIndexLock); lockErr != nil {
			return lockErr
		}
		if req.ClearIndexLock {
			err = s.Local.RemoveSafeCheckout(ctx, wt.Path, branch, defaultBranch)
		}
	}
	if err != nil {
		if errors.Is(err, localgit.ErrInvalidBranch) ||
			errors.Is(err, localgit.ErrDirtyTree) ||
			errors.Is(err, localgit.ErrDiverged) ||
			errors.Is(err, localgit.ErrMissingBranch) {
			return badRequestCause("", err)
		}
		return fmt.Errorf("remove checkout: %w", err)
	}
	return nil
}

func findSafeWorktree(local *board.LocalStatus, branch, absPath string, canon func(string) string) (board.LocalWorktree, bool) {
	if local == nil {
		return board.LocalWorktree{}, false
	}
	branch = strings.TrimSpace(branch)
	for _, wt := range local.Worktrees {
		if strings.TrimSpace(wt.Branch) != branch {
			continue
		}
		if !sameCheckoutPath(wt.Path, absPath, canon) && !sameCheckoutPath(wt.AppearancePath, absPath, canon) {
			continue
		}
		if wt.PruneHint != board.PruneSafe {
			continue
		}
		wt.Path = absPath
		return wt, true
	}
	return board.LocalWorktree{}, false
}

func sameCheckoutPath(a, b string, canon func(string) string) bool {
	if canon != nil {
		a = canon(a)
		b = canon(b)
	}
	a = cleanRepoPath(a)
	b = cleanRepoPath(b)
	return a != "" && a == b
}

func cleanRepoPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
}

func canonicalCheckoutPath(ctx context.Context, local LocalGit, path string) string {
	path = cleanRepoPath(path)
	if path == "" || local == nil {
		return path
	}
	if c, ok := local.(interface {
		CanonicalCheckoutPath(context.Context, string) string
	}); ok {
		return c.CanonicalCheckoutPath(ctx, path)
	}
	return path
}
