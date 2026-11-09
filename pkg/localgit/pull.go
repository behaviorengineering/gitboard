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

// PullFFOnly fast-forwards local branch to match origin/<branch> and reports
// per-stage timings. repoPath must resolve to the lease's common git directory.
//
// Fetches origin/<branch> first, then:
//   - checked out + clean: git merge --ff-only origin/<branch>
//   - not checked out: git fetch origin <branch>:<branch> (ff-only)
//
// When the branch is checked out, stale nested submodule checkouts are synced
// with `git submodule update --init --recursive` before refusing a dirty tree
// and again after a successful fast-forward, so nested submodule pins
// match the parent tip. The post-merge submodule update is skipped when the
// merge did not advance HEAD or when submodule pins are unchanged between pre
// and post merge tips.
func (m *Mutator) PullFFOnly(ctx context.Context, repoPath, branch string) (PullPhases, error) {
	if err := ValidateBranchName(branch); err != nil {
		return PullPhases{}, err
	}
	if _, _, err := m.bind(ctx, repoPath); err != nil {
		return PullPhases{}, err
	}
	return m.pullFFOnly(ctx, repoPath, branch)
}

// Add sums o into p, for a retried pull that ran some stages twice.
func (p *PullPhases) Add(o PullPhases) {
	p.FetchMs += o.FetchMs
	p.SubmodulePreMs += o.SubmodulePreMs
	p.MergeMs += o.MergeMs
	p.SubmodulePostMs += o.SubmodulePostMs
}

// PhaseTiming is one pull stage that ran and its wall time.
type PhaseTiming struct {
	// Name is the stage: fetch, submodule_pre, merge or submodule_post.
	Name     string
	Duration time.Duration
}

// Timings lists the stages that took time, in the order a pull runs them.
func (p PullPhases) Timings() []PhaseTiming {
	all := []PhaseTiming{
		{"fetch", time.Duration(p.FetchMs) * time.Millisecond},
		{"submodule_pre", time.Duration(p.SubmodulePreMs) * time.Millisecond},
		{"merge", time.Duration(p.MergeMs) * time.Millisecond},
		{"submodule_post", time.Duration(p.SubmodulePostMs) * time.Millisecond},
	}
	out := all[:0]
	for _, t := range all {
		if t.Duration > 0 {
			out = append(out, t)
		}
	}
	return out
}

// timed runs fn and adds its wall time to *dst, on success or failure.
func timed(dst *int64, fn func() error) error {
	start := time.Now()
	err := fn()
	*dst += time.Since(start).Milliseconds()
	return err
}

// pullFFOnly is PullFFOnly on a path already bound to the lease.
func (m *Mutator) pullFFOnly(ctx context.Context, repoPath, branch string) (PullPhases, error) {
	var phases PullPhases
	in := m.in
	branch = strings.TrimSpace(branch)
	if err := ValidateBranchName(branch); err != nil {
		return phases, err
	}
	abs, err := ResolvePath(repoPath)
	if err != nil {
		return phases, fmt.Errorf("repo path: %w", err)
	}
	if err := in.requireGitDir(ctx, abs); err != nil {
		return phases, err
	}
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", headsPrefix+branch); err != nil {
		return phases, fmt.Errorf("%w: %q: %w", ErrMissingBranch, branch, err)
	}

	checkout, checkedOut := in.worktreeOnBranch(ctx, abs, branch)
	// Refresh origin/<branch> before comparing or merging.
	fetchDir := abs
	if checkedOut {
		if err := m.prepareCheckedOut(ctx, checkout, &phases); err != nil {
			return phases, err
		}
		fetchDir = checkout
	}
	err = timed(&phases.FetchMs, func() error {
		_, err := in.git(ctx, fetchDir, "fetch", "origin", branch)
		return err
	})
	if err != nil {
		return phases, fmt.Errorf("fetch origin %s: %w", branch, err)
	}
	if err := in.compareToOrigin(ctx, abs, branch); err != nil {
		return phases, err
	}
	if checkedOut {
		return phases, m.ffCheckedOut(ctx, checkout, branch, &phases)
	}
	return phases, m.ffRef(ctx, abs, branch, &phases)
}

// prepareCheckedOut refuses a dirty checkout. Submodule working trees often
// look dirty when they lag the parent tip, so it syncs submodules once and
// checks again before refusing.
func (m *Mutator) prepareCheckedOut(ctx context.Context, checkout string, phases *PullPhases) error {
	detail, err := m.in.inspectWorktree(ctx, checkout, false)
	if err != nil {
		return fmt.Errorf("inspect checkout: %w", err)
	}
	if !detail.Dirty {
		return nil
	}
	syncErr := timed(&phases.SubmodulePreMs, func() error {
		_, err := m.updateSubmodules(ctx, checkout)
		return err
	})
	if syncErr == nil {
		detail, err = m.in.inspectWorktree(ctx, checkout, false)
		if err != nil {
			return fmt.Errorf("inspect checkout: %w", err)
		}
	}
	if detail.Dirty {
		return fmt.Errorf("%w at %s; commit or stash before pull", ErrDirtyTree, checkout)
	}
	return nil
}

// compareToOrigin allows a fast-forward only when branch is strictly behind
// origin/<branch>: up to date is ErrUpToDate, ahead is ErrDiverged.
func (in *Inspector) compareToOrigin(ctx context.Context, abs, branch string) error {
	remoteRef := originPrefix + branch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", remoteRef); err != nil {
		return fmt.Errorf("missing %s after fetch: %w", remoteRef, err)
	}
	ahead, behind, err := in.leftRight(ctx, abs, remoteRef, headsPrefix+branch)
	if err != nil {
		return fmt.Errorf("could not compare %s to origin/%s: %w", branch, branch, err)
	}
	if behind == 0 {
		return fmt.Errorf("%w: %q", ErrUpToDate, branch)
	}
	if ahead > 0 {
		return fmt.Errorf("%w: %q (%d ahead, %d behind)", ErrDiverged, branch, ahead, behind)
	}
	return nil
}

// ffCheckedOut merges origin/<branch> --ff-only in checkout, then updates
// submodules when the merge moved HEAD and changed submodule pins.
func (m *Mutator) ffCheckedOut(ctx context.Context, checkout, branch string, phases *PullPhases) error {
	in := m.in
	preMergeSHA := in.revParseHead(ctx, checkout)
	err := timed(&phases.MergeMs, func() error {
		_, err := in.git(ctx, checkout, "merge", "--ff-only", "origin/"+branch)
		return err
	})
	if err != nil {
		return fmt.Errorf("ff-only merge origin/%s: %w", branch, err)
	}
	postMergeSHA := in.revParseHead(ctx, checkout)
	if preMergeSHA != "" && preMergeSHA == postMergeSHA {
		return nil
	}
	if !in.submodulePinsChanged(ctx, checkout, preMergeSHA, postMergeSHA) {
		return nil
	}
	err = timed(&phases.SubmodulePostMs, func() error {
		_, err := m.updateSubmodules(ctx, checkout)
		return err
	})
	if err != nil {
		return fmt.Errorf("after fast-forward: %w", err)
	}
	return nil
}

// ffRef fast-forwards a branch that is not checked out with
// `git fetch origin <branch>:<branch>`, which git refuses unless ff-only.
func (m *Mutator) ffRef(ctx context.Context, abs, branch string, phases *PullPhases) error {
	spec := branch + ":" + branch
	err := timed(&phases.FetchMs, func() error {
		_, err := m.in.git(ctx, abs, "fetch", "origin", spec)
		return err
	})
	if err != nil {
		return fmt.Errorf("fetch origin %s: %w", spec, err)
	}
	return nil
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
func (m *Mutator) updateSubmodules(ctx context.Context, dir string) (bool, error) {
	if _, err := os.Stat(filepath.Join(dir, ".gitmodules")); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		// Missing permission or other stat errors: attempt update (fail closed on real work).
	}
	status, err := m.in.git(ctx, dir, "submodule", "status", "--recursive")
	if err == nil && len(bytes.TrimSpace(status)) == 0 {
		return false, nil
	}
	if _, err := m.in.git(ctx, dir, "submodule", "update", "--init", "--recursive"); err != nil {
		return false, fmt.Errorf("submodule update --init --recursive: %w", err)
	}
	return true, nil
}

// ensureBranchFFFromOrigin fast-forwards branch from origin when behind-only.
// Already up to date is success. Diverged or dirty trees fail.
func (m *Mutator) ensureBranchFFFromOrigin(ctx context.Context, repoPath, branch string) error {
	_, err := m.pullFFOnly(ctx, repoPath, branch)
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
