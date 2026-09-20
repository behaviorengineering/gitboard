package cliexec

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestClarifyProcessErrorTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)
	got := clarifyProcessError(ctx, 45*time.Second, errors.New("signal: killed"))
	if got == nil || !strings.Contains(got.Error(), "timed out after") {
		t.Fatalf("got %v", got)
	}
}

func TestClarifyProcessErrorCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := clarifyProcessError(ctx, 45*time.Second, errors.New("signal: killed"))
	if got == nil || !strings.Contains(got.Error(), "canceled") {
		t.Fatalf("got %v", got)
	}
}

func TestClarifyProcessErrorKilledWithoutContext(t *testing.T) {
	got := clarifyProcessError(context.Background(), 45*time.Second, errors.New("signal: killed"))
	if got == nil || !strings.Contains(got.Error(), "process killed") {
		t.Fatalf("got %v", got)
	}
}
