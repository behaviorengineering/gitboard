package localgit

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Checkout roles for on-disk appearances of one forge remote.
const (
	RoleStandalone = "standalone"
	RoleSubmodule  = "submodule"
)

// Checkout is one distinct on-disk clone of a forge remote.
type Checkout struct {
	Path           string // absolute preferred worktree path
	CommonGitDir   string
	Superproject   string // absolute parent working tree when a submodule; empty if standalone
	Role           string // standalone | submodule
	RelPath        string // path relative to Superproject when submodule
	ParentBasename string // basename(Superproject) when submodule
}

// Appearances returns all checkouts for a forge track key.
func (d Discovery) Appearances(host, repoPath string) []Checkout {
	if d.ByKey == nil {
		return nil
	}
	out := d.ByKey[TrackKey(host, repoPath)]
	if len(out) == 0 {
		return nil
	}
	cp := make([]Checkout, len(out))
	copy(cp, out)
	return cp
}

// resolvePath picks an explicit local_path or the primary scanned checkout path.
func resolvePath(localPath string, host, repoPath string, disc Discovery) (string, bool) {
	if strings.TrimSpace(localPath) != "" {
		return strings.TrimSpace(localPath), true
	}
	c, ok := PickPrimary("", disc.Appearances(host, repoPath))
	if !ok {
		return "", false
	}
	return c.Path, true
}

// PickPrimary selects the preferred checkout.
// Order: explicit localPath (if it matches an appearance, or as a synthetic path),
// then standalone, then shallowest absolute path.
func PickPrimary(localPath string, checkouts []Checkout) (Checkout, bool) {
	localPath = strings.TrimSpace(localPath)
	if localPath != "" {
		abs, err := ExpandPath(localPath)
		if err == nil {
			for _, c := range checkouts {
				if filepath.Clean(c.Path) == filepath.Clean(abs) {
					return c, true
				}
			}
		}
		// Explicit override even when not in the scan list.
		return Checkout{Path: localPath, Role: RoleStandalone}, true
	}
	if len(checkouts) == 0 {
		return Checkout{}, false
	}
	sorted := make([]Checkout, len(checkouts))
	copy(sorted, checkouts)
	sort.SliceStable(sorted, func(i, j int) bool {
		si := sorted[i].Role == RoleStandalone
		sj := sorted[j].Role == RoleStandalone
		if si != sj {
			return si
		}
		return pathDepth(sorted[i].Path) < pathDepth(sorted[j].Path) ||
			(pathDepth(sorted[i].Path) == pathDepth(sorted[j].Path) && sorted[i].Path < sorted[j].Path)
	})
	return sorted[0], true
}

func pathDepth(p string) int {
	p = filepath.Clean(p)
	if p == "" || p == string(filepath.Separator) {
		return 0
	}
	return strings.Count(p, string(filepath.Separator))
}

// homeShortPath replaces the home directory prefix with ~/.
func homeShortPath(abs string) string {
	abs = filepath.Clean(strings.TrimSpace(abs))
	if abs == "" {
		return abs
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return abs
	}
	home = filepath.Clean(home)
	if abs == home {
		return "~"
	}
	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(abs, prefix) {
		return "~/" + abs[len(prefix):]
	}
	return abs
}

// DisplayID builds the UI identifier for a checkout.
// parentLabel is the forge project label when known; otherwise ParentBasename is used.
func DisplayID(c Checkout, parentLabel string) string {
	if c.Role == RoleSubmodule && c.Superproject != "" {
		label := strings.TrimSpace(parentLabel)
		if label == "" {
			label = c.ParentBasename
		}
		if label == "" {
			label = filepath.Base(c.Superproject)
		}
		rel := c.RelPath
		if rel == "" || rel == "." {
			rel = filepath.Base(c.Path)
		}
		return label + " → " + rel
	}
	return homeShortPath(c.Path)
}

// FillCheckoutMeta sets Role, RelPath, and ParentBasename from Path and Superproject.
func FillCheckoutMeta(c *Checkout) {
	if c == nil {
		return
	}
	c.Path = filepath.Clean(c.Path)
	if strings.TrimSpace(c.Superproject) == "" {
		c.Role = RoleStandalone
		c.RelPath = ""
		c.ParentBasename = ""
		return
	}
	c.Superproject = filepath.Clean(c.Superproject)
	c.Role = RoleSubmodule
	c.ParentBasename = filepath.Base(c.Superproject)

	child := c.Path
	parent := c.Superproject
	if resolved, err := filepath.EvalSymlinks(child); err == nil {
		child = resolved
	}
	if resolved, err := filepath.EvalSymlinks(parent); err == nil {
		parent = resolved
		c.ParentBasename = filepath.Base(parent)
	}
	rel, err := filepath.Rel(parent, child)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		c.RelPath = filepath.Base(c.Path)
		return
	}
	c.RelPath = rel
}
