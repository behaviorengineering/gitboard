package remotegit

import (
	"context"
	"time"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
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

// SummaryOpts controls which forge slices use the TTL cache.
type SummaryOpts struct {
	Fresh     bool
	Cache     *TTLCache
	HeadsTTL  time.Duration // 0 = always refetch heads
	MergedTTL time.Duration // 0 = always refetch merged
}

// Client summarizes one forge via its official CLI.
type Client interface {
	AuthStatus(ctx context.Context) (installed, authed bool, detail string)
	ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error)
	FailedJobs(ctx context.Context, p config.Project, runID string) ([]board.FailedJob, error)
	JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error)
}

// RepoRef is a discoverable repository from an org or group.
type RepoRef struct {
	Host     config.Host
	Path     string // owner/repo
	Name     string
	Archived bool
}
