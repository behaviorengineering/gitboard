// Package localgit is public API: local checkout discovery and status through
// the git CLI. Inspector only reads. Every write (fetch, pull, safe prune,
// index.lock clear) goes through a Mutator, which only a MutationLease from
// MutationCoordinator hands out, so writers to one repository are serialised.
package localgit

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
)

// Worktree is one checkout linked to a repository (including the main tree).
type Worktree struct {
	Path     string `json:"path"`
	Branch   string `json:"branch,omitempty"`
	Tag      string `json:"tag,omitempty"` // Exact tag when Detached and HEAD is tagged.
	Detached bool   `json:"detached,omitempty"`
	Bare     bool   `json:"bare,omitempty"`
	Locked   bool   `json:"locked,omitempty"`
	Prunable bool   `json:"prunable,omitempty"`
	Main     bool   `json:"main,omitempty"`
	Dirty    bool   `json:"dirty,omitempty"`
	Ahead    int    `json:"ahead,omitempty"`
	Behind   int    `json:"behind,omitempty"`
	Upstream string `json:"upstream,omitempty"`
}

// Status is local checkout health for one forge project.
type Status struct {
	Mapped        bool         `json:"mapped"`
	Path          string       `json:"path,omitempty"`
	Error         string       `json:"error,omitempty"`
	Branch        string       `json:"branch,omitempty"`
	Tag           string       `json:"tag,omitempty"` // Exact tag when Detached and HEAD is tagged.
	Detached      bool         `json:"detached,omitempty"`
	Dirty         bool         `json:"dirty,omitempty"`
	Ahead         int          `json:"ahead,omitempty"`
	Behind        int          `json:"behind,omitempty"`
	Upstream      string       `json:"upstream,omitempty"`
	DefaultBranch string       `json:"default_branch,omitempty"`
	DefaultBehind int          `json:"default_behind,omitempty"`
	DefaultAhead  int          `json:"default_ahead,omitempty"`
	OriginSync    []BranchSync `json:"origin_sync,omitempty"`
	Worktrees     []Worktree   `json:"worktrees,omitempty"`
}

// BranchSync is local refs/heads/<name> versus refs/remotes/origin/<name>.
type BranchSync struct {
	Name   string `json:"name"`
	Ahead  int    `json:"ahead,omitempty"`  // local has commits origin lacks.
	Behind int    `json:"behind,omitempty"` // origin has commits local lacks.
}

// Inspector reads local git state via the git CLI. It never writes; see Mutator.
type Inspector struct {
	Run cliexec.Exec

	version *versionProbe // nil: treat git as older than 2.41.
}

// NewInspector returns an Inspector that runs git on run and probes the git
// version once, on first use. Panics if run is nil.
func NewInspector(run cliexec.Exec) *Inspector {
	if run == nil {
		panic("localgit.NewInspector: Exec is required")
	}
	return &Inspector{Run: run, version: &versionProbe{}}
}

// git is the one place an Inspector runs git, so a nil receiver or a literal
// Inspector{} without Run returns ErrInspectorMissing instead of panicking.
func (in *Inspector) git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	if in == nil || in.Run == nil {
		return nil, ErrInspectorMissing
	}
	argv := append([]string{"-C", dir}, args...)
	return in.Run.Run(ctx, "git", argv...)
}

// InspectPath reports status for a checkout and all of its worktrees.
func (in *Inspector) InspectPath(ctx context.Context, path string) Status {
	abs, err := ExpandPath(path)
	if err != nil {
		return Status{Mapped: true, Error: err.Error()}
	}
	if _, err := os.Stat(abs); err != nil {
		return Status{Mapped: true, Path: abs, Error: fmt.Sprintf("path missing: %v", err)}
	}
	if !in.isGitDir(ctx, abs) {
		return Status{Mapped: true, Path: abs, Error: "not a git repository"}
	}

	trees, err := in.listWorktrees(ctx, abs)
	if err != nil {
		// Fall back to single-tree inspect.
		wt, wtErr := in.inspectWorktree(ctx, abs, true)
		if wtErr != nil {
			return Status{Mapped: true, Path: abs, Error: errors.Join(err, wtErr).Error()}
		}
		return summarize(abs, []Worktree{wt})
	}
	for i := range trees {
		if trees[i].Bare {
			continue
		}
		trees[i].Path = in.CanonicalCheckoutPath(ctx, trees[i].Path)
		detail, detailErr := in.inspectWorktree(ctx, trees[i].Path, trees[i].Main)
		if detailErr != nil {
			// Fail closed for prune: never leave Dirty as a false zero-value.
			trees[i].Dirty = true
			if trees[i].Branch == "" {
				trees[i].Branch = "unknown"
			}
			continue
		}
		detail.Main = trees[i].Main
		detail.Bare = trees[i].Bare
		detail.Locked = trees[i].Locked
		detail.Prunable = trees[i].Prunable
		if trees[i].Branch != "" && detail.Branch == "" {
			detail.Branch = trees[i].Branch
		}
		trees[i] = detail
	}
	return summarize(abs, trees)
}

func summarize(path string, trees []Worktree) Status {
	st := Status{Mapped: true, Path: path, Worktrees: trees}
	primary := pickPrimary(trees)
	if primary == nil {
		return st
	}
	st.Path = primary.Path
	st.Branch = primary.Branch
	st.Tag = primary.Tag
	st.Detached = primary.Detached
	st.Dirty = primary.Dirty
	st.Ahead = primary.Ahead
	st.Behind = primary.Behind
	st.Upstream = primary.Upstream
	return st
}

func pickPrimary(trees []Worktree) *Worktree {
	for i := range trees {
		if trees[i].Main && !trees[i].Bare {
			return &trees[i]
		}
	}
	for i := range trees {
		if !trees[i].Bare {
			return &trees[i]
		}
	}
	return nil
}

func (in *Inspector) isGitDir(ctx context.Context, dir string) bool {
	return in.requireGitDir(ctx, dir) == nil
}

// requireGitDir returns an error that keeps the git cause (timeout, missing
// binary, permission) when dir is not usable as a git repository.
func (in *Inspector) requireGitDir(ctx context.Context, dir string) error {
	if _, err := in.git(ctx, dir, "rev-parse", "--git-dir"); err != nil {
		return fmt.Errorf("not a git repository %s: %w", dir, err)
	}
	return nil
}

func (in *Inspector) listWorktrees(ctx context.Context, dir string) ([]Worktree, error) {
	out, err := in.git(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeList(out)
}

// parseWorktreeList reads `git worktree list --porcelain`. The first entry is
// the main worktree.
func parseWorktreeList(out []byte) ([]Worktree, error) {
	var trees []Worktree
	sc := bufio.NewScanner(bytes.NewReader(out))
	open := false // a "worktree" line started the current block.
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			open = false
			continue
		}
		key, val, _ := strings.Cut(line, " ")
		if key == "worktree" {
			trees = append(trees, Worktree{Path: val, Main: len(trees) == 0})
			open = true
			continue
		}
		if open {
			applyWorktreeAttr(&trees[len(trees)-1], key, val)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(trees) == 0 {
		return nil, fmt.Errorf("no worktrees listed")
	}
	return trees, nil
}

// applyWorktreeAttr sets one porcelain attribute line on wt.
func applyWorktreeAttr(wt *Worktree, key, val string) {
	switch key {
	case "bare":
		wt.Bare = true
	case "detached":
		wt.Detached = true
	case "branch":
		wt.Branch = strings.TrimPrefix(val, headsPrefix)
	case "locked":
		wt.Locked = true
	case "prunable":
		wt.Prunable = true
	}
}

func looksLikeGitDir(path string) bool {
	slash := filepath.ToSlash(path)
	if strings.Contains(slash, "/.git/modules/") {
		return true
	}
	return filepath.Base(slash) == gitDirName || strings.HasSuffix(slash, "/"+gitDirName)
}

// CanonicalCheckoutPath returns the working tree for path.
// Submodule `git worktree list` often reports the git dir under .git/modules;
// --show-toplevel follows core.worktree to the real checkout.
func (in *Inspector) CanonicalCheckoutPath(ctx context.Context, path string) string {
	abs, err := ResolvePath(path)
	if err != nil {
		return strings.TrimSpace(path)
	}
	if in == nil || !looksLikeGitDir(abs) {
		return abs
	}
	top, err := in.git(ctx, abs, "rev-parse", "--show-toplevel")
	if err != nil {
		return abs
	}
	out := strings.TrimSpace(string(top))
	if out == "" {
		return abs
	}
	if resolved, err := filepath.EvalSymlinks(out); err == nil {
		out = resolved
	}
	return filepath.Clean(out)
}

func (in *Inspector) inspectWorktree(ctx context.Context, dir string, isMain bool) (Worktree, error) {
	wt := Worktree{Path: dir, Main: isMain}

	branchOut, err := in.git(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return wt, err
	}
	branch := strings.TrimSpace(string(branchOut))
	if branch == "HEAD" {
		wt.Detached = true
		short, shortErr := in.git(ctx, dir, "rev-parse", "--short", "HEAD")
		if shortErr == nil {
			wt.Branch = strings.TrimSpace(string(short))
		}
		wt.Tag = in.exactTagAtHEAD(ctx, dir)
	} else {
		wt.Branch = branch
	}

	dirtyOut, err := in.git(ctx, dir, "status", "--porcelain")
	if err != nil {
		return wt, err
	}
	wt.Dirty = len(bytes.TrimSpace(dirtyOut)) > 0

	upOut, upErr := in.git(ctx, dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	if upErr == nil {
		wt.Upstream = strings.TrimSpace(string(upOut))
		ahead, behind, lrErr := in.leftRight(ctx, dir, "@{upstream}", "HEAD")
		if lrErr == nil {
			wt.Ahead = ahead
			wt.Behind = behind
		}
	}
	return wt, nil
}

// exactTagAtHEAD returns a tag name when HEAD points exactly at a tag.
// Prefer lightweight/version tags via describe; empty when HEAD is not tagged.
func (in *Inspector) exactTagAtHEAD(ctx context.Context, dir string) string {
	out, err := in.git(ctx, dir, "describe", "--exact-match", "--tags", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// leftRight counts commits only on right (ahead) and only on left (behind).
// `git rev-list --left-right --count left...right` prints "<left> <right>".
func (in *Inspector) leftRight(ctx context.Context, dir, left, right string) (ahead, behind int, err error) {
	out, err := in.git(ctx, dir, "rev-list", "--left-right", "--count", left+"..."+right)
	if err != nil {
		return 0, 0, fmt.Errorf("compare %s...%s: %w", left, right, err)
	}
	behind, ahead, err = parseLeftRightCount(string(out))
	if err != nil {
		return 0, 0, fmt.Errorf("compare %s...%s: %w", left, right, err)
	}
	return ahead, behind, nil
}

func (in *Inspector) listRefShortNames(ctx context.Context, dir, prefix string) ([]string, error) {
	out, err := in.git(ctx, dir, "for-each-ref", "--format=%(refname:short)", prefix)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	var names []string
	for sc.Scan() {
		n := strings.TrimSpace(sc.Text())
		if n == "" {
			continue
		}
		names = append(names, n)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

// OriginRemote returns the origin URL for a checkout.
func (in *Inspector) OriginRemote(ctx context.Context, dir string) (string, error) {
	out, err := in.git(ctx, dir, "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CommonGitDir resolves the shared git dir (useful for worktree identity).
func (in *Inspector) CommonGitDir(ctx context.Context, dir string) (string, error) {
	out, err := in.git(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(string(out))), nil
}
