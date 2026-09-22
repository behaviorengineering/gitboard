package localgit

import (
	"net/url"
	"strings"
)

// RemoteRef is a forge identity parsed from a git remote URL.
type RemoteRef struct {
	Host string // github | gitlab
	Path string // owner/repo (no .git)
}

// ParseRemoteURL maps common GitHub/GitLab clone URLs to host+path.
func ParseRemoteURL(raw string) (RemoteRef, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RemoteRef{}, false
	}
	raw = strings.TrimSuffix(raw, ".git")

	// git@host:owner/repo or ssh://git@host/owner/repo
	if strings.HasPrefix(raw, "git@") {
		rest := strings.TrimPrefix(raw, "git@")
		host, path, ok := strings.Cut(rest, ":")
		if !ok {
			return RemoteRef{}, false
		}
		return remoteFromHostPath(host, path)
	}
	if strings.HasPrefix(raw, "ssh://") {
		u, err := url.Parse(raw)
		if err != nil {
			return RemoteRef{}, false
		}
		return remoteFromHostPath(u.Host, strings.TrimPrefix(u.Path, "/"))
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil {
			return RemoteRef{}, false
		}
		return remoteFromHostPath(u.Host, strings.TrimPrefix(u.Path, "/"))
	}
	return RemoteRef{}, false
}

func remoteFromHostPath(host, path string) (RemoteRef, bool) {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, ok := strings.Cut(host, ":"); ok { // host:port
		host = h
	}
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return RemoteRef{}, false
	}
	// Keep full path for nested GitLab groups (group/sub/repo).
	path = strings.Join(parts, "/")

	var forge string
	switch {
	case host == "github.com" || strings.HasSuffix(host, ".github.com"):
		forge = "github"
	case host == "gitlab.com" || strings.Contains(host, "gitlab"):
		forge = "gitlab"
	default:
		return RemoteRef{}, false
	}
	return RemoteRef{Host: forge, Path: path}, true
}
