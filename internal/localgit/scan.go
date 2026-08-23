package localgit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Discovery maps forge track keys to a representative checkout path.
type Discovery struct {
	// ByKey maps "github:owner/repo" → absolute checkout path (prefer main worktree).
	ByKey map[string]string
}

const maxScanDepth = 5

// ScanRoots walks roots for git checkouts and indexes them by origin remote.
func (in *Inspector) ScanRoots(ctx context.Context, roots []string) Discovery {
	d := Discovery{ByKey: map[string]string{}}
	seenCommon := map[string]string{} // common git dir → preferred path
	for _, root := range roots {
		abs, err := ExpandPath(root)
		if err != nil {
			continue
		}
		st, err := os.Stat(abs)
		if err != nil || !st.IsDir() {
			continue
		}
		_ = filepath.WalkDir(abs, func(path string, de os.DirEntry, walkErr error) error {
			if walkErr != nil {
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
	}
	return d
}

func (in *Inspector) indexCheckout(ctx context.Context, path string, d Discovery, seenCommon map[string]string) {
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
	// Prefer the main worktree when multiple checkouts share a common dir.
	trees, listErr := in.listWorktrees(ctx, path)
	preferred := path
	if listErr == nil {
		for _, wt := range trees {
			if wt.Main && !wt.Bare {
				preferred = wt.Path
				break
			}
		}
	}
	if existing, ok := seenCommon[common]; ok {
		// Keep first preferred; still ensure ByKey is set.
		_ = existing
	} else {
		seenCommon[common] = preferred
	}
	if _, exists := d.ByKey[key]; !exists {
		d.ByKey[key] = preferred
	}
}

// ResolvePath picks an explicit local_path or a scanned match.
func ResolvePath(localPath string, host, repoPath string, disc Discovery) (string, bool) {
	if strings.TrimSpace(localPath) != "" {
		return strings.TrimSpace(localPath), true
	}
	if disc.ByKey == nil {
		return "", false
	}
	p, ok := disc.ByKey[TrackKey(host, repoPath)]
	return p, ok
}
