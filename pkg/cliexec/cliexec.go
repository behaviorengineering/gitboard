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

// DefaultMaxConcurrent caps how many CLI processes one runner starts at once.
// It is a safety net above the dashboard's own worker pools, so a refresh
// queues instead of exhausting process and file-descriptor limits.
const DefaultMaxConcurrent = 16

// New returns a runner with gitvalet's default timeout and DefaultMaxConcurrent.
func New() *Runner {
	r := gitexec.New()
	r.MaxConcurrent = DefaultMaxConcurrent
	return r
}
