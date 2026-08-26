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

// BranchRef is a branch surfaced on the dashboard.
type BranchRef struct {
	Name       string `json:"name"`
	Default    bool   `json:"default,omitempty"`
	OpenReview bool   `json:"open_review,omitempty"`
	ReviewID   int    `json:"review_id,omitempty"`
	Draft      bool   `json:"draft,omitempty"`
	Conflict   bool   `json:"conflict,omitempty"`
	Stale      bool   `json:"stale,omitempty"`
	CIStatus   string `json:"ci_status,omitempty"`
	CIURL      string `json:"ci_url,omitempty"`
	RunID      string `json:"run_id,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	WebURL     string `json:"web_url,omitempty"`
}

// MergedReview is a recently merged PR/MR keyed by source branch name.
type MergedReview struct {
	Branch   string `json:"branch"`
	ID       int    `json:"id,omitempty"`
	URL      string `json:"url,omitempty"`
	MergedAt string `json:"merged_at,omitempty"`
}

// Prune hint values for local worktrees whose remote head is gone.
const (
	PruneSafe   = "safe"   // forge reports a merged PR/MR for this branch
	PruneLikely = "likely" // remote gone, no open review, not dirty
)

// ProjectSummary is one row in the dashboard.
type ProjectSummary struct {
	ID        string         `json:"id"`
	Label     string         `json:"label"`
	Host      string         `json:"host"`
	Path      string         `json:"path,omitempty"`
	Org       string         `json:"org,omitempty"`
	OpenURL   string         `json:"open_url"`
	CI        *CIStatus      `json:"ci,omitempty"`
	Branches  []BranchRef    `json:"branches,omitempty"`
	Merged    []MergedReview `json:"merged,omitempty"`
	OpenItems OpenItems      `json:"open_items"`
	Local     *LocalStatus   `json:"local,omitempty"`
	Error     string         `json:"error,omitempty"`

	// RemoteNames is the full remote head set for prune matching (not sent to UI).
	// Loaded via paginated forge API calls; may be incomplete only when RemoteNamesOK is false.
	RemoteNames []string `json:"-"`
	// RemoteNamesOK is true when heads were loaded successfully (live or cache).
	// When false, EnrichPruneHints suppresses all prune hints (fail closed).
	RemoteNamesOK bool `json:"-"`
	// MergedOK is true when the merged slice was loaded successfully (live or fresh cache).
	MergedOK bool `json:"-"`
}

// LocalAppearance roles match localgit checkout roles.
const (
	AppearanceStandalone = "standalone"
	AppearanceSubmodule  = "submodule"
)

// LocalAppearance is one on-disk checkout of a forge project (standalone or submodule).
type LocalAppearance struct {
	Role          string             `json:"role,omitempty"` // standalone | submodule
	Path          string             `json:"path,omitempty"`
	DisplayID     string             `json:"display_id,omitempty"`
	ParentPath    string             `json:"parent_path,omitempty"`
	ParentLabel   string             `json:"parent_label,omitempty"`
	RelPath       string             `json:"rel_path,omitempty"`
	Primary       bool               `json:"primary,omitempty"`
	Error         string             `json:"error,omitempty"`
	Branch        string             `json:"branch,omitempty"`
	Detached      bool               `json:"detached,omitempty"`
	Dirty         bool               `json:"dirty,omitempty"`
	Ahead         int                `json:"ahead,omitempty"`
	Behind        int                `json:"behind,omitempty"`
	Upstream      string             `json:"upstream,omitempty"`
	DefaultBranch string             `json:"default_branch,omitempty"`
	DefaultBehind int                `json:"default_behind,omitempty"`
	DefaultAhead  int                `json:"default_ahead,omitempty"`
	OriginSync    []BranchOriginSync `json:"origin_sync,omitempty"`
	Worktrees     []LocalWorktree    `json:"worktrees,omitempty"`
}

// LocalStatus is checkout / worktree health on disk.
// Top-level fields mirror the primary appearance for backward compatibility.
type LocalStatus struct {
	Mapped        bool               `json:"mapped"`
	Path          string             `json:"path,omitempty"`
	Error         string             `json:"error,omitempty"`
	Branch        string             `json:"branch,omitempty"`
	Detached      bool               `json:"detached,omitempty"`
	Dirty         bool               `json:"dirty,omitempty"`
	Ahead         int                `json:"ahead,omitempty"`
	Behind        int                `json:"behind,omitempty"`
	Upstream      string             `json:"upstream,omitempty"`
	DefaultBranch string             `json:"default_branch,omitempty"`
	DefaultBehind int                `json:"default_behind,omitempty"`
	DefaultAhead  int                `json:"default_ahead,omitempty"`
	OriginSync    []BranchOriginSync `json:"origin_sync,omitempty"`
	Worktrees     []LocalWorktree    `json:"worktrees,omitempty"`
	Appearances   []LocalAppearance  `json:"appearances,omitempty"`
}

// BranchOriginSync is local refs/heads/<name> versus origin/<name>.
type BranchOriginSync struct {
	Name   string `json:"name"`
	Ahead  int    `json:"ahead,omitempty"`
	Behind int    `json:"behind,omitempty"`
}

// LocalWorktree is one git worktree for a mapped project.
type LocalWorktree struct {
	Path            string `json:"path"`
	Branch          string `json:"branch,omitempty"`
	Detached        bool   `json:"detached,omitempty"`
	Bare            bool   `json:"bare,omitempty"`
	Main            bool   `json:"main,omitempty"`
	Dirty           bool   `json:"dirty,omitempty"`
	Ahead           int    `json:"ahead,omitempty"`
	Behind          int    `json:"behind,omitempty"`
	Upstream        string `json:"upstream,omitempty"`
	PruneHint       string `json:"prune_hint,omitempty"` // safe | likely
	MergedID        int    `json:"merged_id,omitempty"`
	MergedURL       string `json:"merged_url,omitempty"`
	MergedAt        string `json:"merged_at,omitempty"`
	AppearancePath  string `json:"appearance_path,omitempty"`
	AppearanceLabel string `json:"appearance_label,omitempty"`
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
	GeneratedAt         string           `json:"generated_at"`
	PollIntervalSeconds int              `json:"poll_interval_seconds"`
	Tooling             Tooling          `json:"tooling"`
	Projects            []ProjectSummary `json:"projects"`
}

// RepoRef is a discoverable repository from an org or group.
type RepoRef struct {
	Host     config.Host
	Path     string // owner/repo
	Name     string
	Archived bool
}

// Client summarizes one forge via its official CLI.
type Client interface {
	AuthStatus(ctx context.Context) (installed, authed bool, detail string)
	ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (ProjectSummary, error)
	FailedJobs(ctx context.Context, p config.Project, runID string) ([]FailedJob, error)
	JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error)
}
