package localgit

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Sentinel errors for pull / fast-forward operations.
var (
	ErrInvalidBranch = errors.New("invalid branch name")
	ErrMissingBranch = errors.New("missing local branch")
	ErrDirtyTree     = errors.New("working tree dirty")
	ErrDiverged      = errors.New("local branch diverged from origin")
	ErrUpToDate      = errors.New("local branch already up to date with origin")
)

// branchNameOK allows typical git branch names used in refspecs (no :, leading -, or whitespace).
var branchNameOK = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// ValidateBranchName rejects names unsafe to pass as git ref / refspec components.
func ValidateBranchName(branch string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return fmt.Errorf("%w: empty", ErrInvalidBranch)
	}
	if strings.EqualFold(branch, "HEAD") {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if strings.HasPrefix(branch, "-") || strings.HasPrefix(branch, ".") {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if strings.Contains(branch, "..") || strings.Contains(branch, "@{") {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if !branchNameOK.MatchString(branch) {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if len(branch) > 255 {
		return fmt.Errorf("%w: too long", ErrInvalidBranch)
	}
	return nil
}

// PullFFOnly fast-forwards local branch to match origin/<branch>.
//
// Fetches origin/<branch> first, then:
//   - checked out + clean: git merge --ff-only origin/<branch>
//   - not checked out: git fetch origin <branch>:<branch> (ff-only)
func (in *Inspector) PullFFOnly(ctx context.Context, repoPath, branch string) error {
	if in == nil {
		return fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	if err := ValidateBranchName(branch); err != nil {
		return err
	}

	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("repo path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	if !in.isGitDir(ctx, abs) {
		return fmt.Errorf("not a git repository: %s", abs)
	}

	localRef := "refs/heads/" + branch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", localRef); err != nil {
		return fmt.Errorf("%w: %q", ErrMissingBranch, branch)
	}

	checkout, checkedOut := in.worktreeOnBranch(ctx, abs, branch)
	if checkedOut {
		detail, err := in.inspectWorktree(ctx, checkout, false)
		if err != nil {
			return fmt.Errorf("inspect checkout: %w", err)
		}
		if detail.Dirty {
			return fmt.Errorf("%w at %s; commit or stash before pull", ErrDirtyTree, checkout)
		}
	}

	// Refresh origin/<branch> before comparing / merging.
	fetchDir := abs
	if checkedOut {
		fetchDir = checkout
	}
	if _, err := in.git(ctx, fetchDir, "fetch", "origin", branch); err != nil {
		return fmt.Errorf("fetch origin %s: %w", branch, err)
	}

	remoteRef := "refs/remotes/origin/" + branch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", remoteRef); err != nil {
		return fmt.Errorf("missing %s after fetch", remoteRef)
	}
	ahead, behind, ok := in.leftRight(ctx, abs, remoteRef, localRef)
	if !ok {
		return fmt.Errorf("could not compare %s to origin/%s", branch, branch)
	}
	if behind == 0 {
		return fmt.Errorf("%w: %q", ErrUpToDate, branch)
	}
	if ahead > 0 {
		return fmt.Errorf("%w: %q (%d ahead, %d behind)", ErrDiverged, branch, ahead, behind)
	}

	if checkedOut {
		if _, err := in.git(ctx, checkout, "merge", "--ff-only", "origin/"+branch); err != nil {
			return fmt.Errorf("ff-only merge origin/%s: %w", branch, err)
		}
		return nil
	}

	spec := branch + ":" + branch
	if _, err := in.git(ctx, abs, "fetch", "origin", spec); err != nil {
		return fmt.Errorf("fetch origin %s: %w", spec, err)
	}
	return nil
}

// ensureBranchFFFromOrigin fast-forwards branch from origin when behind-only.
// Already up to date is success. Diverged or dirty trees fail.
func (in *Inspector) ensureBranchFFFromOrigin(ctx context.Context, repoPath, branch string) error {
	err := in.PullFFOnly(ctx, repoPath, branch)
	if err == nil || errors.Is(err, ErrUpToDate) {
		return nil
	}
	return err
}

func (in *Inspector) worktreeOnBranch(ctx context.Context, dir, branch string) (path string, ok bool) {
	trees, err := in.listWorktrees(ctx, dir)
	if err != nil {
		return "", false
	}
	for _, wt := range trees {
		if wt.Bare || wt.Detached {
			continue
		}
		if wt.Branch == branch {
			path := filepath.Clean(wt.Path)
			if resolved, err := filepath.EvalSymlinks(path); err == nil {
				path = resolved
			}
			return path, true
		}
	}
	return "", false
}
