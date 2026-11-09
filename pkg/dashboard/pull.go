package dashboard

import (
	"context"
	"errors"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// PullFFRequest identifies a local branch to fast-forward from origin.
type PullFFRequest struct {
	ProjectID      string `json:"project_id"`
	Branch         string `json:"branch"`
	RepoPath       string `json:"repo_path"`
	ClearIndexLock bool   `json:"clear_index_lock,omitempty"`
}

// PullFFResult is returned after a successful fast-forward, with the time
// each pull stage took.
type PullFFResult struct {
	OK     bool   `json:"ok"`
	Branch string `json:"branch"`
	Path   string `json:"path"`
	localgit.PullPhases
}

// PullFF validates the mapped checkout, then fast-forwards the branch from origin.
// Ahead/behind is decided inside localgit.Mutator.PullFFOnly after a fresh fetch.
// Already up to date is success: the board may still show a stale ↓N until refresh.
// Validation uses a path allowlist (cached scan) instead of a full attachLocal forge refresh.
func (c *Commands) PullFF(ctx context.Context, doc boardconfig.File, req PullFFRequest) (PullFFResult, error) {
	const op = "dashboard.PullFF"
	s, err := c.requireLocal()
	if err != nil {
		return PullFFResult{}, err
	}
	t, err := parseMutation(op, doc, mutationRequest{
		ProjectID: req.ProjectID, Branch: req.Branch, Path: req.RepoPath, PathField: "repo_path",
	})
	if err != nil {
		return PullFFResult{}, err
	}
	if err := s.requireMapped(ctx, op, doc, t, "repo_path"); err != nil {
		return PullFFResult{}, err
	}
	common, err := s.commonDirFor(ctx, op, t.path)
	if err != nil {
		return PullFFResult{}, err
	}
	var phases localgit.PullPhases
	err = s.withLease(ctx, common, func(m LocalMutator) (bool, error) {
		err := runWithIndexLock(ctx, m, t.path, req.ClearIndexLock, func() error {
			p, err := m.PullFFOnly(ctx, t.path, t.branch)
			phases.Add(p)
			// Idempotent with ensureBranchFFFromOrigin: a fresh fetch can show the
			// checkout is already caught up while the board still painted ↓N.
			if errors.Is(err, localgit.ErrUpToDate) {
				return nil
			}
			return err
		})
		if err != nil {
			return false, gitMutationErr(op, err)
		}
		return true, nil
	})
	if err != nil {
		return PullFFResult{}, err
	}
	s.afterMutation(common, t.project)
	return PullFFResult{OK: true, Branch: t.branch, Path: t.path, PullPhases: phases}, nil
}

// requireMapped refuses t.path unless it is a mapped checkout of t.project.
func (s *Service) requireMapped(ctx context.Context, op string, doc boardconfig.File, t mutationTarget, pathField string) error {
	allowed, err := s.mappedPathAllowlist(ctx, doc, t.project, t.path)
	if err != nil {
		return err
	}
	if !allowed {
		return invalid(op, pathField+" is not a mapped checkout for this project")
	}
	return nil
}

func repoPathAllowed(local *board.LocalStatus, abs string) bool {
	return repoPathAllowedCanon(local, abs, nil)
}

func repoPathAllowedCanon(local *board.LocalStatus, abs string, canon func(string) string) bool {
	if local == nil {
		return false
	}
	check := func(c string) bool {
		return sameCheckoutPath(c, abs, canon)
	}
	if check(local.Path) {
		return true
	}
	for _, wt := range local.Worktrees {
		if check(wt.Path) || check(wt.AppearancePath) {
			return true
		}
	}
	for _, app := range local.Appearances {
		if check(app.Path) {
			return true
		}
		for _, wt := range app.Worktrees {
			if check(wt.Path) || check(wt.AppearancePath) {
				return true
			}
		}
	}
	return false
}

// RequireMappedPath expands path and ensures it belongs to the project's mapped checkouts.
func (c *Commands) RequireMappedPath(ctx context.Context, doc boardconfig.File, projectID, repoPath string) (string, error) {
	const op = "dashboard.RequireMappedPath"
	s, err := c.requireLocal()
	if err != nil {
		return "", err
	}
	projectID = strings.TrimSpace(projectID)
	repoPath = strings.TrimSpace(repoPath)
	if projectID == "" || repoPath == "" {
		return "", invalid(op, "project_id and path are required")
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return "", invalid(op, "unknown project")
	}
	abs, err := localgit.ResolvePath(repoPath)
	if err != nil {
		return "", invalidCause(op, "path is not a usable path", err)
	}
	if err := s.requireMapped(ctx, op, doc, mutationTarget{project: p, path: abs}, "path"); err != nil {
		return "", err
	}
	return abs, nil
}
