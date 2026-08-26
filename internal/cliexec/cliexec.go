package cliexec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Exec is the injectable CLI surface used by forge and local git callers.
type Exec interface {
	LookPath(name string) (string, error)
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	RunJSON(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Runner executes external CLIs with a timeout.
type Runner struct {
	Timeout time.Duration
}

// New returns a runner with a sensible default timeout.
func New() *Runner {
	return &Runner{Timeout: 45 * time.Second}
}

// LookPath reports whether name is on PATH.
func (r *Runner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// LookPath reports whether name is on PATH (package helper for callers without a Runner).
func LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// RunJSON runs argv and expects JSON on stdout.
func (r *Runner) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := r.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return []byte("[]"), nil
	}
	return []byte(trimmed), nil
}

// Run executes a command and returns combined relevant output.
func (r *Runner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if r == nil {
		r = New()
	}
	runCtx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), msg)
	}
	return stdout.Bytes(), nil
}
