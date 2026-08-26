package localgit

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// RemoveSafeCheckout drops a local branch checkout that the board marked safe to remove.
//
// Linked worktree: remove the worktree path, then delete the branch from the main tree.
// Main worktree on the branch: switch to defaultBranch, then delete the branch.
//
// Uses `git branch -D` because squash-merged branches are often not ancestors of default.
// Callers must re-check forge prune safety first. Branch names are validated; dirty trees refuse.
func (in *Inspector) RemoveSafeCheckout(ctx context.Context, worktreePath, branch, defaultBranch string) error {
	if in == nil {
		return fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	defaultBranch = strings.TrimSpace(defaultBranch)
	if branch == "" {
		return fmt.Errorf("branch is required")
	}
	if err := ValidateBranchName(branch); err != nil {
		return err
	}
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	if err := ValidateBranchName(defaultBranch); err != nil {
		return fmt.Errorf("default branch: %w", err)
	}
	if branch == defaultBranch {
		return fmt.Errorf("refusing to remove default branch %q", branch)
	}

	abs, err := ExpandPath(worktreePath)
	if err != nil {
		return fmt.Errorf("worktree path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	trees, err := in.listWorktrees(ctx, abs)
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	var target *Worktree
	var main *Worktree
	for i := range trees {
		wt := &trees[i]
		path := filepath.Clean(wt.Path)
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		wt.Path = path
		if wt.Main {
			main = wt
		}
		if path == abs {
			target = wt
		}
	}
	if target == nil {
		return fmt.Errorf("worktree %s not found", abs)
	}
	if target.Bare {
		return fmt.Errorf("refusing to remove bare worktree")
	}
	if target.Detached {
		return fmt.Errorf("refusing to remove detached HEAD checkout")
	}
	if target.Locked {
		return fmt.Errorf("refusing to remove locked worktree")
	}
	if target.Branch != branch {
		if target.Branch == "" {
			return fmt.Errorf("worktree branch unknown; expected %q", branch)
		}
		return fmt.Errorf("worktree is on %q, not %q", target.Branch, branch)
	}

	detail, err := in.inspectWorktree(ctx, abs, target.Main)
	if err != nil {
		return fmt.Errorf("re-check worktree: %w", err)
	}
	if detail.Dirty {
		return fmt.Errorf("%w at %s; commit or stash before remove", ErrDirtyTree, abs)
	}

	if target.Main {
		cur, err := in.git(ctx, abs, "branch", "--show-current")
		if err != nil {
			return fmt.Errorf("current branch: %w", err)
		}
		if got := strings.TrimSpace(string(cur)); got != "" && got != branch {
			return fmt.Errorf("checkout is on %q, not %q", got, branch)
		}
		if _, err := in.git(ctx, abs, "switch", defaultBranch); err != nil {
			return fmt.Errorf("switch to %s: %w", defaultBranch, err)
		}
		if _, err := in.git(ctx, abs, "branch", "-D", branch); err != nil {
			return fmt.Errorf("delete branch %s: %w", branch, err)
		}
		return nil
	}

	if main == nil {
		return fmt.Errorf("main worktree not found for %s", abs)
	}
	if _, err := in.git(ctx, main.Path, "worktree", "remove", abs); err != nil {
		return fmt.Errorf("worktree remove: %w", err)
	}
	if _, err := in.git(ctx, main.Path, "branch", "-D", branch); err != nil {
		return fmt.Errorf("worktree removed at %s; delete branch %s: %w", abs, branch, err)
	}
	return nil
}
