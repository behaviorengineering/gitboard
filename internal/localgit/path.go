package localgit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// TrackKey is host:lowercase(owner/repo) for matching.
func TrackKey(host, path string) string {
	return strings.ToLower(strings.TrimSpace(host)) + ":" + strings.ToLower(strings.Trim(path, "/"))
}
