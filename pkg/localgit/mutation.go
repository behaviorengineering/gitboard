package localgit

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// ErrMutationCanceled means the request context ended while waiting for the
// per-repository mutation coordinator.
var ErrMutationCanceled = errors.New("mutation wait canceled")

// ErrCommonDirRequired means a mutating operation could not resolve a stable
// common git directory to coordinate on.
var ErrCommonDirRequired = errors.New("common git directory required")

// ErrMutationCoordinatorRequired means a mutating operation needed a coordinator
// and none was wired (nil *MutationCoordinator).
var ErrMutationCoordinatorRequired = errors.New("mutation coordinator required")

// OSLocker acquires a cross-process advisory lock for one common git directory.
// Unlock must be safe to call once; it should release even if the process exits
// abnormally when the implementation uses OS file locks.
type OSLocker interface {
	Lock(ctx context.Context, commonDir string) (unlock func() error, err error)
}

// MutationLease is exclusive ownership of one common git directory.
type MutationLease struct {
	commonDir string
	coord     *MutationCoordinator
	osUnlock  func() error
	waitMs    int64
	epoch     uint64
	released  bool
}

// CommonDir returns the coordinated common git directory.
func (l *MutationLease) CommonDir() string {
	if l == nil {
		return ""
	}
	return l.commonDir
}

// WaitMs is how long Acquire blocked before ownership was granted.
func (l *MutationLease) WaitMs() int64 {
	if l == nil {
		return 0
	}
	return l.waitMs
}

// Epoch is the mutation epoch observed when the lease was acquired.
func (l *MutationLease) Epoch() uint64 {
	if l == nil {
		return 0
	}
	return l.epoch
}

// BumpEpoch advances the common-directory epoch after a successful mutation.
func (l *MutationLease) BumpEpoch() uint64 {
	if l == nil || l.coord == nil {
		return 0
	}
	return l.coord.bumpEpoch(l.commonDir)
}

// Release drops the OS lock then the in-process lock. Safe to call once.
func (l *MutationLease) Release() {
	if l == nil || l.released {
		return
	}
	l.released = true
	if l.osUnlock != nil {
		_ = l.osUnlock()
		l.osUnlock = nil
	}
	if l.coord != nil {
		l.coord.releaseProcess(l.commonDir)
	}
}

// MutationCoordinator serializes mutating git work per common git directory.
// Unrelated repositories stay parallel. An optional OSLocker coordinates
// separate gitboard processes on unix and Windows; unsupported platforms skip
// the OS lock and keep in-process serialization only. Git's own index.lock
// still governs external git.
type MutationCoordinator struct {
	mu     sync.Mutex
	states map[string]*mutationDirState
	osLock OSLocker
}

type mutationDirState struct {
	// token is a buffered channel of size 1 used as a context-aware mutex.
	token chan struct{}
	epoch uint64
}

// NewMutationCoordinator builds a coordinator. A nil osLocker disables
// cross-process locking (in-process serialization still applies). Pass
// DefaultOSLocker() for the platform advisory lock (unix flock / Windows LockFileEx).
func NewMutationCoordinator(osLocker OSLocker) *MutationCoordinator {
	return &MutationCoordinator{
		states: map[string]*mutationDirState{},
		osLock: osLocker,
	}
}

// DefaultOSLocker returns the platform cross-process lock implementation
// (nil on GOOS without a lock backend).
func DefaultOSLocker() OSLocker {
	return defaultOSLocker()
}

// Epoch returns the current mutation epoch for commonDir (0 when unknown).
func (c *MutationCoordinator) Epoch(commonDir string) uint64 {
	if c == nil {
		return 0
	}
	commonDir = normalizeCommonDir(commonDir)
	if commonDir == "" {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	st := c.states[commonDir]
	if st == nil {
		return 0
	}
	return st.epoch
}

func (c *MutationCoordinator) bumpEpoch(commonDir string) uint64 {
	if c == nil {
		return 0
	}
	commonDir = normalizeCommonDir(commonDir)
	if commonDir == "" {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	st := c.stateLocked(commonDir)
	st.epoch++
	return st.epoch
}

// Acquire waits for exclusive ownership of commonDir until ctx is done.
// Lock order: resolve commonDir → in-process token → OS lock.
func (c *MutationCoordinator) Acquire(ctx context.Context, commonDir string) (*MutationLease, error) {
	if c == nil {
		return nil, ErrMutationCoordinatorRequired
	}
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil context", ErrMutationCanceled)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMutationCanceled, err)
	}
	commonDir = normalizeCommonDir(commonDir)
	if commonDir == "" {
		return nil, ErrCommonDirRequired
	}

	c.mu.Lock()
	st := c.stateLocked(commonDir)
	token := st.token
	c.mu.Unlock()

	start := time.Now()
	select {
	case token <- struct{}{}:
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %w", ErrMutationCanceled, ctx.Err())
	}
	waitMs := time.Since(start).Milliseconds()

	// Sample epoch after the token is held so the lease matches grant time,
	// not wait-start (concurrent bumpers may have advanced while queued).
	c.mu.Lock()
	epoch := c.stateLocked(commonDir).epoch
	osLock := c.osLock
	c.mu.Unlock()

	lease := &MutationLease{
		commonDir: commonDir,
		coord:     c,
		waitMs:    waitMs,
		epoch:     epoch,
	}

	if osLock != nil {
		unlock, err := osLock.Lock(ctx, commonDir)
		if err != nil {
			c.releaseProcess(commonDir)
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("%w: %w", ErrMutationCanceled, err)
			}
			return nil, fmt.Errorf("os mutation lock: %w", err)
		}
		lease.osUnlock = unlock
	}
	return lease, nil
}

func (c *MutationCoordinator) stateLocked(commonDir string) *mutationDirState {
	if c.states == nil {
		c.states = map[string]*mutationDirState{}
	}
	st, ok := c.states[commonDir]
	if !ok {
		st = &mutationDirState{token: make(chan struct{}, 1)}
		c.states[commonDir] = st
	}
	return st
}

func (c *MutationCoordinator) releaseProcess(commonDir string) {
	if c == nil {
		return
	}
	commonDir = normalizeCommonDir(commonDir)
	c.mu.Lock()
	st := c.states[commonDir]
	c.mu.Unlock()
	if st == nil {
		return
	}
	select {
	case <-st.token:
	default:
	}
}

func normalizeCommonDir(commonDir string) string {
	commonDir = filepath.Clean(commonDir)
	if commonDir == "." || commonDir == "" {
		return ""
	}
	return commonDir
}

// ResolveCommonDir returns a stable common git directory for mutations.
// Empty or failed resolution is an error (no silent path-key fallback).
func ResolveCommonDir(ctx context.Context, common func(context.Context, string) (string, error), repoPath string) (string, error) {
	repoPath = filepath.Clean(repoPath)
	if repoPath == "" || repoPath == "." {
		return "", ErrCommonDirRequired
	}
	if common == nil {
		return "", ErrCommonDirRequired
	}
	cd, err := common(ctx, repoPath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCommonDirRequired, err)
	}
	cd = normalizeCommonDir(cd)
	if cd == "" {
		return "", ErrCommonDirRequired
	}
	return cd, nil
}
