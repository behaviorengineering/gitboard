package dashboard

import (
	"context"
	"errors"
	"math"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// ConfirmCode names a follow-up the UI may ask the operator to approve.
type ConfirmCode string

// Confirm codes returned to the UI for recoverable operator approval.
const ConfirmStaleIndexLock ConfirmCode = "stale_index_lock"

// ConfirmRequiredError is the JSON payload the UI needs to ask the operator to
// approve a follow-up and retry. It travels inside a gberr.Error with
// gberr.CodeConfirmRequired; Cause keeps the localgit error for errors.Is.
type ConfirmRequiredError struct {
	Code       ConfirmCode `json:"code"`
	Message    string      `json:"message"`
	RepoPath   string      `json:"repo_path"`
	LockPath   string      `json:"lock_path,omitempty"`
	AgeSeconds int         `json:"age_seconds"`
	Cause      error       `json:"-"`
}

func (e ConfirmRequiredError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return string(e.Code)
	}
	return "confirmation required"
}

// Unwrap returns the underlying cause when present.
func (e ConfirmRequiredError) Unwrap() error {
	return e.Cause
}

// IsConfirmRequired reports whether err needs a UI confirm-and-retry.
func IsConfirmRequired(err error) bool {
	var e ConfirmRequiredError
	return errors.As(err, &e)
}

// mapIndexLockErr maps index.lock failures to a domain error: a stale lock
// asks for confirmation, a fresh lock is a conflict. Other errors return nil.
func mapIndexLockErr(op string, err error) error {
	if err == nil {
		return nil
	}
	var stale *localgit.StaleIndexLockError
	if errors.As(err, &stale) {
		ageSec := int(math.Round(stale.Age.Seconds()))
		if ageSec < 1 {
			ageSec = 1
		}
		confirm := ConfirmRequiredError{
			Code:       ConfirmStaleIndexLock,
			Message:    "a stale index.lock blocks git; approve removing it to continue",
			RepoPath:   stale.RepoPath,
			LockPath:   stale.LockPath,
			AgeSeconds: ageSec,
			Cause:      err,
		}
		return gberr.Wrap(confirm, gberr.CodeConfirmRequired, op, confirm.Message).
			With("lock_path", stale.LockPath)
	}
	if errors.Is(err, localgit.ErrIndexLockBusy) {
		return gberr.Wrap(err, gberr.CodeConflict, op, "index.lock is in use; wait for the other git process to finish")
	}
	return nil
}

// ensureWritableIndex runs m.EnsureWritableIndex and maps index.lock failures
// to a confirm-required or conflict error for the UI.
func ensureWritableIndex(ctx context.Context, m LocalMutator, repoPath string, clearStale bool) error {
	const op = "dashboard.ensureWritableIndex"
	err := m.EnsureWritableIndex(ctx, repoPath, clearStale)
	if err == nil {
		return nil
	}
	if mapped := mapIndexLockErr(op, err); mapped != nil {
		return mapped
	}
	return gitMutationErr(op, err)
}

// recoverIndexLock re-checks index.lock on disk after a git mutation failed.
// It returns retry=true when a stale lock was cleared with operator approval,
// lockErr when a lock explains the failure, and (false, nil) when no lock is
// present (or the probe itself fails) so the caller reports its original error.
func recoverIndexLock(ctx context.Context, m LocalMutator, repoPath string, clearStale bool) (retry bool, lockErr error) {
	const op = "dashboard.recoverIndexLock"
	mapped := mapIndexLockErr(op, m.EnsureWritableIndex(ctx, repoPath, false))
	if mapped == nil {
		return false, nil
	}
	if !clearStale || !IsConfirmRequired(mapped) {
		return false, mapped
	}
	if err := ensureWritableIndex(ctx, m, repoPath, true); err != nil {
		return false, err
	}
	return true, nil
}
