package localgit

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestMutatorRefusesRepoOutsideLease(t *testing.T) {
	dir := t.TempDir()
	leased := filepath.Join(dir, "a.git")
	other := filepath.Join(dir, "b.git")
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir": []byte(other + "\n"),
	}}
	m := leasedMutator(t, fx, leased)
	err := m.EnsureWritableIndex(context.Background(), filepath.Join(dir, "b"), false)
	if !errors.Is(err, ErrOutsideLease) {
		t.Fatalf("want ErrOutsideLease, got %v", err)
	}
	for _, call := range fx.calls {
		if !strings.Contains(call, "--git-common-dir") {
			t.Fatalf("refused mutation must not run other git commands: %v", fx.calls)
		}
	}
}

func TestMutatorFailsAfterRelease(t *testing.T) {
	dir := t.TempDir()
	fx := &fakeExec{responses: map[string][]byte{
		"--git-common-dir": []byte(dir + "\n"),
	}}
	lease, err := NewMutationCoordinator(nil).Acquire(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	m, err := lease.Mutator(fx)
	if err != nil {
		t.Fatal(err)
	}
	releaseOrFail(t, lease)
	if _, err := m.PullFFOnly(context.Background(), dir, "main"); !errors.Is(err, ErrLeaseNotHeld) {
		t.Fatalf("pull after release: want ErrLeaseNotHeld, got %v", err)
	}
	if _, err := lease.Mutator(fx); !errors.Is(err, ErrLeaseNotHeld) {
		t.Fatalf("mutator after release: want ErrLeaseNotHeld, got %v", err)
	}
	if len(fx.calls) != 0 {
		t.Fatalf("released lease must not run git: %v", fx.calls)
	}
}

func TestLeaseMutatorRequiresExec(t *testing.T) {
	lease, err := NewMutationCoordinator(nil).Acquire(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer releaseOrFail(t, lease)
	if _, err := lease.Mutator(nil); err == nil {
		t.Fatal("want error for nil Exec")
	}
}
