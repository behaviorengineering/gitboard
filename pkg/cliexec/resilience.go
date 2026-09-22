package cliexec

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
)

const (
	retryMaxRetries   = 2
	retryBackoffMin   = 100 * time.Millisecond
	retryBackoffMax   = time.Second
	retryJitterFactor = 0.2
	breakerFailures   = 5
	breakerDelay      = 30 * time.Second
)

// ErrMissingDeadline is returned when Run is called with a context that has no deadline.
var ErrMissingDeadline = errors.New("cliexec: missing deadline")

type runnerResilience struct {
	retry    retrypolicy.RetryPolicy[[]byte]
	breakers sync.Map // map[string]circuitbreaker.CircuitBreaker[[]byte]
}

func newRunnerResilience() *runnerResilience {
	return &runnerResilience{
		retry: retrypolicy.NewBuilder[[]byte]().
			HandleIf(func(_ []byte, err error) bool { return isTransientCLIError(err) }).
			AbortOnErrors(context.Canceled).
			AbortIf(func(_ []byte, err error) bool {
				return err != nil && errors.Is(err, circuitbreaker.ErrOpen)
			}).
			WithBackoff(retryBackoffMin, retryBackoffMax).
			WithJitterFactor(retryJitterFactor).
			WithMaxRetries(retryMaxRetries).
			ReturnLastFailure().
			Build(),
	}
}

func (rr *runnerResilience) breakerFor(name string) circuitbreaker.CircuitBreaker[[]byte] {
	key := strings.TrimSpace(name)
	if key == "" {
		key = "unknown"
	}
	if v, ok := rr.breakers.Load(key); ok {
		cb, ok := v.(circuitbreaker.CircuitBreaker[[]byte])
		if !ok {
			return newCLIBreaker()
		}
		return cb
	}
	cb := newCLIBreaker()
	actual, _ := rr.breakers.LoadOrStore(key, cb)
	out, ok := actual.(circuitbreaker.CircuitBreaker[[]byte])
	if !ok {
		return cb
	}
	return out
}

func newCLIBreaker() circuitbreaker.CircuitBreaker[[]byte] {
	return circuitbreaker.NewBuilder[[]byte]().
		HandleIf(func(_ []byte, err error) bool { return isTransientCLIError(err) }).
		WithFailureThreshold(breakerFailures).
		WithDelay(breakerDelay).
		Build()
}

func (r *Runner) ensureResilience() *runnerResilience {
	if r.resilience == nil {
		r.resilience = newRunnerResilience()
	}
	return r.resilience
}

func (r *Runner) runWithResilience(ctx context.Context, name string, once func() ([]byte, error)) ([]byte, error) {
	if _, ok := ctx.Deadline(); !ok {
		return nil, ErrMissingDeadline
	}
	rr := r.ensureResilience()
	return failsafe.With(rr.breakerFor(name), rr.retry).
		WithContext(ctx).
		Get(once)
}

// isTransientCLIError reports whether err is worth retrying under load.
func isTransientCLIError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, circuitbreaker.ErrOpen) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, ErrMissingDeadline) {
		return false
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "timed out after"):
		return true
	case strings.Contains(msg, "process killed"):
		return true
	case strings.Contains(msg, "signal: killed"):
		return true
	default:
		return false
	}
}
