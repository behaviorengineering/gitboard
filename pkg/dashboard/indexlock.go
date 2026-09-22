package dashboard

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// Confirm codes returned to the UI for recoverable operator approval.
const ConfirmStaleIndexLock = "stale_index_lock"

// ConfirmRequiredError asks the UI to approve a follow-up (HTTP 409).
type ConfirmRequiredError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RepoPath   string `json:"repo_path"`
	LockPath   string `json:"lock_path,omitempty"`
	AgeSeconds int    `json:"age_seconds"`
}

func (e ConfirmRequiredError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return e.Code
	}
	return "confirmation required"
}

// IsConfirmRequired reports whether err needs a UI confirm-and-retry.
func IsConfirmRequired(err error) bool {
	var e ConfirmRequiredError
	return errors.As(err, &e)
}

func mapIndexLockErr(err error) error {
	if err == nil {
		return nil
	}
	var stale *localgit.StaleIndexLockError
	if errors.As(err, &stale) {
		ageSec := int(math.Round(stale.Age.Seconds()))
		if ageSec < 1 {
			ageSec = 1
		}
		return ConfirmRequiredError{
			Code:       ConfirmStaleIndexLock,
			Message:    stale.Error(),
			RepoPath:   stale.RepoPath,
			LockPath:   stale.LockPath,
			AgeSeconds: ageSec,
		}
	}
	if errors.Is(err, localgit.ErrIndexLockBusy) {
		return badRequestCause("", err)
	}
	return nil
}

func (c *Commands) ensureWritableIndex(ctx context.Context, repoPath string, clearStale bool) error {
	s, err := c.requireLocal()
	if err != nil {
		return err
	}
	err = s.Local.EnsureWritableIndex(ctx, repoPath, clearStale)
	if err == nil {
		return nil
	}
	if mapped := mapIndexLockErr(err); mapped != nil {
		return mapped
	}
	return fmt.Errorf("dashboard.ensureWritableIndex: %w", err)
}
