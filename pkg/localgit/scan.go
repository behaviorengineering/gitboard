package localgit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Discovery maps forge track keys to all distinct on-disk checkouts.
type Discovery struct {
	// ByKey maps "github:owner/repo" → checkouts (deduped by common git dir).
	ByKey map[string][]Checkout
	// Warnings lists configured roots that were skipped (bad path, missing,
	// not a directory) or whose walk stopped early, so an empty board can be
	// explained to the operator.
	Warnings []string
}

const maxScanDepth = 5

// skipDirNames are directories ScanRoots never descends into.
var skipDirNames = map[string]struct{}{"node_modules": {}, "vendor": {}, gitDirName: {}}

// ScanRoots walks roots for git checkouts and indexes them by origin remote.
func (in *Inspector) ScanRoots(ctx context.Context, roots []string) Discovery {
	d := Discovery{ByKey: map[string][]Checkout{}}
	seenCommon := map[string]struct{}{} // common git dir already indexed.
	for _, root := range roots {
		if err := in.walkRoot(ctx, root, &d, seenCommon); err != nil {
			d.Warnings = append(d.Warnings, fmt.Sprintf("local root %q: %v", root, err))
		}
		if ctx.Err() != nil {
			// Context cancelled; stop processing further roots.
			return d
		}
	}
	return d
}

// walkRoot indexes every checkout under root, down to maxScanDepth. It
// returns why root was skipped or why its walk stopped early.
func (in *Inspector) walkRoot(ctx context.Context, root string, d *Discovery, seenCommon map[string]struct{}) error {
	abs, err := ExpandPath(root)
	if err != nil {
		return err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("not a directory")
	}
	walkErr := filepath.WalkDir(abs, func(path string, de os.DirEntry, entryErr error) error {
		// Skip individual unreadable entries but propagate ctx cancel.
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if entryErr != nil || !de.IsDir() {
			return nil
		}
		skip, checkout := walkDecision(abs, path, de.Name())
		if checkout {
			in.indexCheckout(ctx, path, d, seenCommon)
		}
		// Keep walking below a checkout so nested clones and submodules are indexed too.
		return skip
	})
	if walkErr != nil {
		return fmt.Errorf("walk stopped: %w", walkErr)
	}
	return nil
}

// walkDecision says whether WalkDir should skip the directory at path under
// root (filepath.SkipDir) and whether path is a checkout to index.
func walkDecision(root, path, name string) (skip error, checkout bool) {
	if _, ok := skipDirNames[name]; ok {
		return filepath.SkipDir, false
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, false
	}
	if rel != "." && strings.Count(rel, string(os.PathSeparator))+1 > maxScanDepth {
		return filepath.SkipDir, false
	}
	return nil, isWorkingTreeRoot(path)
}

func (in *Inspector) indexCheckout(ctx context.Context, path string, d *Discovery, seenCommon map[string]struct{}) {
	if !in.isGitDir(ctx, path) {
		return
	}
	url, err := in.OriginRemote(ctx, path)
	if err != nil {
		return
	}
	ref, ok := ParseRemoteURL(url)
	if !ok {
		return
	}
	common, err := CommonKey(ctx, in.CommonGitDir, Checkout{Path: path})
	if err != nil {
		// Indexed under its own path: it still shows on the board, and
		// mutations re-resolve the common dir strictly.
		d.Warnings = append(d.Warnings, fmt.Sprintf("checkout %q: %v", path, err))
	}
	if _, exists := seenCommon[common]; exists {
		return
	}
	c := Checkout{
		Path:         in.preferredWorktree(ctx, path, common),
		CommonGitDir: common,
	}
	c.Superproject = in.Superproject(ctx, c.Path)
	FillCheckoutMeta(&c)
	seenCommon[common] = struct{}{}
	key := TrackKey(ref.Host, ref.Path)
	d.ByKey[key] = append(d.ByKey[key], c)
}

// preferredWorktree returns the main worktree when several checkouts share
// common, but never replaces a working tree with the bare or common git dir.
func (in *Inspector) preferredWorktree(ctx context.Context, path, common string) string {
	trees, err := in.listWorktrees(ctx, path)
	if err != nil {
		return path
	}
	for _, wt := range trees {
		if !wt.Main || wt.Bare || wt.Path == "" {
			continue
		}
		cand := filepath.Clean(wt.Path)
		if cand != common && isWorkingTreeRoot(cand) {
			return cand
		}
	}
	return path
}

// isWorkingTreeRoot reports whether path has a .git file or directory (a work tree).
func isWorkingTreeRoot(path string) bool {
	_, err := os.Stat(filepath.Join(path, gitDirName))
	return err == nil
}

// Superproject returns the superproject working tree for a submodule checkout.
func (in *Inspector) Superproject(ctx context.Context, dir string) string {
	out, err := in.git(ctx, dir, "rev-parse", "--show-superproject-working-tree")
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	if s == "" || s == "." {
		return ""
	}
	return filepath.Clean(s)
}
