// Package cliexec re-exports the shared gitvalet git executor.
package cliexec

import (
	"time"

	"github.com/behaviorengineering/gitvalet/pkg/gitexec"
)

// Exec is the injectable CLI surface used by forge and local git callers.
type Exec = gitexec.Exec

// Runner executes external CLIs with a timeout and failsafe-go resilience.
type Runner = gitexec.Runner

// ErrMissingDeadline is returned when Run is called without a context deadline.
var ErrMissingDeadline = gitexec.ErrMissingDeadline

// New returns a runner with a sensible default timeout.
func New() *Runner {
	return gitexec.New()
}

// SetTimeout is a test helper to adjust the per-invocation cap.
func SetTimeout(r *Runner, d time.Duration) {
	if r == nil {
		return
	}
	r.Timeout = d
}
