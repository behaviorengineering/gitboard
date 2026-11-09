package localgit

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// RemoveSafeCheckout drops a local branch checkout that the board marked safe to remove.
//
// Linked worktree: remove the worktree path, then delete the branch from the main tree.
// Main worktree on the branch: bring defaultBranch in line with origin (ff-only when
// possible; if default is not checked out and has diverged, move the local tip to
// origin/<default>), switch to defaultBranch, then delete the branch.
//
// Uses `git branch -D` because squash-merged branches are often not ancestors of default.
// Callers must re-check forge prune safety first. Branch names are validated; dirty trees refuse.
// worktreePath must resolve to the lease's common git directory.
func (m *Mutator) RemoveSafeCheckout(ctx context.Context, worktreePath, branch, defaultBranch string) error {
	if _, _, err := m.bind(ctx, worktreePath); err != nil {
		return err
	}
	branch, defaultBranch, err := normalisePruneArgs(branch, defaultBranch)
	if err != nil {
		return err
	}
	abs, err := ResolvePath(worktreePath)
	if err != nil {
		return fmt.Errorf("worktree path: %w", err)
	}
	abs = m.in.CanonicalCheckoutPath(ctx, abs)

	target, main, err := m.in.locateWorktree(ctx, abs)
	if err != nil {
		return err
	}
	if err := checkRemovable(target, branch); err != nil {
		return err
	}
	detail, err := m.in.inspectWorktree(ctx, abs, target.Main)
	if err != nil {
		return fmt.Errorf("re-check worktree: %w", err)
	}
	if detail.Dirty {
		return fmt.Errorf("%w at %s; commit or stash before remove", ErrDirtyTree, abs)
	}
	if target.Main {
		return m.removeMainCheckout(ctx, abs, branch, defaultBranch)
	}
	if main == nil {
		return fmt.Errorf("main worktree not found for %s", abs)
	}
	return m.removeLinkedWorktree(ctx, main.Path, abs, branch)
}

// normalisePruneArgs trims and validates the branch to remove and the
// default branch to land on, which falls back to main and must differ.
func normalisePruneArgs(branch, defaultBranch string) (string, string, error) {
	branch = strings.TrimSpace(branch)
	defaultBranch = strings.TrimSpace(defaultBranch)
	if branch == "" {
		return "", "", fmt.Errorf("branch is required")
	}
	if err := ValidateBranchName(branch); err != nil {
		return "", "", err
	}
	if defaultBranch == "" {
		defaultBranch = defaultBranchFallback
	}
	if err := ValidateBranchName(defaultBranch); err != nil {
		return "", "", fmt.Errorf("default branch: %w", err)
	}
	if branch == defaultBranch {
		return "", "", fmt.Errorf("refusing to remove default branch %q", branch)
	}
	return branch, defaultBranch, nil
}

// locateWorktree finds the worktree at abs and the repository's main
// worktree, comparing canonical paths.
func (in *Inspector) locateWorktree(ctx context.Context, abs string) (target, main *Worktree, err error) {
	trees, err := in.listWorktrees(ctx, abs)
	if err != nil {
		return nil, nil, fmt.Errorf("list worktrees: %w", err)
	}
	for i := range trees {
		wt := &trees[i]
		wt.Path = in.CanonicalCheckoutPath(ctx, wt.Path)
		if wt.Main {
			main = wt
		}
		if wt.Path == abs {
			target = wt
		}
	}
	if target == nil {
		return nil, nil, fmt.Errorf("worktree %s not found", abs)
	}
	return target, main, nil
}

// checkRemovable refuses bare, detached and locked worktrees, and a
// worktree that is not on branch.
func checkRemovable(target *Worktree, branch string) error {
	switch {
	case target.Bare:
		return fmt.Errorf("refusing to remove bare worktree")
	case target.Detached:
		return fmt.Errorf("refusing to remove detached HEAD checkout")
	case target.Locked:
		return fmt.Errorf("refusing to remove locked worktree")
	case target.Branch == "":
		return fmt.Errorf("worktree branch unknown; expected %q", branch)
	case target.Branch != branch:
		return fmt.Errorf("worktree is on %q, not %q", target.Branch, branch)
	}
	return nil
}

// removeMainCheckout lands the main worktree on an up-to-date default
// branch, then deletes branch.
func (m *Mutator) removeMainCheckout(ctx context.Context, abs, branch, defaultBranch string) error {
	in := m.in
	cur, err := in.git(ctx, abs, "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("current branch: %w", err)
	}
	if got := strings.TrimSpace(string(cur)); got != "" && got != branch {
		return fmt.Errorf("checkout is on %q, not %q", got, branch)
	}
	// Land on an up-to-date default instead of a stale tip after switch.
	if err := m.ensureDefaultReadyForSafePrune(ctx, abs, defaultBranch); err != nil {
		return fmt.Errorf("update default branch %s before remove: %w", defaultBranch, err)
	}
	if _, err := in.git(ctx, abs, "switch", defaultBranch); err != nil {
		return fmt.Errorf("switch to %s: %w", defaultBranch, err)
	}
	if _, err := in.git(ctx, abs, "branch", "-D", branch); err != nil {
		return fmt.Errorf("delete branch %s: %w", branch, err)
	}
	return nil
}

// removeLinkedWorktree removes the linked worktree at abs, then deletes
// branch from the main worktree.
func (m *Mutator) removeLinkedWorktree(ctx context.Context, mainPath, abs, branch string) error {
	if _, err := m.in.git(ctx, mainPath, "worktree", "remove", abs); err != nil {
		return fmt.Errorf("worktree remove: %w", err)
	}
	if _, err := m.in.git(ctx, mainPath, "branch", "-D", branch); err != nil {
		return fmt.Errorf("worktree removed at %s; delete branch %s: %w", abs, branch, err)
	}
	return nil
}

// ensureDefaultReadyForSafePrune brings defaultBranch to origin before a main-worktree prune.
// Prefers ff-only. If the branch is not checked out and has diverged, moves the local tip
// to origin/<branch> so prune can land cleanly (discards local-only commits on that tip).
func (m *Mutator) ensureDefaultReadyForSafePrune(ctx context.Context, repoPath, branch string) error {
	err := m.ensureBranchFFFromOrigin(ctx, repoPath, branch)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrDiverged) {
		return err
	}
	if _, checkedOut := m.in.worktreeOnBranch(ctx, repoPath, branch); checkedOut {
		return err
	}
	return m.resetUnusedBranchToOrigin(ctx, repoPath, branch)
}

// resetUnusedBranchToOrigin force-updates a non-checked-out local branch to origin/<branch>.
func (m *Mutator) resetUnusedBranchToOrigin(ctx context.Context, repoPath, branch string) error {
	in := m.in
	branch = strings.TrimSpace(branch)
	if err := ValidateBranchName(branch); err != nil {
		return err
	}
	if _, checkedOut := in.worktreeOnBranch(ctx, repoPath, branch); checkedOut {
		return fmt.Errorf("%w: %q is checked out; cannot reset for safe prune", ErrDiverged, branch)
	}
	if _, err := in.git(ctx, repoPath, "fetch", "origin", branch); err != nil {
		return fmt.Errorf("fetch origin %s: %w", branch, err)
	}
	remoteRef := "origin/" + branch
	if _, err := in.git(ctx, repoPath, "rev-parse", "--verify", remoteRef); err != nil {
		return fmt.Errorf("missing %s after fetch: %w", remoteRef, err)
	}
	if _, err := in.git(ctx, repoPath, "branch", "-f", branch, remoteRef); err != nil {
		return fmt.Errorf("reset local %s to %s: %w", branch, remoteRef, err)
	}
	return nil
}
