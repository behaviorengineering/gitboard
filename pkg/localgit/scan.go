package localgit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Discovery maps forge track keys to all distinct on-disk checkouts.
type Discovery struct {
	// ByKey maps "github:owner/repo" → checkouts (deduped by common git dir).
	ByKey map[string][]Checkout
}

const maxScanDepth = 5

// ScanRoots walks roots for git checkouts and indexes them by origin remote.
func (in *Inspector) ScanRoots(ctx context.Context, roots []string) Discovery {
	d := Discovery{ByKey: map[string][]Checkout{}}
	seenCommon := map[string]struct{}{} // common git dir already indexed
	for _, root := range roots {
		abs, err := ExpandPath(root)
		if err != nil {
			continue
		}
		st, err := os.Stat(abs)
		if err != nil || !st.IsDir() {
			continue
		}
		walkErr := filepath.WalkDir(abs, func(path string, de os.DirEntry, entryErr error) error {
			if entryErr != nil {
				// Skip individual unreadable entries but propagate ctx cancel.
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !de.IsDir() {
				return nil
			}
			name := de.Name()
			if name == "node_modules" || name == "vendor" || name == ".git" {
				if name == ".git" {
					return filepath.SkipDir
				}
				return filepath.SkipDir
			}
			rel, err := filepath.Rel(abs, path)
			if err != nil {
				return nil
			}
			if rel != "." {
				depth := strings.Count(rel, string(os.PathSeparator)) + 1
				if depth > maxScanDepth {
					return filepath.SkipDir
				}
			}
			gitMeta := filepath.Join(path, ".git")
			if _, err := os.Stat(gitMeta); err != nil {
				return nil
			}
			in.indexCheckout(ctx, path, d, seenCommon)
			// Keep walking so nested clones / submodules are indexed too.
			return nil
		})
		if walkErr != nil && ctx.Err() != nil {
			// Context canceled; stop processing further roots.
			return d
		}
	}
	return d
}

func (in *Inspector) indexCheckout(ctx context.Context, path string, d Discovery, seenCommon map[string]struct{}) {
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
	key := TrackKey(ref.Host, ref.Path)
	common, err := in.CommonGitDir(ctx, path)
	if err != nil {
		common = path
	}
	common = filepath.Clean(common)
	if _, exists := seenCommon[common]; exists {
		return
	}

	// Prefer the main worktree when multiple checkouts share a common dir,
	// but never replace a working tree with the bare/common git directory.
	trees, listErr := in.listWorktrees(ctx, path)
	preferred := path
	if listErr == nil {
		for _, wt := range trees {
			if !wt.Main || wt.Bare || wt.Path == "" {
				continue
			}
			cand := filepath.Clean(wt.Path)
			if cand == common || !isWorkingTreeRoot(cand) {
				continue
			}
			preferred = cand
			break
		}
	}

	c := Checkout{
		Path:         preferred,
		CommonGitDir: common,
		Superproject: in.Superproject(ctx, preferred),
	}
	FillCheckoutMeta(&c)
	seenCommon[common] = struct{}{}
	d.ByKey[key] = append(d.ByKey[key], c)
}

// isWorkingTreeRoot reports whether path has a .git file or directory (a work tree).
func isWorkingTreeRoot(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
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
