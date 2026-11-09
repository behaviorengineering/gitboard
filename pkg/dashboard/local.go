package dashboard

import (
	"context"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// LocalScanner discovers checkouts under scan roots and reads their status.
type LocalScanner interface {
	ScanRoots(ctx context.Context, roots []string) localgit.Discovery
	InspectPath(ctx context.Context, path string) localgit.Status
	// OriginSync compares local branches to origin for the repository at
	// dir; every worktree of that repository shares the result.
	OriginSync(ctx context.Context, dir string) (localgit.RepoSync, error)
	CommonGitDir(ctx context.Context, repoPath string) (string, error)
	ListLocalHeads(ctx context.Context, repoPath string) ([]string, error)
	OriginRemote(ctx context.Context, dir string) (string, error)
}

// LocalComparer answers branch and path questions for prune hints, sync
// investigation and mapped-path checks. It never writes.
type LocalComparer interface {
	ContentOnDefault(ctx context.Context, repoPath, branch, defaultBranch string) (ok bool, reason string, err error)
	CompareSync(ctx context.Context, repoPath, branch string) (localgit.SyncInspection, error)
	CanonicalCheckoutPath(ctx context.Context, path string) string
}

// LocalMutator changes one repository while its mutation lease is held.
// *localgit.Mutator implements it.
type LocalMutator interface {
	FetchOriginSmart(ctx context.Context, repoPath string, policy localgit.FetchPolicy, cache *localgit.OriginFetchCache) (updated bool, err error)
	PullFFOnly(ctx context.Context, repoPath, branch string) (localgit.PullPhases, error)
	RemoveSafeCheckout(ctx context.Context, worktreePath, branch, defaultBranch string) error
	EnsureWritableIndex(ctx context.Context, repoPath string, clearStale bool) error
}

// LocalGit is the local checkout surface Collect and Commands use: reads, plus
// a mutator that only a held lease can hand out, so every write runs under
// the per-repository MutationCoordinator. NewLocalGit wires the git CLI.
type LocalGit interface {
	LocalScanner
	LocalComparer
	// Mutator returns a writer bound to lease; it fails once the lease is released.
	Mutator(lease *localgit.MutationLease) (LocalMutator, error)
}

// gitLocal adapts the git CLI inspector to LocalGit.
type gitLocal struct {
	*localgit.Inspector
}

// NewLocalGit returns the git CLI LocalGit: a read-only localgit.Inspector
// plus lease-bound localgit.Mutator writers. Panics if run is nil.
func NewLocalGit(run cliexec.Exec) LocalGit {
	return gitLocal{Inspector: localgit.NewInspector(run)}
}

// Mutator returns the lease's localgit.Mutator.
func (g gitLocal) Mutator(lease *localgit.MutationLease) (LocalMutator, error) {
	m, err := lease.Mutator(g.Run)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Ensure the git CLI types satisfy the dashboard interfaces.
var (
	_ LocalGit     = gitLocal{}
	_ LocalMutator = (*localgit.Mutator)(nil)
)
