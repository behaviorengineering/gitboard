package dashboard

import (
	"context"
	"errors"
	"fmt"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// invalid reports bad caller input (gberr.CodeInvalid). msg is client-safe.
func invalid(op, msg string) error {
	return gberr.New(gberr.CodeInvalid, op, msg)
}

// invalidCause is invalid with the cause kept for errors.Is and logs.
func invalidCause(op, msg string, cause error) error {
	if cause == nil {
		return invalid(op, msg)
	}
	return gberr.Wrap(cause, gberr.CodeInvalid, op, msg)
}

// conflict reports current state that refuses the action (gberr.CodeConflict).
func conflict(op, msg string) error {
	return gberr.New(gberr.CodeConflict, op, msg)
}

// gitMutationErr maps a localgit mutation failure to a domain error with a
// client-safe message. The cause (paths, git stderr) stays in the chain.
func gitMutationErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := gberr.As(err); ok {
		return err
	}
	// Cancellation is checked first: a timed-out rev-parse also carries
	// ErrMissingBranch, and the operator must not be told the branch is gone.
	switch {
	case errors.Is(err, localgit.ErrMutationCanceled),
		errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return gberr.Wrap(err, gberr.CodeCanceled, op, "request ended before git finished")
	case errors.Is(err, localgit.ErrInvalidBranch):
		return gberr.Wrap(err, gberr.CodeInvalid, op, localgit.ErrInvalidBranch.Error())
	case errors.Is(err, localgit.ErrMissingBranch):
		return gberr.Wrap(err, gberr.CodeInvalid, op, localgit.ErrMissingBranch.Error())
	case errors.Is(err, localgit.ErrDirtyTree):
		return gberr.Wrap(err, gberr.CodeConflict, op, localgit.ErrDirtyTree.Error())
	case errors.Is(err, localgit.ErrDiverged):
		return gberr.Wrap(err, gberr.CodeConflict, op, localgit.ErrDiverged.Error())
	default:
		return gberr.Wrap(err, gberr.CodeUnavailable, op, "git command failed")
	}
}

// releaseLease releases lease and joins a failed OS unlock into *errp, so a
// caller with a named error return learns that another gitboard process may
// stay blocked.
func releaseLease(lease *localgit.MutationLease, errp *error) {
	if rerr := lease.Release(); rerr != nil {
		*errp = errors.Join(*errp, fmt.Errorf("release mutation lease: %w", rerr))
	}
}

// appendErrText adds err to a display-only status error so a later failure
// in the same row is shown next to the first one instead of dropped.
func appendErrText(existing string, err error) string {
	if err == nil {
		return existing
	}
	if existing == "" {
		return err.Error()
	}
	return existing + "; " + err.Error()
}
