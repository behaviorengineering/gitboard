//go:build windows

package localgit

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

func defaultOSLocker() OSLocker {
	return LockFileOSLocker{}
}

// LockFileOSLocker uses LockFileEx; the kernel drops the lock when the handle closes.
type LockFileOSLocker struct{}

// Lock opens commonDir/gitboard.mutation.lock and waits for an exclusive LockFileEx.
func (LockFileOSLocker) Lock(ctx context.Context, commonDir string) (func() error, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil context", ErrMutationCanceled)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f, err := openMutationLockFile(commonDir)
	if err != nil {
		return nil, err
	}

	const (
		lockFlags = windows.LOCKFILE_EXCLUSIVE_LOCK | windows.LOCKFILE_FAIL_IMMEDIATELY
		lockBytes = 1
	)
	for {
		var overlapped windows.Overlapped
		err := windows.LockFileEx(windows.Handle(f.Fd()), lockFlags, 0, lockBytes, 0, &overlapped)
		if err == nil {
			return func() error {
				var unlockOverlapped windows.Overlapped
				_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, lockBytes, 0, &unlockOverlapped)
				return f.Close()
			}, nil
		}
		if err != windows.ERROR_LOCK_VIOLATION && err != windows.ERROR_IO_PENDING {
			_ = f.Close()
			return nil, fmt.Errorf("lockfile mutation lock: %w", err)
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = f.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
