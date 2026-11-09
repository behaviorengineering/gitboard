package boardconfig

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnknownHost is returned by ParseHost for a value outside the supported forges.
var ErrUnknownHost = errors.New("host must be github, gitlab, azuredevops, or bitbucket")

// Hosts lists every supported forge in display order.
func Hosts() []Host {
	return []Host{HostGitHub, HostGitLab, HostAzureDevOps, HostBitbucket}
}

// Valid reports whether h is one of the supported forge hosts.
func (h Host) Valid() bool {
	switch h {
	case HostGitHub, HostGitLab, HostAzureDevOps, HostBitbucket:
		return true
	default:
		return false
	}
}

// ParseHost trims and lower-cases s and returns it as a supported Host.
// Any other value returns ErrUnknownHost.
func ParseHost(s string) (Host, error) {
	h := Host(strings.ToLower(strings.TrimSpace(s)))
	if !h.Valid() {
		return "", ErrUnknownHost
	}
	return h, nil
}

// UsesMergeRequests reports whether the forge calls its reviews merge
// requests (GitLab) rather than pull requests, which picks the open-items
// counter a row fills.
func (h Host) UsesMergeRequests() bool {
	return h == HostGitLab
}

// SplitRepoPath splits a tracked project path into its segments and checks
// the shape the forge needs: org/project/repo for Azure DevOps, owner/repo
// (with optional nested groups) for every other host. Leading and trailing
// slashes are ignored; an empty segment is rejected.
func SplitRepoPath(h Host, path string) ([]string, error) {
	minParts, hint := 2, "owner/repo"
	switch h {
	case HostAzureDevOps:
		minParts, hint = 3, "org/project/repo"
	case HostBitbucket:
		hint = "workspace/repo"
	case HostGitHub, HostGitLab:
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < minParts {
		return nil, fmt.Errorf("path must be %s", hint)
	}
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("path must be %s", hint)
		}
	}
	return parts, nil
}
