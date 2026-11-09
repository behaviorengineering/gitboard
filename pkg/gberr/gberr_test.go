package gberr_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

func TestWrapKeepsChainAndCode(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("read meta: %w", fs.ErrNotExist)
	err := fmt.Errorf("handler: %w", gberr.Wrap(cause, gberr.CodeNotFound, "pruneagent.Session", "session not found").
		With("session_id", "abc"))

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("errors.Is lost the cause: %v", err)
	}
	if got := gberr.CodeOf(err); got != gberr.CodeNotFound {
		t.Fatalf("CodeOf = %q", got)
	}
	e, ok := gberr.As(err)
	if !ok || e.Msg != "session not found" || e.Op != "pruneagent.Session" {
		t.Fatalf("As = %+v, %v", e, ok)
	}
	if got := gberr.FieldsOf(err)["session_id"]; got != "abc" {
		t.Fatalf("FieldsOf session_id = %q", got)
	}
	want := "pruneagent.Session: session not found: read meta: file does not exist"
	if e.Error() != want {
		t.Fatalf("Error() = %q, want %q", e.Error(), want)
	}
}

func TestCodeOfOutermostWins(t *testing.T) {
	t.Parallel()
	inner := gberr.New(gberr.CodeInvalid, "inner", "bad")
	outer := gberr.Wrap(inner, gberr.CodeUnavailable, "outer", "forge failed")
	if got := gberr.CodeOf(outer); got != gberr.CodeUnavailable {
		t.Fatalf("CodeOf = %q", got)
	}
	fields := gberr.FieldsOf(gberr.Wrap(inner.With("k", "inner"), gberr.CodeInternal, "o", "m").With("k", "outer"))
	if fields["k"] != "outer" {
		t.Fatalf("outer field must win: %v", fields)
	}
}

func TestCodeOfUntyped(t *testing.T) {
	t.Parallel()
	if got := gberr.CodeOf(nil); got != "" {
		t.Fatalf("nil: %q", got)
	}
	if got := gberr.CodeOf(errors.New("boom")); got != gberr.CodeInternal {
		t.Fatalf("plain: %q", got)
	}
	if got := gberr.CodeOf(fmt.Errorf("run: %w", context.DeadlineExceeded)); got != gberr.CodeCanceled {
		t.Fatalf("deadline: %q", got)
	}
	if !gberr.HasCode(fmt.Errorf("x: %w", context.Canceled), gberr.CodeCanceled) {
		t.Fatal("HasCode cancelled")
	}
}

func TestErrorTextWithEmptyParts(t *testing.T) {
	t.Parallel()
	if got := gberr.New(gberr.CodeInvalid, "", "").Error(); got != "invalid" {
		t.Fatalf("empty: %q", got)
	}
	if got := gberr.New(gberr.CodeInvalid, "", "bad input").Error(); got != "bad input" {
		t.Fatalf("msg only: %q", got)
	}
}
