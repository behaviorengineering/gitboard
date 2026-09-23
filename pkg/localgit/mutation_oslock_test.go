//go:build unix || windows

package localgit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFlockOSLockerCrossProcessExclusive(t *testing.T) {
	dir := t.TempDir()
	// Mimic a git common dir by ensuring the lock file parent exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	a := NewMutationCoordinator(DefaultOSLocker())
	b := NewMutationCoordinator(DefaultOSLocker())
	// Both use real OS locks via DefaultOSLocker.

	leaseA, err := a.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer leaseA.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, err = b.Acquire(ctx, dir)
	if !errors.Is(err, ErrMutationCanceled) {
		t.Fatalf("second coordinator should wait then cancel, got %v", err)
	}

	leaseA.Release()

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	leaseB, err := b.Acquire(ctx2, dir)
	if err != nil {
		t.Fatalf("after release: %v", err)
	}
	leaseB.Release()

	lockPath := filepath.Join(dir, mutationLockName)
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("expected lock file left for reclaim: %v", err)
	}
}
