package dashboard

import (
	"context"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
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
func (c *Commands) InvestigateSync(ctx context.Context, doc boardconfig.File, req SyncInvestigationRequest) (SyncInvestigation, error) {
	const op = "dashboard.InvestigateSync"
	s, err := c.requireLocal()
	if err != nil {
		return SyncInvestigation{}, err
	}
	t, err := parseMutation(op, doc, mutationRequest{
		ProjectID: req.ProjectID, Branch: req.Branch, Path: req.RepoPath, PathField: "repo_path",
	})
	if err != nil {
		return SyncInvestigation{}, err
	}
	if err := s.requireMapped(ctx, op, doc, t, "repo_path"); err != nil {
		return SyncInvestigation{}, err
	}
	common, err := s.commonDirFor(ctx, op, t.path)
	if err != nil {
		return SyncInvestigation{}, err
	}
	if err := s.refreshOriginForInvestigate(ctx, t, common); err != nil {
		return SyncInvestigation{}, err
	}
	result, err := s.local.CompareSync(ctx, t.path, t.branch)
	if err != nil {
		return SyncInvestigation{}, gitMutationErr(op, err)
	}
	return SyncInvestigation{ProjectID: t.project.ID, Result: result}, nil
}

// investigateFetchTTL is the TTL recorded by the forced investigate fetch.
const investigateFetchTTL = time.Minute

// refreshOriginForInvestigate runs a fresh full fetch (git fetch --prune
// origin) under the lease before the read-only compare.
func (s *Service) refreshOriginForInvestigate(ctx context.Context, t mutationTarget, common string) error {
	var updated bool
	err := s.withLease(ctx, common, func(m LocalMutator) (bool, error) {
		policy := localgit.FetchPolicy{TTL: investigateFetchTTL, Fresh: true, Branches: []string{t.branch}}
		var err error
		updated, err = m.FetchOriginSmart(ctx, t.path, policy, s.originFetch)
		if err != nil {
			return false, gitMutationErr("dashboard.InvestigateSync fetch", err)
		}
		return updated, nil
	})
	if err != nil {
		return err
	}
	if updated {
		s.InvalidateOrigin(common)
	}
	return nil
}
