package remotegit

import (
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
)

func baseSummary(p config.Project) board.ProjectSummary {
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
func HostCapabilities(host config.Host) board.Capabilities {
	switch host {
	case config.HostGitHub, config.HostGitLab:
		return board.Capabilities{FailedJobs: true}
	default:
		return board.Capabilities{FailedJobs: false}
	}
}

func pathOrg(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func truncateLog(s string) string {
	const max = 96 * 1024
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}

func githubHasConflict(mergeable, mergeState string) bool {
	m := strings.ToUpper(strings.TrimSpace(mergeable))
	s := strings.ToUpper(strings.TrimSpace(mergeState))
	return m == "CONFLICTING" || s == "DIRTY" || s == "CONFLICTING"
}

func gitlabHasConflict(hasConflicts bool, mergeStatus string) bool {
	if hasConflicts {
		return true
	}
	s := strings.ToLower(strings.TrimSpace(mergeStatus))
	return s == "cannot_be_merged" || s == "cannot_be_merged_recheck"
}
