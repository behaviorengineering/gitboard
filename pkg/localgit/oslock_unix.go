//go:build unix

package localgit

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

func defaultOSLocker() OSLocker {
	return FlockOSLocker{}
}

// FlockOSLocker uses an exclusive flock that the kernel drops when the process exits.
type FlockOSLocker struct{}

// Lock opens commonDir/gitboard.mutation.lock and waits for an exclusive flock.
func (FlockOSLocker) Lock(ctx context.Context, commonDir string) (func() error, error) {
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

	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			return func() error {
				_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
				return f.Close()
			}, nil
		}
		if err != unix.EWOULDBLOCK && err != unix.EAGAIN {
			_ = f.Close()
			return nil, fmt.Errorf("flock mutation lock: %w", err)
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
