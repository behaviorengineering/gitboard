package localgit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultStaleIndexLockAge is how old an index.lock must be before gitboard
// treats it as leftover from a crashed or abandoned git process.
const DefaultStaleIndexLockAge = 30 * time.Second

// ErrStaleIndexLock means a leftover index.lock blocks writes; the board may
// ask the user to approve removing it.
var ErrStaleIndexLock = errors.New("stale index.lock")

// ErrIndexLockBusy means index.lock is recent; another git process may still
// be running, so the board must not delete it.
var ErrIndexLockBusy = errors.New("index.lock in use")

// StaleIndexLockError carries paths and age for a confirm-to-clear response.
type StaleIndexLockError struct {
	RepoPath string
	LockPath string
	Age      time.Duration
}

func (e *StaleIndexLockError) Error() string {
	if e == nil {
		return ErrStaleIndexLock.Error()
	}
	age := e.Age.Round(time.Second)
	if age < time.Second {
		age = time.Second
	}
	return fmt.Sprintf("%s at %s (age %s); another git process may have crashed", ErrStaleIndexLock.Error(), e.LockPath, age)
}

func (e *StaleIndexLockError) Is(target error) bool {
	return target == ErrStaleIndexLock
}

func (e *StaleIndexLockError) Unwrap() error {
	return ErrStaleIndexLock
}

// IsIndexLockError reports whether err looks like a git index.lock conflict.
func IsIndexLockError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrStaleIndexLock) || errors.Is(err, ErrIndexLockBusy) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "index.lock")
}

// EnsureWritableIndex checks for index.lock under repoPath.
// When clearStale is false and a stale lock exists, it returns *StaleIndexLockError.
// When clearStale is true, it removes a stale lock. A fresh lock always returns
// ErrIndexLockBusy (never deleted).
func (in *Inspector) EnsureWritableIndex(ctx context.Context, repoPath string, clearStale bool) error {
	return in.EnsureWritableIndexAge(ctx, repoPath, clearStale, DefaultStaleIndexLockAge)
}

// EnsureWritableIndexAge is EnsureWritableIndex with an explicit stale threshold.
func (in *Inspector) EnsureWritableIndexAge(ctx context.Context, repoPath string, clearStale bool, staleAfter time.Duration) error {
	if in == nil {
		return fmt.Errorf("inspector missing")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if staleAfter <= 0 {
		staleAfter = DefaultStaleIndexLockAge
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("repo path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	lockPath, err := in.resolveIndexLockPath(ctx, abs)
	if err != nil {
		return err
	}
	info, err := os.Stat(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat index.lock: %w", err)
	}
	age := time.Since(info.ModTime())
	if age < 0 {
		age = 0
	}
	if age < staleAfter {
		return fmt.Errorf("%w at %s (age %s); wait for the other git process to finish", ErrIndexLockBusy, lockPath, age.Round(time.Second))
	}
	if !clearStale {
		return &StaleIndexLockError{RepoPath: abs, LockPath: lockPath, Age: age}
	}
	if err := os.Remove(lockPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("remove stale index.lock: %w", err)
	}
	return nil
}

func (in *Inspector) resolveIndexLockPath(ctx context.Context, absRepo string) (string, error) {
	out, err := in.git(ctx, absRepo, "rev-parse", "--git-path", "index.lock")
	if err != nil {
		return "", fmt.Errorf("resolve index.lock path: %w", err)
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return "", fmt.Errorf("resolve index.lock path: empty")
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(absRepo, p)
	}
	return filepath.Clean(p), nil
}
