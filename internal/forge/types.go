package forge

import (
	"context"

	"github.com/behaviorengineering/gitboard/internal/config"
)

// CIStatus is the latest pipeline or workflow run.
type CIStatus struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion,omitempty"`
	Ref        string `json:"ref,omitempty"`
	Name       string `json:"name,omitempty"`
	WebURL     string `json:"web_url,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	RunID      string `json:"run_id,omitempty"`
}

// OpenItems counts open review work.
type OpenItems struct {
	PullRequests  int `json:"pull_requests"`
	MergeRequests int `json:"merge_requests"`
}

// ProjectSummary is one row in the dashboard.
type ProjectSummary struct {
	ID        string     `json:"id"`
	Label     string     `json:"label"`
	Host      string     `json:"host"`
	OpenURL   string     `json:"open_url"`
	CI        *CIStatus  `json:"ci,omitempty"`
	OpenItems OpenItems  `json:"open_items"`
	Error     string     `json:"error,omitempty"`
}

// FailedJob is a failed CI job suitable for triage.
type FailedJob struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Stage  string `json:"stage,omitempty"`
	WebURL string `json:"web_url,omitempty"`
}

// Tooling reports which forge CLIs are available and authenticated.
type Tooling struct {
	GitHub struct {
		Installed bool   `json:"installed"`
		Authed    bool   `json:"authed"`
		Detail    string `json:"detail,omitempty"`
	} `json:"github"`
	GitLab struct {
		Installed bool   `json:"installed"`
		Authed    bool   `json:"authed"`
		Detail    string `json:"detail,omitempty"`
	} `json:"gitlab"`
}

// Dashboard is the aggregated pane payload.
type Dashboard struct {
	GeneratedAt string           `json:"generated_at"`
	Tooling     Tooling          `json:"tooling"`
	Projects    []ProjectSummary `json:"projects"`
}

// Client summarizes one forge via its official CLI.
type Client interface {
	AuthStatus(ctx context.Context) (installed, authed bool, detail string)
	ProjectSummary(ctx context.Context, p config.Project) (ProjectSummary, error)
	FailedJobs(ctx context.Context, p config.Project, runID string) ([]FailedJob, error)
	JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error)
}
