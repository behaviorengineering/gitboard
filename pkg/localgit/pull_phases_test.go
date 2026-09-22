package localgit

import (
	"context"
	"errors"
	"testing"
)

func TestPullFFOnlyWithPhasesRejectsInvalidBranch(t *testing.T) {
	in := NewInspector(&fakeExec{responses: map[string][]byte{}})
	_, err := in.PullFFOnlyWithPhases(context.Background(), t.TempDir(), "evil:ref")
	if !errors.Is(err, ErrInvalidBranch) {
		t.Fatalf("want ErrInvalidBranch, got %v", err)
	}
}

func TestSubmodulePinsChangedIdentity(t *testing.T) {
	in := NewInspector(&fakeExec{responses: map[string][]byte{}})
	ctx := context.Background()
	if in.submodulePinsChanged(ctx, t.TempDir(), "abc", "abc") {
		t.Fatal("same SHA must not report pin change")
	}
	if !in.submodulePinsChanged(ctx, t.TempDir(), "abc", "") {
		t.Fatal("missing SHA must fail closed")
	}
	if in.submodulePinsChanged(ctx, t.TempDir(), "", "") {
		t.Fatal("empty SHAs must not report pin change")
	}
}

func TestSubmodulePinsChangedWithoutGitmodules(t *testing.T) {
	in := NewInspector(&fakeExec{responses: map[string][]byte{}})
	if in.submodulePinsChanged(context.Background(), t.TempDir(), "abc", "def") != false {
		t.Fatal("dir without .gitmodules must report no pin change")
	}
}
