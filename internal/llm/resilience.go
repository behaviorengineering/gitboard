package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
)

const (
	llmRetryMax     = 2
	llmBackoffMin   = 100 * time.Millisecond
	llmBackoffMax   = time.Second
	llmJitterFactor = 0.2
	llmBreakerFails = 5
	llmBreakerDelay = 30 * time.Second
)

// ErrMissingDeadline is returned when Chat is called without a ctx deadline.
var ErrMissingDeadline = errors.New("llm: missing deadline")

type chatResult struct {
	content string
	model   string
}

// transientHTTPError marks 429/5xx responses as retryable for failsafe HandleIf.
type transientHTTPError struct {
	status int
	msg    string
}

func (e *transientHTTPError) Error() string {
	if e == nil {
		return "llm: transient http error"
	}
	return e.msg
}

type clientResilience struct {
	retry   retrypolicy.RetryPolicy[chatResult]
	breaker circuitbreaker.CircuitBreaker[chatResult]
}

func newClientResilience() *clientResilience {
	return &clientResilience{
		retry: retrypolicy.NewBuilder[chatResult]().
			HandleIf(func(_ chatResult, err error) bool { return isTransientLLMError(err) }).
			AbortOnErrors(context.Canceled).
			AbortIf(func(_ chatResult, err error) bool {
				return err != nil && errors.Is(err, circuitbreaker.ErrOpen)
			}).
			WithBackoff(llmBackoffMin, llmBackoffMax).
			WithJitterFactor(llmJitterFactor).
			WithMaxRetries(llmRetryMax).
			ReturnLastFailure().
			Build(),
		breaker: circuitbreaker.NewBuilder[chatResult]().
			HandleIf(func(_ chatResult, err error) bool { return isTransientLLMError(err) }).
			WithFailureThreshold(llmBreakerFails).
			WithDelay(llmBreakerDelay).
			Build(),
	}
}

func (c *Client) ensureResilience() *clientResilience {
	c.resilienceMu.Lock()
	defer c.resilienceMu.Unlock()
	if c.resilience == nil {
		c.resilience = newClientResilience()
	}
	return c.resilience
}

func (c *Client) runChatWithResilience(ctx context.Context, once func() (chatResult, error)) (chatResult, error) {
	if _, ok := ctx.Deadline(); !ok {
		return chatResult{}, ErrMissingDeadline
	}
	rr := c.ensureResilience()
	return failsafe.With(rr.breaker, rr.retry).
		WithContext(ctx).
		Get(once)
}

func isTransientLLMError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, ErrMissingDeadline) {
		return false
	}
	if errors.Is(err, circuitbreaker.ErrOpen) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var th *transientHTTPError
	if errors.As(err, &th) {
		return th.status == http.StatusTooManyRequests || th.status >= 500
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	return strings.Contains(err.Error(), "llm chat:")
}

func httpStatusError(statusLine string, statusCode int, body string) error {
	msg := fmt.Sprintf("llm %s: %s", statusLine, strings.TrimSpace(body))
	if statusCode == http.StatusTooManyRequests || statusCode >= 500 {
		return &transientHTTPError{status: statusCode, msg: msg}
	}
	return errors.New(msg)
}
