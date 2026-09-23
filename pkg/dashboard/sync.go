package dashboard

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
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

// InvestigateSync validates a mapped checkout, refreshes origin under the
// mutation lease, then compares refs without holding the lease.
func (c *Commands) InvestigateSync(ctx context.Context, doc config.File, req SyncInvestigationRequest) (SyncInvestigation, error) {
	var zero SyncInvestigation
	s, err := c.requireLocal()
	if err != nil {
		return zero, err
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	repoPath := strings.TrimSpace(req.RepoPath)
	if projectID == "" || branch == "" || repoPath == "" {
		return zero, badRequest("project_id, branch, and repo_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return zero, badRequestCause("", err)
	}
	abs, err := c.RequireMappedPath(ctx, doc, projectID, repoPath)
	if err != nil {
		return zero, err
	}
	common, err := localgit.ResolveCommonDir(ctx, s.Local.CommonGitDir, abs)
	if err != nil {
		return zero, badRequestCause("resolve common git dir", err)
	}

	if err := c.refreshOriginForInvestigate(ctx, s, abs, common, branch); err != nil {
		return zero, err
	}

	result, err := s.Local.CompareSync(ctx, abs, branch)
	if err != nil {
		return zero, fmt.Errorf("investigate sync: %w", err)
	}
	return SyncInvestigation{
		ProjectID: projectID,
		Result:    result,
	}, nil
}

func (c *Commands) refreshOriginForInvestigate(ctx context.Context, s *Service, abs, common, branch string) error {
	lease, err := s.Mutations.Acquire(ctx, common)
	if err != nil {
		if errors.Is(err, localgit.ErrMutationCanceled) {
			return err
		}
		return fmt.Errorf("mutation lock: %w", err)
	}
	defer lease.Release()

	// Fresh full fetch matches prior InspectSync (git fetch --prune origin).
	updated, err := s.Local.FetchOriginSmart(ctx, abs, time.Minute, true, s.OriginFetch, []string{branch})
	if err != nil {
		return fmt.Errorf("investigate fetch: %w", err)
	}
	if updated {
		lease.BumpEpoch()
		s.InvalidateOrigin(common)
	}
	return nil
}
