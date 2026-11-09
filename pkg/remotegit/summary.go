package remotegit

import (
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
)

// BaseSummary returns the identity fields of a project row (id, label, host,
// trimmed path, org, open URL and capabilities) with no forge data. Loaded
// rows, config-only shell rows and error stubs all start here, so they agree
// on one identity per project.
func BaseSummary(p boardconfig.Project) board.ProjectSummary {
	return board.ProjectSummary{
		ID:           p.ID,
		Label:        p.Label,
		Host:         string(p.Host),
		Path:         strings.Trim(p.Path, "/"),
		Org:          pathOrg(p.Path),
		OpenURL:      p.OpenURL(),
		Capabilities: HostCapabilities(p.Host),
	}
}

// HostCapabilities returns the UI action surface for a forge host.
// GitHub and GitLab support failed-job listing and AI triage; Azure DevOps and
// Bitbucket expose CI status and links only until those adapters implement logs.
func HostCapabilities(host boardconfig.Host) board.Capabilities {
	switch host {
	case boardconfig.HostGitHub, boardconfig.HostGitLab:
		return board.Capabilities{FailedJobs: true}
	case boardconfig.HostAzureDevOps, boardconfig.HostBitbucket:
		return board.Capabilities{FailedJobs: false}
	default:
		return board.Capabilities{FailedJobs: false}
	}
}

// pathOrg returns every path segment but the last (owner, group or
// org/project), or "" for a single-segment path.
func pathOrg(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}
