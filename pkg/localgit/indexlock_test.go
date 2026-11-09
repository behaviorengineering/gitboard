package localgit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureWritableIndexNoLock(t *testing.T) {
	dir := t.TempDir()
	common := filepath.Join(dir, ".git")
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
		"--git-common-dir":                []byte(common + "\n"),
	}}
	m := leasedMutator(t, fk, common)
	if err := m.EnsureWritableIndex(context.Background(), dir, false); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestEnsureWritableIndexFreshIsBusy(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o750); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(lock, now, now); err != nil {
		t.Fatal(err)
	}
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
	}}
	m := leasedMutator(t, fk, dir)
	err := m.ensureWritableIndex(context.Background(), dir, true, time.Minute)
	if !errors.Is(err, ErrIndexLockBusy) {
		t.Fatalf("want ErrIndexLockBusy, got %v", err)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("fresh lock must not be removed: %v", err)
	}
}

func TestEnsureWritableIndexStaleNeedsConfirm(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o750); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * time.Minute)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
	}}
	m := leasedMutator(t, fk, dir)
	err := m.ensureWritableIndex(context.Background(), dir, false, 30*time.Second)
	var stale *StaleIndexLockError
	if !errors.As(err, &stale) {
		t.Fatalf("want StaleIndexLockError, got %v", err)
	}
	if !errors.Is(err, ErrStaleIndexLock) {
		t.Fatalf("want ErrStaleIndexLock, got %v", err)
	}
	if stale.LockPath == "" || stale.RepoPath == "" {
		t.Fatalf("missing paths: %+v", stale)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("lock should remain until clear: %v", err)
	}
}

func TestEnsureWritableIndexStaleCleared(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o750); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * time.Minute)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
	}}
	m := leasedMutator(t, fk, dir)
	if err := m.ensureWritableIndex(context.Background(), dir, true, 30*time.Second); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("lock should be gone, stat=%v", err)
	}
}

func TestIsIndexLockError(t *testing.T) {
	if !IsIndexLockError(fmt.Errorf("pull: %w", &StaleIndexLockError{LockPath: "/tmp/repo/.git/index.lock"})) {
		t.Fatal("stale lock error must match")
	}
	if !IsIndexLockError(fmt.Errorf("%w at /tmp/x", ErrIndexLockBusy)) {
		t.Fatal("busy lock error must match")
	}
	// Message text alone is not a classification: a path or stderr that
	// mentions index.lock must not trigger lock recovery.
	if IsIndexLockError(errors.New(`fatal: Unable to create '/tmp/repo/.git/index.lock': File exists.`)) {
		t.Fatal("message text must not match")
	}
	if IsIndexLockError(errors.New("merge conflict")) {
		t.Fatal("unrelated error")
	}
}
