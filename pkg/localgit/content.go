package localgit

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ContentOnDefault reports whether branch tip has no unique tree content versus default.
// Prefers refs/remotes/origin/<default> when that ref exists; otherwise refs/heads/<default>.
// True when branch is an ancestor of the baseline, or tip trees are identical (squash case).
// Fail closed: git command failures yield ok=false with a reason; only validation returns err.
func (in *Inspector) ContentOnDefault(ctx context.Context, repoPath, branch, defaultBranch string) (ok bool, reason string, err error) {
	if in == nil {
		return false, "", fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	defaultBranch = strings.TrimSpace(defaultBranch)
	if branch == "" {
		return false, "", fmt.Errorf("branch is required")
	}
	if err := ValidateBranchName(branch); err != nil {
		return false, "", err
	}
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	if err := ValidateBranchName(defaultBranch); err != nil {
		return false, "", fmt.Errorf("default branch: %w", err)
	}
	if branch == defaultBranch {
		return false, "branch is default", nil
	}

	abs, err := ExpandPath(repoPath)
	if err != nil {
		return false, "", fmt.Errorf("repo path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	baseline, baselineReason, okBaseline := in.resolveContentBaseline(ctx, abs, defaultBranch)
	if !okBaseline {
		return false, baselineReason, nil
	}

	if _, err := in.git(ctx, abs, "merge-base", baseline, branch); err != nil {
		return false, "unrelated histories", nil
	}

	if _, err := in.git(ctx, abs, "merge-base", "--is-ancestor", branch, baseline); err == nil {
		return true, "ancestor of " + baseline, nil
	} else if code, hasCode := commandExitCode(err); hasCode && code == 1 {
		// Not an ancestor; fall through to tip-tree comparison.
	} else {
		return false, "ancestor check: " + err.Error(), nil
	}

	if _, err := in.git(ctx, abs, "diff", "--quiet", baseline, branch); err == nil {
		return true, "identical trees vs " + baseline, nil
	} else if code, hasCode := commandExitCode(err); hasCode && code == 1 {
		return false, "trees differ from " + baseline, nil
	}
	return false, "diff: " + err.Error(), nil
}

func (in *Inspector) resolveContentBaseline(ctx context.Context, abs, defaultBranch string) (ref, reason string, ok bool) {
	originRef := "refs/remotes/origin/" + defaultBranch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", originRef); err == nil {
		return originRef, "", true
	}
	localRef := "refs/heads/" + defaultBranch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", localRef); err == nil {
		return localRef, "", true
	}
	return "", "missing baseline " + defaultBranch, false
}

type exitCoder interface {
	ExitCode() int
}

func commandExitCode(err error) (int, bool) {
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode(), true
	}
	return 0, false
}
