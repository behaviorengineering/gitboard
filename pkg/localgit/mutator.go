package localgit

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
)

// ErrLeaseNotHeld means a Mutator was used without a live MutationLease
// (nil mutator, nil lease, or a lease that was already released).
var ErrLeaseNotHeld = errors.New("mutation lease not held")

// ErrOutsideLease means a mutation targeted a repository whose common git
// directory is not the one the lease holds.
var ErrOutsideLease = errors.New("repository outside mutation lease")

// Mutator changes one repository while its lease is held. It is the only way
// to fetch, pull, remove a checkout or clear an index.lock, so every local git
// write is serialised by MutationCoordinator. Get one from
// MutationLease.Mutator; every method checks that its path resolves to the
// lease's common git directory.
type Mutator struct {
	in    *Inspector
	lease *MutationLease
}

// Mutator returns a mutator bound to this lease; it fails once the lease is released.
func (l *MutationLease) Mutator(run cliexec.Exec) (*Mutator, error) {
	if l == nil || l.released.Load() {
		return nil, ErrLeaseNotHeld
	}
	if run == nil {
		return nil, errors.New("localgit.MutationLease.Mutator: Exec is required")
	}
	return &Mutator{in: &Inspector{Run: run}, lease: l}, nil
}

// bind expands repoPath and checks that it belongs to the held lease.
// It returns the absolute path and the resolved common git directory.
func (m *Mutator) bind(ctx context.Context, repoPath string) (abs, common string, err error) {
	if m == nil || m.in == nil || m.lease == nil || m.lease.released.Load() {
		return "", "", ErrLeaseNotHeld
	}
	abs, err = ExpandPath(repoPath)
	if err != nil {
		return "", "", fmt.Errorf("repo path: %w", err)
	}
	common, err = ResolveCommonDir(ctx, m.in.CommonGitDir, abs)
	if err != nil {
		return "", "", err
	}
	if !sameDir(common, m.lease.commonDir) {
		return "", "", fmt.Errorf("%w: %s uses %s, lease holds %s", ErrOutsideLease, abs, common, m.lease.commonDir)
	}
	return abs, common, nil
}

// sameDir compares two cleaned directories, falling back to resolved symlinks.
func sameDir(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if a == b {
		return true
	}
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	return errA == nil && errB == nil && ra == rb
}
