package server

import (
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
)

// The embedded UI reads these JSON shapes. Typed structs keep the wire
// contract checkable: a renamed field fails to compile instead of failing in
// the browser.

// metaResponse is GET /api/meta.
type metaResponse struct {
	PollIntervalSeconds int                 `json:"poll_interval_seconds"`
	UI                  board.UIConfig      `json:"ui"`
	Views               []board.ViewSummary `json:"views"`
	Tooling             *board.Tooling      `json:"tooling,omitempty"`
	// Token is the launch token mutating requests must send in TokenHeader.
	// Empty on a loopback-only server.
	Token string `json:"token,omitempty"`
}

// failuresResponse is GET /api/failures.
type failuresResponse struct {
	Jobs []board.FailedJob `json:"jobs"`
}

// okResponse acknowledges a command with no other result.
type okResponse struct {
	OK bool `json:"ok"`
}

// candidatesResponse is GET /api/sync/candidates.
type candidatesResponse struct {
	Candidates []projectsync.Candidate `json:"candidates"`
	Warnings   []string                `json:"warnings"`
}

// rootsResponse is GET and PUT /api/local/roots.
type rootsResponse struct {
	Roots []string `json:"roots"`
}

// syncProject is one tracked project in syncState.
type syncProject struct {
	ID        string           `json:"id"`
	Label     string           `json:"label"`
	Host      boardconfig.Host `json:"host"`
	Path      string           `json:"path"`
	LocalPath string           `json:"local_path"`
}

// syncState is the Manage view of the config: tracked projects, views, sync
// sources and local roots.
type syncState struct {
	Projects []syncProject             `json:"projects"`
	Views    []board.ViewSummary       `json:"views"`
	ViewDefs []boardconfig.View        `json:"view_defs"`
	Sync     dashboard.SyncSourceLists `json:"sync"`
	Local    rootsResponse             `json:"local"`
}

// streamShell is the dashboard in the first NDJSON event. Tooling is left
// out: a zero value would read as missing CLIs, so the UI keeps the last
// strip until the done event.
type streamShell struct {
	GeneratedAt         string                 `json:"generated_at"`
	PollIntervalSeconds int                    `json:"poll_interval_seconds"`
	UI                  board.UIConfig         `json:"ui"`
	Views               []board.ViewSummary    `json:"views"`
	ActiveView          string                 `json:"active_view"`
	Projects            []board.ProjectSummary `json:"projects"`
}

// Stream event types, in the order the UI receives them.
const (
	eventShell   = "shell"
	eventProject = "project"
	eventError   = "error"
	eventDone    = "done"
)

// streamEvent is one NDJSON line of GET /api/dashboard?stream=1.
type streamEvent struct {
	Type          string                `json:"type"`
	Dashboard     *streamShell          `json:"dashboard,omitempty"`
	Project       *board.ProjectSummary `json:"project,omitempty"`
	Error         string                `json:"error,omitempty"`
	Code          gberr.Code            `json:"code,omitempty"`
	Tooling       *board.Tooling        `json:"tooling,omitempty"`
	GeneratedAt   string                `json:"generated_at,omitempty"`
	LocalWarnings []string              `json:"local_warnings,omitempty"`
}

func syncSourcesOf(doc boardconfig.File) dashboard.SyncSourceLists {
	return dashboard.SyncSourceLists{
		GitHubOrgs:          doc.Sync.GitHub.Orgs,
		GitLabGroups:        doc.Sync.GitLab.Groups,
		AzureDevOpsOrgs:     doc.Sync.AzureDevOps.Orgs,
		BitbucketWorkspaces: doc.Sync.Bitbucket.Workspaces,
	}
}

func syncStateOf(doc boardconfig.File) syncState {
	projects := make([]syncProject, 0, len(doc.Projects))
	for _, p := range doc.Projects {
		projects = append(projects, syncProject{
			ID:        p.ID,
			Label:     p.Label,
			Host:      p.Host,
			Path:      p.Path,
			LocalPath: p.LocalPath,
		})
	}
	views := doc.EffectiveViews()
	viewDefs := make([]boardconfig.View, len(views))
	copy(viewDefs, views)
	return syncState{
		Projects: projects,
		Views:    dashboard.ViewSummaries(doc),
		ViewDefs: viewDefs,
		Sync:     syncSourcesOf(doc),
		Local:    rootsResponse{Roots: doc.Local.Roots},
	}
}
