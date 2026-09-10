package dashboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

// SyncInvestigationRequest identifies a mapped checkout and branch to inspect.
type SyncInvestigationRequest struct {
	ProjectID string `json:"project_id"`
	Branch    string `json:"branch"`
	RepoPath  string `json:"repo_path"`
}

// SyncInvestigation is the safe read-only result returned to the dashboard.
type SyncInvestigation struct {
	ProjectID string                  `json:"project_id"`
	Result    localgit.SyncInspection `json:"result"`
}

// InvestigateSync validates a mapped checkout, refreshes origin, and compares refs.
func (c *Commands) InvestigateSync(ctx context.Context, doc config.File, req SyncInvestigationRequest) (SyncInvestigation, error) {
	var zero SyncInvestigation
	if c == nil || c.Service == nil || c.Local == nil {
		return zero, fmt.Errorf("local git inspector missing")
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	repoPath := strings.TrimSpace(req.RepoPath)
	if projectID == "" || branch == "" || repoPath == "" {
		return zero, badRequest("project_id, branch, and repo_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return zero, badRequest(err.Error())
	}
	abs, err := c.RequireMappedPath(ctx, doc, projectID, repoPath)
	if err != nil {
		return zero, err
	}
	result, err := c.Local.InspectSync(ctx, abs, branch)
	if err != nil {
		return zero, fmt.Errorf("investigate sync: %w", err)
	}
	return SyncInvestigation{
		ProjectID: projectID,
		Result:    result,
	}, nil
}
