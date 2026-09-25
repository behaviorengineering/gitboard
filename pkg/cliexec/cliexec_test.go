package cliexec

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunRequiresDeadline(t *testing.T) {
	r := New()
	_, err := r.Run(context.Background(), "true")
	if !errors.Is(err, ErrMissingDeadline) {
		t.Fatalf("got %v want ErrMissingDeadline", err)
	}
}

func TestRunEcho(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := New()
	out, err := r.Run(ctx, "echo", "ok")
	if err != nil || string(out) != "ok\n" && string(out) != "ok" {
		t.Fatalf("out=%q err=%v", out, err)
	}
}
