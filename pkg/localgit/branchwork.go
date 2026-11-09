package localgit

import (
	"bytes"
	"context"
	"fmt"
)

// IsDirty reports whether the worktree at dir has uncommitted changes.
func (in *Inspector) IsDirty(ctx context.Context, dir string) (bool, error) {
	out, err := in.git(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("status: %w", err)
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// SharesHistory reports whether base and branch have a common ancestor. git
// reports "no merge base" as a failure too, so any error means unrelated
// as far as the caller can tell; err carries the cause.
func (in *Inspector) SharesHistory(ctx context.Context, dir, base, branch string) (bool, error) {
	if _, err := in.git(ctx, dir, "merge-base", base, branch); err != nil {
		return false, fmt.Errorf("merge-base %s %s: %w", base, branch, err)
	}
	return true, nil
}

// AheadBehind counts commits only on branch (ahead) and only on base (behind).
func (in *Inspector) AheadBehind(ctx context.Context, dir, base, branch string) (ahead, behind int, err error) {
	return in.leftRight(ctx, dir, base, branch)
}

// UniqueCommits lists `git log --oneline base..branch`: commits on branch
// that base lacks, newest first.
func (in *Inspector) UniqueCommits(ctx context.Context, dir, base, branch string) ([]string, error) {
	out, err := in.git(ctx, dir, "log", "--oneline", base+".."+branch)
	if err != nil {
		return nil, fmt.Errorf("log %s..%s: %w", base, branch, err)
	}
	return nonEmptyLines(string(out)), nil
}

// ChangedFiles lists files branch changed since it left base
// (`git diff --name-only base...branch`).
func (in *Inspector) ChangedFiles(ctx context.Context, dir, base, branch string) ([]string, error) {
	out, err := in.git(ctx, dir, "diff", "--name-only", base+"..."+branch)
	if err != nil {
		return nil, fmt.Errorf("diff %s...%s: %w", base, branch, err)
	}
	return nonEmptyLines(string(out)), nil
}
