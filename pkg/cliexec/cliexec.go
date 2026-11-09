// Package cliexec is public API: the injectable CLI runner that forge clients
// (pkg/remotegit) and local git (pkg/localgit) call. It re-exports the shared
// gitvalet executor; retry, circuit breaking and the per-call timeout live in
// gitvalet/pkg/gitexec.
package cliexec

import (
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
