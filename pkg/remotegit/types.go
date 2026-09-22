package remotegit

import (
	"context"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
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
	ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error)
	FailedJobs(ctx context.Context, p config.Project, runID string) ([]board.FailedJob, error)
	JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error)
	MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error)
}

// RepoRef is a discoverable repository from an org or group.
type RepoRef struct {
	Host     config.Host
	Path     string // owner/repo
	Name     string
	Archived bool
}
