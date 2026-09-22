package localgit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

// PullPhases records wall time per pull stage in milliseconds.
type PullPhases struct {
	FetchMs         int64 `json:"fetch_ms,omitempty"`
	SubmodulePreMs  int64 `json:"submodule_pre_ms,omitempty"`
	MergeMs         int64 `json:"merge_ms,omitempty"`
	SubmodulePostMs int64 `json:"submodule_post_ms,omitempty"`
}

// PullFFOnly fast-forwards local branch to match origin/<branch>.
//
// Fetches origin/<branch> first, then:
//   - checked out + clean: git merge --ff-only origin/<branch>
//   - not checked out: git fetch origin <branch>:<branch> (ff-only)
//
// When the branch is checked out, stale nested submodule checkouts are synced
// with `git submodule update --init --recursive` before refusing a dirty tree
// and again after a successful fast-forward, so pins like .cursor/packs/shared
// match the parent tip.
func (in *Inspector) PullFFOnly(ctx context.Context, repoPath, branch string) error {
	_, err := in.PullFFOnlyWithPhases(ctx, repoPath, branch)
	return err
}

// PullFFOnlyWithPhases is PullFFOnly with per-stage timings.
// The post-merge submodule update is skipped when the merge did not advance
// HEAD or when submodule pins are unchanged between pre and post merge tips.
func (in *Inspector) PullFFOnlyWithPhases(ctx context.Context, repoPath, branch string) (PullPhases, error) {
	var phases PullPhases
	if in == nil {
		return phases, ErrInspectorMissing
	}
	branch = strings.TrimSpace(branch)
	if err := ValidateBranchName(branch); err != nil {
		return phases, err
	}

	abs, err := ExpandPath(repoPath)
	if err != nil {
		return phases, fmt.Errorf("repo path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	if !in.isGitDir(ctx, abs) {
		return phases, fmt.Errorf("not a git repository: %s", abs)
	}

	localRef := "refs/heads/" + branch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", localRef); err != nil {
		return phases, fmt.Errorf("%w: %q", ErrMissingBranch, branch)
	}

	checkout, checkedOut := in.worktreeOnBranch(ctx, abs, branch)
	if checkedOut {
		detail, err := in.inspectWorktree(ctx, checkout, false)
		if err != nil {
			return phases, fmt.Errorf("inspect checkout: %w", err)
		}
		if detail.Dirty {
			// Submodule working trees often look dirty when they lag the parent tip.
			start := time.Now()
			_, syncErr := in.updateSubmodules(ctx, checkout)
			phases.SubmodulePreMs = time.Since(start).Milliseconds()
			if syncErr == nil {
				detail, err = in.inspectWorktree(ctx, checkout, false)
				if err != nil {
					return phases, fmt.Errorf("inspect checkout: %w", err)
				}
			}
		}
		if detail.Dirty {
			return phases, fmt.Errorf("%w at %s; commit or stash before pull", ErrDirtyTree, checkout)
		}
	}

	// Refresh origin/<branch> before comparing / merging.
	fetchDir := abs
	if checkedOut {
		fetchDir = checkout
	}
	start := time.Now()
	if _, err := in.git(ctx, fetchDir, "fetch", "origin", branch); err != nil {
		phases.FetchMs = time.Since(start).Milliseconds()
		return phases, fmt.Errorf("fetch origin %s: %w", branch, err)
	}
	phases.FetchMs = time.Since(start).Milliseconds()

	remoteRef := "refs/remotes/origin/" + branch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", remoteRef); err != nil {
		return phases, fmt.Errorf("missing %s after fetch", remoteRef)
	}
	ahead, behind, ok := in.leftRight(ctx, abs, remoteRef, localRef)
	if !ok {
		return phases, fmt.Errorf("could not compare %s to origin/%s", branch, branch)
	}
	if behind == 0 {
		return phases, fmt.Errorf("%w: %q", ErrUpToDate, branch)
	}
	if ahead > 0 {
		return phases, fmt.Errorf("%w: %q (%d ahead, %d behind)", ErrDiverged, branch, ahead, behind)
	}

	if checkedOut {
		preMergeSHA := in.revParseHead(ctx, checkout)
		start := time.Now()
		if _, err := in.git(ctx, checkout, "merge", "--ff-only", "origin/"+branch); err != nil {
			phases.MergeMs = time.Since(start).Milliseconds()
			return phases, fmt.Errorf("ff-only merge origin/%s: %w", branch, err)
		}
		phases.MergeMs = time.Since(start).Milliseconds()
		postMergeSHA := in.revParseHead(ctx, checkout)
		if preMergeSHA != "" && preMergeSHA == postMergeSHA {
			return phases, nil
		}
		if !in.submodulePinsChanged(ctx, checkout, preMergeSHA, postMergeSHA) {
			return phases, nil
		}
		start = time.Now()
		didWork, err := in.updateSubmodules(ctx, checkout)
		phases.SubmodulePostMs = time.Since(start).Milliseconds()
		_ = didWork
		if err != nil {
			return phases, fmt.Errorf("after fast-forward: %w", err)
		}
		return phases, nil
	}

	spec := branch + ":" + branch
	start = time.Now()
	if _, err := in.git(ctx, abs, "fetch", "origin", spec); err != nil {
		phases.FetchMs += time.Since(start).Milliseconds()
		return phases, fmt.Errorf("fetch origin %s: %w", spec, err)
	}
	phases.FetchMs += time.Since(start).Milliseconds()
	return phases, nil
}

// revParseHead returns HEAD SHA for dir, or empty on error.
func (in *Inspector) revParseHead(ctx context.Context, dir string) string {
	out, err := in.git(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// submodulePinsChanged reports whether submodule pins may have changed
// between oldSHA and newSHA. Fail closed: any doubt returns true.
func (in *Inspector) submodulePinsChanged(ctx context.Context, dir, oldSHA, newSHA string) bool {
	if oldSHA == "" || newSHA == "" || oldSHA == newSHA {
		return oldSHA != newSHA
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitmodules")); err != nil {
		return !os.IsNotExist(err)
	}
	if _, err := in.git(ctx, dir, "diff", "--quiet", oldSHA, newSHA, "--", ".gitmodules"); err != nil {
		return true
	}
	raw, err := in.git(ctx, dir, "diff", "--raw", oldSHA, newSHA, "--")
	if err != nil {
		return true
	}
	return bytes.Contains(raw, []byte("160000"))
}

// updateSubmodules checks nested submodules out to the commits recorded by dir's tip.
// Skips network work when the tree has no .gitmodules file.
// Reports whether it ran an update.
func (in *Inspector) updateSubmodules(ctx context.Context, dir string) (bool, error) {
	if _, err := os.Stat(filepath.Join(dir, ".gitmodules")); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		// Missing permission or other stat errors: attempt update (fail closed on real work).
	}
	status, err := in.git(ctx, dir, "submodule", "status", "--recursive")
	if err == nil && len(bytes.TrimSpace(status)) == 0 {
		return false, nil
	}
	if _, err := in.git(ctx, dir, "submodule", "update", "--init", "--recursive"); err != nil {
		return false, fmt.Errorf("submodule update --init --recursive: %w", err)
	}
	return true, nil
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
