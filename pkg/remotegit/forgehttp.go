package remotegit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// Forge REST API resilience. Retries cover 429 and 5xx responses and
// transient transport failures; 401, 403, 404 and other 4xx fail at once.
const (
	httpRetryMax     = 2
	httpBackoffMin   = 200 * time.Millisecond
	httpBackoffMax   = 2 * time.Second
	httpJitterFactor = 0.2
	httpBreakerFails = 5
	httpBreakerDelay = 30 * time.Second
	// httpMaxBody caps how much of one response is read.
	httpMaxBody = 8 << 20
	// httpDrainBody is how much of an error body is read before close, so
	// the connection can be reused.
	httpDrainBody = 64 << 10
)

// errBodyTooLarge is returned when a forge response exceeds httpMaxBody.
var errBodyTooLarge = errors.New("forge api response exceeds the size limit")

// httpStatusError is a non-2xx forge API response. It carries no body text:
// a server that echoes request headers must not leak credentials into logs.
type httpStatusError struct {
	forge  string
	status int
	// retryAfter is the server's Retry-After wait, 0 when absent.
	retryAfter time.Duration
	// pastDeadline is set when retryAfter would outlast the caller's
	// deadline, so the policy fails now instead of waiting to be cut off.
	pastDeadline bool
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("%s api: http %d %s", e.forge, e.status, http.StatusText(e.status))
}

// transient reports a status worth another attempt: 429 and 5xx.
func (e *httpStatusError) transient() bool {
	if e.pastDeadline {
		return false
	}
	return e.status == http.StatusTooManyRequests || e.status >= http.StatusInternalServerError
}

// forgeHTTP sends GET requests to one forge REST API under a retry policy and
// a circuit breaker shared by every call on the client, with one client span
// per request. It is the only place remotegit calls http.Client.Do.
type forgeHTTP struct {
	forge   string
	client  *http.Client
	retry   retrypolicy.RetryPolicy[[]byte]
	breaker circuitbreaker.CircuitBreaker[[]byte]
}

// newForgeHTTP builds the transport for forge. client must not set Timeout:
// the caller's context deadline bounds each request.
func newForgeHTTP(forge string, client *http.Client) *forgeHTTP {
	return &forgeHTTP{
		forge:  forge,
		client: client,
		retry: retrypolicy.NewBuilder[[]byte]().
			HandleIf(func(_ []byte, err error) bool { return isTransientHTTPError(err) }).
			AbortOnErrors(context.Canceled).
			AbortIf(func(_ []byte, err error) bool { return errors.Is(err, circuitbreaker.ErrOpen) }).
			WithBackoff(httpBackoffMin, httpBackoffMax).
			WithJitterFactor(httpJitterFactor).
			WithDelayFunc(retryAfterDelay).
			WithMaxRetries(httpRetryMax).
			ReturnLastFailure().
			Build(),
		breaker: circuitbreaker.NewBuilder[[]byte]().
			HandleIf(func(_ []byte, err error) bool { return isTransientHTTPError(err) }).
			WithFailureThreshold(httpBreakerFails).
			WithDelay(httpBreakerDelay).
			Build(),
	}
}

// get fetches endpoint and returns the 2xx body. authorise sets credentials
// on each attempt's request headers; it must not log them. ctx must carry a
// deadline: without one get returns cliexec.ErrMissingDeadline and sends
// nothing.
func (f *forgeHTTP) get(ctx context.Context, endpoint string, authorise func(*http.Request)) (body []byte, err error) {
	if ctx == nil {
		return nil, errors.New("forge api: context is required")
	}
	if _, ok := ctx.Deadline(); !ok {
		return nil, cliexec.ErrMissingDeadline
	}
	target, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("forge api: endpoint: %w", err)
	}
	ctx, span := otel.Tracer("gitboard").Start(ctx, f.forge+" GET",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.request.method", http.MethodGet),
			attribute.String("server.address", target.Host),
			attribute.String("url.path", target.Path),
		),
	)
	attempts := 0
	defer func() {
		span.SetAttributes(attribute.Int("gitboard.http.attempts", attempts))
		var se *httpStatusError
		if errors.As(err, &se) {
			span.SetAttributes(attribute.Int("http.response.status_code", se.status))
		}
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	body, err = failsafe.With(f.breaker, f.retry).
		WithContext(ctx).
		Get(func() ([]byte, error) {
			attempts++
			return f.getOnce(ctx, endpoint, authorise)
		})
	return body, err
}

// getOnce is one attempt: a fresh request, the body closed before return.
func (f *forgeHTTP) getOnce(ctx context.Context, endpoint string, authorise func(*http.Request)) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("forge api: request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if authorise != nil {
		authorise(req)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, httpDrainBody))
		return nil, newHTTPStatusError(ctx, f.forge, resp)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, httpMaxBody+1))
	if err != nil {
		return nil, fmt.Errorf("forge api: read body: %w", err)
	}
	if len(body) > httpMaxBody {
		return nil, errBodyTooLarge
	}
	return body, nil
}

// newHTTPStatusError reads the status and Retry-After of a failed response.
func newHTTPStatusError(ctx context.Context, forge string, resp *http.Response) *httpStatusError {
	e := &httpStatusError{forge: forge, status: resp.StatusCode}
	if e.status == http.StatusTooManyRequests || e.status == http.StatusServiceUnavailable {
		e.retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
	}
	if dl, ok := ctx.Deadline(); ok && e.retryAfter > 0 && paddedRetryAfter(e.retryAfter) >= time.Until(dl) {
		e.pastDeadline = true
	}
	return e
}

// parseRetryAfter reads a Retry-After header in delay-seconds or HTTP-date
// form; anything else is 0.
func parseRetryAfter(h string) time.Duration {
	h = strings.TrimSpace(h)
	if h == "" {
		return 0
	}
	if secs, err := strconv.Atoi(h); err == nil {
		if secs <= 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if at, err := http.ParseTime(h); err == nil {
		if d := time.Until(at); d > 0 {
			return d
		}
	}
	return 0
}

// paddedRetryAfter lifts d so the policy's jitter (down to 1-httpJitterFactor
// of the delay) never retries before the server's Retry-After.
func paddedRetryAfter(d time.Duration) time.Duration {
	return time.Duration(float64(d) / (1 - httpJitterFactor))
}

// retryAfterDelay waits for the server's Retry-After when it sent one, and
// otherwise leaves the exponential backoff in charge (-1).
func retryAfterDelay(exec failsafe.ExecutionAttempt[[]byte]) time.Duration {
	var se *httpStatusError
	if errors.As(exec.LastError(), &se) && se.retryAfter > 0 {
		return paddedRetryAfter(se.retryAfter)
	}
	return -1
}

// isTransientHTTPError classifies failures worth another attempt: 429 and
// 5xx, timeouts, refused or reset connections and a response cut short.
// Cancellation, an open breaker, a bad URL, TLS failures and an unknown host
// are permanent.
func isTransientHTTPError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, circuitbreaker.ErrOpen) || errors.Is(err, errBodyTooLarge) {
		return false
	}
	var se *httpStatusError
	if errors.As(err, &se) {
		return se.transient()
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsTimeout || dnsErr.IsTemporary
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var opErr *net.OpError
	return errors.As(err, &opErr) && (opErr.Op == "dial" || opErr.Op == "read")
}

// httpErr maps a forgeHTTP.get failure to a typed error for op. Msg names the
// failure class only; the status and cause stay in Fields and the chain.
func httpErr(err error, op, forge string) error {
	var se *httpStatusError
	switch {
	case errors.Is(err, cliexec.ErrMissingDeadline):
		return gberr.Wrap(err, gberr.CodeInvalid, op, forge+" request needs a deadline")
	case errors.Is(err, context.Canceled):
		return gberr.Wrap(err, gberr.CodeCanceled, op, forge+" request cancelled")
	case errors.Is(err, circuitbreaker.ErrOpen):
		return gberr.Wrap(err, gberr.CodeUnavailable, op, forge+" api is failing; requests are paused")
	case errors.As(err, &se):
		return gberr.Wrap(err, statusCode(se.status), op, statusMsg(forge, se.status)).
			With("http_status", strconv.Itoa(se.status))
	default:
		return gberr.Wrap(err, gberr.CodeUnavailable, op, forge+" api request failed")
	}
}

func statusCode(status int) gberr.Code {
	if status == http.StatusNotFound {
		return gberr.CodeNotFound
	}
	return gberr.CodeUnavailable
}

func statusMsg(forge string, status int) string {
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return forge + " rejected the credentials"
	case status == http.StatusNotFound:
		return forge + " resource not found"
	case status == http.StatusTooManyRequests:
		return forge + " rate limit reached"
	case status >= http.StatusInternalServerError:
		return forge + " api is unavailable"
	default:
		return forge + " api request failed"
	}
}
