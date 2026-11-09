package dashboard

import (
	"context"
	"errors"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// mutationRequest is the shared shape of a pull, prune or sync request.
type mutationRequest struct {
	ProjectID string
	Branch    string
	Path      string
	// PathField names Path in client messages ("repo_path", "worktree_path").
	PathField string
}

// mutationTarget is a validated mutation request: a tracked project, a valid
// branch and a resolved checkout path.
type mutationTarget struct {
	project boardconfig.Project
	branch  string
	path    string
}

// parseMutation trims and validates req: every field required, a safe
// branch name, a tracked project and a usable path (resolved with
// localgit.ResolvePath). It does not check that the path is mapped.
func parseMutation(op string, doc boardconfig.File, req mutationRequest) (mutationTarget, error) {
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	path := strings.TrimSpace(req.Path)
	if projectID == "" || branch == "" || path == "" {
		return mutationTarget{}, invalid(op, "project_id, branch, and "+req.PathField+" are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return mutationTarget{}, invalidCause(op, localgit.ErrInvalidBranch.Error(), err)
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return mutationTarget{}, invalid(op, "unknown project")
	}
	abs, err := localgit.ResolvePath(path)
	if err != nil {
		return mutationTarget{}, invalidCause(op, req.PathField+" is not a usable path", err)
	}
	return mutationTarget{project: p, branch: branch, path: abs}, nil
}

// commonDirFor resolves the common git directory a mutation of path leases.
func (s *Service) commonDirFor(ctx context.Context, op, path string) (string, error) {
	common, err := localgit.ResolveCommonDir(ctx, s.local.CommonGitDir, path)
	if err != nil {
		return "", invalidCause(op, "cannot resolve the common git directory", err)
	}
	return common, nil
}

// withLease takes the mutation lease for common, runs fn with its writer and
// releases the lease. When fn reports a change, the repository's mutation
// epoch is bumped so concurrent board reads drop stale origin sync.
func (s *Service) withLease(ctx context.Context, common string, fn func(LocalMutator) (changed bool, err error)) (err error) {
	lease, m, err := s.leaseMutator(ctx, common)
	if err != nil {
		return err
	}
	defer releaseLease(lease, &err)
	changed, err := fn(m)
	if changed {
		lease.BumpEpoch()
	}
	return err
}

// leaseMutator takes the mutation lease for common and returns its writer.
// The caller must Release the lease. A cancelled wait is gberr.CodeCanceled
// and still matches localgit.ErrMutationCanceled.
func (s *Service) leaseMutator(ctx context.Context, common string) (*localgit.MutationLease, LocalMutator, error) {
	const op = "dashboard.leaseMutator"
	lease, err := s.mutations.Acquire(ctx, common)
	if err != nil {
		if errors.Is(err, localgit.ErrMutationCanceled) {
			return nil, nil, gberr.Wrap(err, gberr.CodeCanceled, op, "request ended while waiting for the repository lock").
				With("common_dir", common)
		}
		return nil, nil, gberr.Wrap(err, gberr.CodeUnavailable, op, "cannot take the repository lock").
			With("common_dir", common)
	}
	m, err := s.local.Mutator(lease)
	if err != nil {
		err = gberr.Wrap(err, gberr.CodeInternal, op, "cannot bind the repository writer").
			With("common_dir", common)
		releaseLease(lease, &err)
		return nil, nil, err
	}
	return lease, m, nil
}

// afterMutation drops the origin-fetch TTL for common and the forge cache
// for p, so the next board read sees the change.
func (s *Service) afterMutation(common string, p boardconfig.Project) {
	s.InvalidateOrigin(common)
	s.InvalidateProject(p.Host, p.Path)
}

// runWithIndexLock runs a git mutation guarded by index.lock checks: a lock
// found before fn is reported (or cleared when clearStale), and after a
// failed fn the lock is re-checked on disk and fn retried once when the
// operator approved clearing a stale lock.
func runWithIndexLock(ctx context.Context, m LocalMutator, path string, clearStale bool, fn func() error) error {
	if err := ensureWritableIndex(ctx, m, path, clearStale); err != nil {
		return err
	}
	err := fn()
	if err == nil {
		return nil
	}
	retry, lockErr := recoverIndexLock(ctx, m, path, clearStale)
	if lockErr != nil {
		return lockErr
	}
	if retry {
		return fn()
	}
	return err
}
