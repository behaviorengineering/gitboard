package dashboard

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// PruneSafeRequest identifies a local checkout marked safe to remove.
type PruneSafeRequest struct {
	ProjectID      string `json:"project_id"`
	Branch         string `json:"branch"`
	WorktreePath   string `json:"worktree_path"`
	ClearIndexLock bool   `json:"clear_index_lock,omitempty"`
}

// PruneSafe re-checks the prune verdict with fresh local state, then removes
// the local checkout only when pruneVerdict still allows it.
func (c *Commands) PruneSafe(ctx context.Context, doc boardconfig.File, req PruneSafeRequest) error {
	const op = "dashboard.PruneSafe"
	s, err := c.requireLocal()
	if err != nil {
		return err
	}
	t, err := parseMutation(op, doc, mutationRequest{
		ProjectID: req.ProjectID, Branch: req.Branch, Path: req.WorktreePath, PathField: "worktree_path",
	})
	if err != nil {
		return err
	}
	t.path = canonicalCheckoutPath(ctx, s.local, t.path)
	wt, defaultBranch, err := s.pruneTarget(ctx, doc, t)
	if err != nil {
		return err
	}
	common, err := s.commonDirFor(ctx, op, wt.Path)
	if err != nil {
		return err
	}
	err = s.withLease(ctx, common, func(m LocalMutator) (bool, error) {
		err := runWithIndexLock(ctx, m, wt.Path, req.ClearIndexLock, func() error {
			return m.RemoveSafeCheckout(ctx, wt.Path, t.branch, defaultBranch)
		})
		if err != nil {
			return false, gitMutationErr(op, err)
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	s.afterMutation(common, t.project)
	return nil
}

// pruneTarget rebuilds t.project's row with fresh local state and prune
// evidence, finds the worktree on t.branch at t.path, and returns it with
// the default branch to land on when pruneVerdict allows the delete.
func (s *Service) pruneTarget(ctx context.Context, doc boardconfig.File, t mutationTarget) (board.LocalWorktree, string, error) {
	const op = "dashboard.PruneSafe"
	opts := remotegit.SummaryOpts{
		Now:       s.clock,
		Cache:     s.cache,
		HeadsTTL:  time.Duration(doc.EffectiveHeadsSeconds()) * time.Second,
		MergedTTL: time.Duration(doc.EffectiveMergedSeconds()) * time.Second,
	}
	row := s.summarize(ctx, t.project, opts)
	if !row.RemoteNamesOK {
		return board.LocalWorktree{}, "", conflict(op, "cannot re-validate remote heads; refuse prune")
	}
	if !row.OpenOK {
		return board.LocalWorktree{}, "", conflict(op, "cannot re-validate open reviews; refuse prune")
	}
	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.local.ScanRoots(ctx, doc.Local.Roots)
	}
	row.Local = s.attachLocal(ctx, t.project, localInputs{
		disc:          disc,
		labels:        projectLabelsByKey(doc.Projects),
		fetchTTL:      time.Duration(doc.EffectiveFetchSeconds()) * time.Second,
		forgeBranches: forgeBranchNames(filterHiddenBranches(row.Branches, doc.UI.HideBranches)),
	})
	if row.Local == nil || !row.Local.Mapped {
		return board.LocalWorktree{}, "", invalid(op, "project has no mapped local checkout")
	}
	canon := func(p string) string { return canonicalCheckoutPath(ctx, s.local, p) }
	if !repoPathAllowedCanon(row.Local, t.path, canon) {
		return board.LocalWorktree{}, "", invalid(op, "worktree_path is not a mapped checkout for this project")
	}
	s.annotatePrune(ctx, t.project, &row, opts)

	wt, ok := findWorktree(row.Local, t.branch, t.path, canon)
	if !ok {
		return board.LocalWorktree{}, "", conflict(op, fmt.Sprintf("branch %q is not checked out at worktree_path", t.branch))
	}
	sc := newPruneScope(&row)
	if v := pruneVerdict(sc.inputs(wt)); !v.Allow {
		return board.LocalWorktree{}, "", conflict(op, fmt.Sprintf("branch %q is not safe to remove: %s", t.branch, v.Reason))
	}
	return wt, sc.defaultBranch, nil
}

// findWorktree returns the worktree on branch whose path (or appearance
// path) is absPath, with Path set to absPath.
func findWorktree(local *board.LocalStatus, branch, absPath string, canon func(string) string) (board.LocalWorktree, bool) {
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

// cleanRepoPath is localgit.ResolvePath, or the cleaned input when the path
// cannot be expanded; empty stays empty.
func cleanRepoPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if resolved, err := localgit.ResolvePath(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

func canonicalCheckoutPath(ctx context.Context, local LocalComparer, path string) string {
	path = cleanRepoPath(path)
	if path == "" || local == nil {
		return path
	}
	return local.CanonicalCheckoutPath(ctx, path)
}
