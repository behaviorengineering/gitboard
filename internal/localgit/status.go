package localgit

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
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
	Ahead  int    `json:"ahead,omitempty"`  // local has commits origin lacks
	Behind int    `json:"behind,omitempty"` // origin has commits local lacks
}

// Inspector reads local git state via the git CLI.
type Inspector struct {
	Run cliexec.Exec
}

// NewInspector returns an Inspector with a short timeout (status is cheap).
func NewInspector(run cliexec.Exec) *Inspector {
	if run == nil {
		run = cliexec.New()
	}
	return &Inspector{Run: run}
}

func (in *Inspector) git(ctx context.Context, dir string, args ...string) ([]byte, error) {
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
			return Status{Mapped: true, Path: abs, Error: err.Error()}
		}
		return summarize(abs, []Worktree{wt})
	}
	for i := range trees {
		if trees[i].Bare {
			continue
		}
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
	_, err := in.git(ctx, dir, "rev-parse", "--git-dir")
	return err == nil
}

func (in *Inspector) listWorktrees(ctx context.Context, dir string) ([]Worktree, error) {
	out, err := in.git(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var trees []Worktree
	var cur *Worktree
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			cur = nil
			continue
		}
		key, val, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			wt := Worktree{Path: val, Main: len(trees) == 0}
			trees = append(trees, wt)
			cur = &trees[len(trees)-1]
		case "bare":
			if cur != nil {
				cur.Bare = true
			}
		case "detached":
			if cur != nil {
				cur.Detached = true
			}
		case "branch":
			if cur != nil {
				cur.Branch = strings.TrimPrefix(val, "refs/heads/")
			}
		case "locked":
			if cur != nil {
				cur.Locked = true
			}
		case "prunable":
			if cur != nil {
				cur.Prunable = true
			}
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

func (in *Inspector) inspectWorktree(ctx context.Context, dir string, isMain bool) (Worktree, error) {
	wt := Worktree{Path: dir, Main: isMain}

	branchOut, err := in.git(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return wt, err
	}
	branch := strings.TrimSpace(string(branchOut))
	if branch == "HEAD" {
		wt.Detached = true
		short, _ := in.git(ctx, dir, "rev-parse", "--short", "HEAD")
		wt.Branch = strings.TrimSpace(string(short))
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
		ahead, behind, ok := in.leftRight(ctx, dir, "@{upstream}", "HEAD")
		if ok {
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

func (in *Inspector) leftRight(ctx context.Context, dir, left, right string) (ahead, behind int, ok bool) {
	// git rev-list --left-right --count A...B → "<behind>\t<ahead>" relative to B vs A
	// With A=upstream and B=HEAD: left=commits reachable from A not B (behind), right=ahead.
	out, err := in.git(ctx, dir, "rev-list", "--left-right", "--count", left+"..."+right)
	if err != nil {
		return 0, 0, false
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0, false
	}
	behind, err1 := strconv.Atoi(parts[0])
	ahead, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return ahead, behind, true
}

// EnrichOriginSync compares each local branch to origin/<same name>.
// DefaultAhead/DefaultBehind copy the default-branch row when both refs exist.
// Comparisons run in parallel to limit wall time when many branches share an origin twin.
func (in *Inspector) EnrichOriginSync(ctx context.Context, st *Status) {
	if st == nil || !st.Mapped || st.Path == "" || st.Error != "" {
		return
	}
	if def, ok := in.defaultBranch(ctx, st.Path); ok {
		st.DefaultBranch = def
	}
	localNames, err := in.listRefShortNames(ctx, st.Path, "refs/heads/")
	if err != nil {
		if st.Error == "" {
			st.Error = "list local branches: " + err.Error()
		}
		return
	}
	remoteNames, err := in.listRefShortNames(ctx, st.Path, "refs/remotes/origin/")
	if err != nil {
		if st.Error == "" {
			st.Error = "list origin branches: " + err.Error()
		}
		return
	}
	remoteSet := make(map[string]struct{}, len(remoteNames))
	for _, n := range remoteNames {
		n = strings.TrimPrefix(n, "origin/")
		if n == "" || n == "HEAD" {
			continue
		}
		remoteSet[n] = struct{}{}
	}
	var names []string
	for _, name := range localNames {
		if _, ok := remoteSet[name]; ok {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		st.OriginSync = nil
		return
	}

	syncs := make([]BranchSync, len(names))
	var wg sync.WaitGroup
	var failed atomic.Bool
	for i, name := range names {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			ahead, behind, ok := in.leftRight(ctx, st.Path, "refs/remotes/origin/"+name, "refs/heads/"+name)
			if !ok {
				failed.Store(true)
				return
			}
			syncs[i] = BranchSync{Name: name, Ahead: ahead, Behind: behind}
		}(i, name)
	}
	wg.Wait()

	if failed.Load() {
		if st.Error == "" {
			st.Error = "could not compare local branches to origin"
		}
		st.OriginSync = nil
		return
	}
	out := make([]BranchSync, 0, len(syncs))
	for _, s := range syncs {
		if s.Name == "" {
			continue
		}
		out = append(out, s)
		if s.Name == st.DefaultBranch {
			st.DefaultAhead = s.Ahead
			st.DefaultBehind = s.Behind
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	st.OriginSync = out
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

func (in *Inspector) defaultBranch(ctx context.Context, dir string) (string, bool) {
	out, err := in.git(ctx, dir, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(string(out))
		ref = strings.TrimPrefix(ref, "refs/remotes/origin/")
		if ref != "" {
			return ref, true
		}
	}
	for _, name := range []string{"main", "master"} {
		if _, err := in.git(ctx, dir, "rev-parse", "--verify", "refs/remotes/origin/"+name); err == nil {
			return name, true
		}
	}
	return "", false
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
