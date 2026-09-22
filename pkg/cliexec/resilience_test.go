package cliexec

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/failsafe-go/failsafe-go/circuitbreaker"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestRunRequiresDeadline(t *testing.T) {
	r := New()
	_, err := r.Run(context.Background(), "true")
	if !errors.Is(err, ErrMissingDeadline) {
		t.Fatalf("got %v want ErrMissingDeadline", err)
	}
}

func TestRunRetriesTransientThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	r := &Runner{
		Timeout:    time.Second,
		resilience: newRunnerResilience(),
		attempt: func(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
			n := calls.Add(1)
			if n < 3 {
				return nil, errors.New("glab ci list: canceled: signal: killed")
			}
			return []byte("ok"), nil
		},
	}
	out, err := r.Run(testCtx(t), "glab", "ci", "list")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if string(out) != "ok" {
		t.Fatalf("out=%q", out)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls=%d want 3", calls.Load())
	}
}

func TestRunDoesNotRetryPermanentExit(t *testing.T) {
	var calls atomic.Int32
	r := &Runner{
		Timeout:    time.Second,
		resilience: newRunnerResilience(),
		attempt: func(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
			calls.Add(1)
			return nil, errors.New("glab: exit status 1: unknown flag")
		},
	}
	_, err := r.Run(testCtx(t), "glab", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1 (no retry on permanent failure)", calls.Load())
	}
}

func TestRunBreakerOpensAndFailsFast(t *testing.T) {
	var calls atomic.Int32
	r := &Runner{
		Timeout:    time.Second,
		resilience: newRunnerResilience(),
		attempt: func(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
			calls.Add(1)
			return nil, errors.New("gh api: timed out after 1s: signal: killed")
		},
	}
	for i := 0; i < 10; i++ {
		_, _ = r.Run(testCtx(t), "gh", "api")
		if r.ensureResilience().breakerFor("gh").IsOpen() {
			break
		}
	}
	if !r.ensureResilience().breakerFor("gh").IsOpen() {
		t.Fatal("expected breaker open")
	}
	before := calls.Load()
	_, err := r.Run(testCtx(t), "gh", "api")
	if !errors.Is(err, circuitbreaker.ErrOpen) && !strings.Contains(err.Error(), "circuit breaker open") {
		t.Fatalf("want ErrOpen, got %v", err)
	}
	if calls.Load() != before {
		t.Fatalf("breaker open still invoked attempt: before=%d after=%d", before, calls.Load())
	}
}

func TestIsTransientCLIError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{context.Canceled, false},
		{circuitbreaker.ErrOpen, false},
		{ErrMissingDeadline, false},
		{context.DeadlineExceeded, true},
		{errors.New("cmd: timed out after 45s: signal: killed"), true},
		{errors.New("cmd: process killed (timeout, cancel, or OOM): signal: killed"), true},
		{errors.New("glab: exit status 1"), false},
	}
	for _, tc := range cases {
		if got := isTransientCLIError(tc.err); got != tc.want {
			t.Fatalf("isTransientCLIError(%v)=%v want %v", tc.err, got, tc.want)
		}
	}
}
