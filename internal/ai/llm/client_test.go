package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestEnabled(t *testing.T) {
	var nilC *Client
	if nilC.Enabled() {
		t.Fatal("nil enabled")
	}
	if (&Client{}).Enabled() {
		t.Fatal("empty base enabled")
	}
	if !(&Client{BaseURL: "http://x"}).Enabled() {
		t.Fatal("want enabled")
	}
}

func TestChat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "m",
			"choices": []map[string]any{
				{"message": map[string]string{"content": "hi"}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, Model: "fallback", HTTP: srv.Client(), resilience: newClientResilience()}
	content, model, err := c.Chat(testCtx(t), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if content != "hi" || model != "m" {
		t.Fatalf("content=%q model=%q", content, model)
	}
}

func TestChatRequiresDeadline(t *testing.T) {
	c := &Client{BaseURL: "http://example.invalid", resilience: newClientResilience()}
	_, _, err := c.Chat(context.Background(), "sys", "user")
	if !errors.Is(err, errMissingDeadline) {
		t.Fatalf("got %v want errMissingDeadline", err)
	}
}

func TestIsTransientLLMError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"canceled", fmt.Errorf("llm chat: %w", context.Canceled), false},
		{"missing deadline", errMissingDeadline, false},
		{"429", httpStatusError("429 Too Many Requests", http.StatusTooManyRequests, ""), true},
		{"503", httpStatusError("503 Service Unavailable", http.StatusServiceUnavailable, ""), true},
		{"400", httpStatusError("400 Bad Request", http.StatusBadRequest, ""), false},
		{"refused", fmt.Errorf("llm chat: %w", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}), true},
		{"reset", fmt.Errorf("llm read: %w", syscall.ECONNRESET), true},
		{"unknown host", fmt.Errorf("llm chat: %w", &net.OpError{Op: "dial", Err: &net.DNSError{Err: "no such host", IsNotFound: true}}), false},
		{"bad scheme", fmt.Errorf("llm chat: %w", &url.Error{Op: "Post", URL: "ftp://x", Err: errors.New("unsupported protocol scheme")}), false},
		// The old classifier retried anything whose text started "llm chat:".
		{"message prefix only", errors.New("llm chat: x509: certificate signed by unknown authority"), false},
	}
	for _, tc := range cases {
		if got := isTransientLLMError(tc.err); got != tc.want {
			t.Errorf("%s: isTransientLLMError(%v) = %v, want %v", tc.name, tc.err, got, tc.want)
		}
	}
}

func TestChatRetriesTransientThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "m",
			"choices": []map[string]any{
				{"message": map[string]string{"content": "ok"}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, Model: "m", HTTP: srv.Client(), resilience: newClientResilience()}
	content, _, err := c.Chat(testCtx(t), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if content != "ok" {
		t.Fatalf("content=%q", content)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls=%d want 3", calls.Load())
	}
}

func TestChatDoesNotRetryClientError(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, Model: "m", HTTP: srv.Client(), resilience: newClientResilience()}
	_, _, err := c.Chat(testCtx(t), "sys", "user")
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1", calls.Load())
	}
}
