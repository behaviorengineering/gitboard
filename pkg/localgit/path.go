package localgit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
)

const (
	// defaultBranchFallback is the default branch assumed when none is given
	// and origin/HEAD does not name one.
	defaultBranchFallback = "main"
	// gitDirName is the git metadata entry at the root of a working tree.
	gitDirName = ".git"
)

// ExpandPath resolves ~ and cleans the path.
func ExpandPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("empty path")
	}
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("abs %s: %w", p, err)
	}
	return abs, nil
}

// ResolvePath expands ~, makes the path absolute, resolves symlinks when it
// can, and cleans it. Every path compared against a mapped checkout goes
// through it, so the allowlist and the mutators agree on one spelling.
func ResolvePath(p string) (string, error) {
	abs, err := ExpandPath(p)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return filepath.Clean(abs), nil
}

// TrackKey is host:lowercase(owner/repo) for matching.
func TrackKey(host boardconfig.Host, path string) string {
	return strings.ToLower(strings.TrimSpace(string(host))) + ":" + strings.ToLower(strings.Trim(strings.TrimSpace(path), "/"))
}
