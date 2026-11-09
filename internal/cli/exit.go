package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// Process exit codes. The agent guide and AGENTS.md list the same table;
// change all three together.
const (
	exitOK          = 0
	exitInternal    = 1
	exitUsage       = 2
	exitConflict    = 3
	exitNotFound    = 4
	exitUnavailable = 5
	exitCanceled    = 6
)

// usageError marks a bad command line: an unknown command or flag, a
// missing argument, or an interactive command run without a terminal.
type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }

func (e usageError) Unwrap() error { return e.err }

// usage marks err as a command-line mistake (exit code 2).
func usage(err error) error {
	if err == nil {
		return nil
	}
	return usageError{err: err}
}

// noArgs is cobra.NoArgs with the error marked as usage.
func noArgs(cmd *cobra.Command, args []string) error {
	return usage(cobra.NoArgs(cmd, args))
}

// exitCode maps a command error to the process exit code: usage errors and
// gberr.CodeInvalid exit 2, the other domain codes have their own value, and
// an error with no code is internal.
func exitCode(err error) int {
	if err == nil {
		return exitOK
	}
	var u usageError
	if errors.As(err, &u) {
		return exitUsage
	}
	switch gberr.CodeOf(err) {
	case gberr.CodeInvalid:
		return exitUsage
	case gberr.CodeConflict, gberr.CodeConfirmRequired:
		return exitConflict
	case gberr.CodeNotFound:
		return exitNotFound
	case gberr.CodeUnavailable:
		return exitUnavailable
	case gberr.CodeCanceled:
		return exitCanceled
	case gberr.CodeInternal:
		return exitInternal
	}
	return exitInternal
}
