package localgit

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMutationCoordinatorSerializesSameDir(t *testing.T) {
	c := NewMutationCoordinator(nil)
	dir := t.TempDir()

	var active int32
	var maxActive int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lease, err := c.Acquire(context.Background(), dir)
			if err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			defer lease.Release()
			n := atomic.AddInt32(&active, 1)
			for {
				old := atomic.LoadInt32(&maxActive)
				if n <= old || atomic.CompareAndSwapInt32(&maxActive, old, n) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt32(&active, -1)
		}()
	}
	wg.Wait()
	if maxActive != 1 {
		t.Fatalf("max active = %d, want 1", maxActive)
	}
}

func TestMutationCoordinatorAllowsDifferentDirs(t *testing.T) {
	c := NewMutationCoordinator(nil)
	a := t.TempDir()
	b := t.TempDir()

	leaseA, err := c.Acquire(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	defer leaseA.Release()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	leaseB, err := c.Acquire(ctx, b)
	if err != nil {
		t.Fatalf("different dirs must not block: %v", err)
	}
	leaseB.Release()
}

func TestMutationCoordinatorCancelsWhileQueued(t *testing.T) {
	c := NewMutationCoordinator(nil)
	dir := t.TempDir()

	held, err := c.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err = c.Acquire(ctx, dir)
	if !errors.Is(err, ErrMutationCanceled) {
		t.Fatalf("want ErrMutationCanceled, got %v", err)
	}
}

func TestMutationCoordinatorNilAcquire(t *testing.T) {
	var c *MutationCoordinator
	_, err := c.Acquire(context.Background(), t.TempDir())
	if !errors.Is(err, ErrMutationCoordinatorRequired) {
		t.Fatalf("got %v, want ErrMutationCoordinatorRequired", err)
	}
}

func TestMutationCoordinatorNilContext(t *testing.T) {
	c := NewMutationCoordinator(nil)
	// Intentional nil: Acquire must fail closed rather than substitute Background.
	_, err := c.Acquire(nil, t.TempDir()) //nolint:staticcheck // SA1012: under test
	if !errors.Is(err, ErrMutationCanceled) {
		t.Fatalf("got %v, want ErrMutationCanceled", err)
	}
}

func TestMutationCoordinatorLeaseEpochAtGrant(t *testing.T) {
	c := NewMutationCoordinator(nil)
	dir := t.TempDir()

	held, err := c.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	held.BumpEpoch()
	held.BumpEpoch()

	started := make(chan struct{})
	errCh := make(chan error, 1)
	var gotEpoch uint64
	go func() {
		close(started)
		lease, err := c.Acquire(context.Background(), dir)
		if err != nil {
			errCh <- err
			return
		}
		gotEpoch = lease.Epoch()
		lease.Release()
		errCh <- nil
	}()
	<-started
	time.Sleep(30 * time.Millisecond)
	held.Release()
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	if gotEpoch != 2 {
		t.Fatalf("lease epoch at grant = %d, want 2 (bumps while queued)", gotEpoch)
	}
}

func TestMutationCoordinatorBumpEpoch(t *testing.T) {
	c := NewMutationCoordinator(nil)
	dir := t.TempDir()
	if c.Epoch(dir) != 0 {
		t.Fatal("expected zero epoch")
	}
	lease, err := c.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if lease.Epoch() != 0 {
		t.Fatalf("lease epoch = %d", lease.Epoch())
	}
	got := lease.BumpEpoch()
	lease.Release()
	if got != 1 || c.Epoch(dir) != 1 {
		t.Fatalf("epoch after bump = %d / %d", got, c.Epoch(dir))
	}
}

func TestMutationCoordinatorReleaseAfterErrorPath(t *testing.T) {
	c := NewMutationCoordinator(nil)
	dir := t.TempDir()

	lease, err := c.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	lease.Release()
	lease.Release() // idempotent

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lease2, err := c.Acquire(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	lease2.Release()
}

func TestResolveCommonDirRejectsEmpty(t *testing.T) {
	_, err := ResolveCommonDir(context.Background(), func(context.Context, string) (string, error) {
		return "", nil
	}, "/tmp/repo")
	if !errors.Is(err, ErrCommonDirRequired) {
		t.Fatalf("got %v", err)
	}
}

type blockingOSLocker struct {
	started chan struct{}
	release chan struct{}
}

func (b *blockingOSLocker) Lock(ctx context.Context, _ string) (func() error, error) {
	close(b.started)
	select {
	case <-b.release:
		return func() error { return nil }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestMutationCoordinatorOSLockCanceled(t *testing.T) {
	blocker := &blockingOSLocker{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	c := NewMutationCoordinator(blocker)
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := c.Acquire(ctx, dir)
		errCh <- err
	}()
	<-blocker.started
	cancel()
	err := <-errCh
	if !errors.Is(err, ErrMutationCanceled) {
		t.Fatalf("want canceled, got %v", err)
	}
}
