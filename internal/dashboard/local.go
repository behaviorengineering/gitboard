package dashboard

import (
	"context"
	"time"

	"github.com/behaviorengineering/gitboard/internal/localgit"
)

// LocalGit is the local checkout surface used by Collect and Commands (injectable for tests).
type LocalGit interface {
	ScanRoots(ctx context.Context, roots []string) localgit.Discovery
	InspectPath(ctx context.Context, path string) localgit.Status
	EnrichOriginSync(ctx context.Context, st *localgit.Status)
	CommonGitDir(ctx context.Context, repoPath string) (string, error)
	FetchOriginCached(ctx context.Context, repoPath string, ttl time.Duration, fresh bool, cache *localgit.OriginFetchCache) error
	OriginRemote(ctx context.Context, dir string) (string, error)
	PullFFOnly(ctx context.Context, repoPath, branch string) error
	RemoveSafeCheckout(ctx context.Context, worktreePath, branch, defaultBranch string) error
	ContentOnDefault(ctx context.Context, repoPath, branch, defaultBranch string) (ok bool, reason string, err error)
}

// Ensure *localgit.Inspector satisfies LocalGit.
var _ LocalGit = (*localgit.Inspector)(nil)
