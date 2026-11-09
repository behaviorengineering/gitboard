package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// logSink collects every line a recordLogger and its children write.
type logSink struct {
	mu    sync.Mutex
	lines []string
}

func (s *logSink) text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s.lines, "\n")
}

// recordLogger keeps log lines so tests can check the cause stayed on the
// server.
type recordLogger struct {
	sink   *logSink
	fields []any
}

func newRecordLogger() *recordLogger { return &recordLogger{sink: &logSink{}} }

func (l *recordLogger) Infof(format string, args ...any)  { l.add(format, args...) }
func (l *recordLogger) Errorf(format string, args ...any) { l.add(format, args...) }

func (l *recordLogger) With(fields ...any) observability.Logger {
	return &recordLogger{sink: l.sink, fields: append(append([]any(nil), l.fields...), fields...)}
}

func (l *recordLogger) add(format string, args ...any) {
	l.sink.mu.Lock()
	defer l.sink.mu.Unlock()
	l.sink.lines = append(l.sink.lines, fmt.Sprint(l.fields...)+" "+fmt.Sprintf(format, args...))
}

func TestStatusForEveryCode(t *testing.T) {
	want := map[gberr.Code]int{
		gberr.CodeInvalid:         http.StatusBadRequest,
		gberr.CodeNotFound:        http.StatusNotFound,
		gberr.CodeConflict:        http.StatusConflict,
		gberr.CodeConfirmRequired: http.StatusConflict,
		gberr.CodeCanceled:        http.StatusGatewayTimeout,
		gberr.CodeUnavailable:     http.StatusServiceUnavailable,
		gberr.CodeInternal:        http.StatusInternalServerError,
		gberr.Code("unknown"):     http.StatusInternalServerError,
	}
	for code, status := range want {
		if got := statusFor(code); got != status {
			t.Errorf("statusFor(%q) = %d, want %d", code, got, status)
		}
	}
}

// secretCause stands in for git stderr, a local path or config parse detail.
const secretCause = "fatal: /Users/me/private/repo/.git/index: permission denied"

func TestWriteErrorNeverSendsRawCause(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		status   int
		wantBody string
	}{
		{
			name:     "domain message",
			err:      gberr.Wrap(errors.New(secretCause), gberr.CodeConflict, "dashboard.PullFF", "working tree dirty"),
			status:   http.StatusConflict,
			wantBody: "working tree dirty",
		},
		{
			name:     "internal hides its message",
			err:      gberr.Wrap(errors.New(secretCause), gberr.CodeInternal, "x", "detail "+secretCause),
			status:   http.StatusInternalServerError,
			wantBody: "internal error",
		},
		{
			name:     "untyped error",
			err:      fmt.Errorf("pull: %s", secretCause),
			status:   http.StatusInternalServerError,
			wantBody: "internal error",
		},
		{
			name:     "context deadline",
			err:      fmt.Errorf("git fetch %s: %w", secretCause, context.DeadlineExceeded),
			status:   http.StatusGatewayTimeout,
			wantBody: "request cancelled or timed out",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := newRecordLogger()
			rec := httptest.NewRecorder()
			writeError(rec, log, "test.op", tc.err)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			body := rec.Body.String()
			if strings.Contains(body, "/Users/me") || strings.Contains(body, "permission denied") {
				t.Fatalf("client received the raw cause: %q", body)
			}
			if strings.TrimSpace(body) != tc.wantBody {
				t.Fatalf("body = %q, want %q", body, tc.wantBody)
			}
			if got := rec.Header().Get(errorCodeHeader); got != string(gberr.CodeOf(tc.err)) {
				t.Fatalf("%s = %q, want %q", errorCodeHeader, got, gberr.CodeOf(tc.err))
			}
			logged := log.sink.text()
			if !strings.Contains(logged, secretCause) || !strings.Contains(logged, "test.op") {
				t.Fatalf("server log must keep the cause and op: %q", logged)
			}
		})
	}
}

func TestWriteErrorConfirmRequiredSendsJSON(t *testing.T) {
	confirm := dashboard.ConfirmRequiredError{
		Code:       dashboard.ConfirmStaleIndexLock,
		Message:    "a stale index.lock blocks git; approve removing it to continue",
		RepoPath:   "/repo",
		LockPath:   "/repo/.git/index.lock",
		AgeSeconds: 90,
		Cause:      errors.New(secretCause),
	}
	err := gberr.Wrap(confirm, gberr.CodeConfirmRequired, "dashboard.PullFF", confirm.Message)
	rec := httptest.NewRecorder()
	writeError(rec, newRecordLogger(), "pull.ff", err)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("confirm response must be JSON, got Content-Type %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q", cc)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if got["code"] != string(dashboard.ConfirmStaleIndexLock) || got["lock_path"] != "/repo/.git/index.lock" {
		t.Fatalf("payload: %v", got)
	}
	if strings.Contains(rec.Body.String(), "permission denied") {
		t.Fatalf("confirm payload leaked the cause: %s", rec.Body.String())
	}
}
