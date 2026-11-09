// Package remotegit is public API: forge clients for GitHub, GitLab, Azure
// DevOps and Bitbucket that build board rows (pkg/board) for tracked projects
// (pkg/boardconfig) through the forge CLIs, with an in-memory TTL cache.
// Test fixtures live in pkg/remotegit/remotegittest.
package remotegit

import (
	"context"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
)

// HeadsSnapshot is the cached remote branch list for one project.
type HeadsSnapshot struct {
	DefaultBranch string
	Heads         []RemoteHead
}

// RemoteHead is one forge branch tip.
type RemoteHead struct {
	Name      string
	UpdatedAt string
	WebURL    string
}

// Client summarizes one forge via its official CLI.
type Client interface {
	AuthStatus(ctx context.Context) (installed, authed bool, detail string)
	ProjectSummary(ctx context.Context, p boardconfig.Project, opts SummaryOpts) (board.ProjectSummary, error)
	FailedJobs(ctx context.Context, p boardconfig.Project, runID string) ([]board.FailedJob, error)
	JobLog(ctx context.Context, p boardconfig.Project, runID, jobID string) (string, error)
	MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error)
}

// RepoRef is a discoverable repository from an org or group.
type RepoRef struct {
	Host boardconfig.Host
	Path string // owner/repo.
	Name string
}
