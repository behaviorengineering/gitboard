package localgit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureWritableIndexNoLock(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
	}}
	in := NewInspector(fk)
	if err := in.EnsureWritableIndex(context.Background(), dir, false); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestEnsureWritableIndexFreshIsBusy(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lock, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(lock, now, now); err != nil {
		t.Fatal(err)
	}
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
	}}
	in := NewInspector(fk)
	err := in.EnsureWritableIndexAge(context.Background(), dir, true, time.Minute)
	if !errors.Is(err, ErrIndexLockBusy) {
		t.Fatalf("want ErrIndexLockBusy, got %v", err)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("fresh lock must not be removed: %v", err)
	}
}

func TestEnsureWritableIndexStaleNeedsConfirm(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lock, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * time.Minute)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
	}}
	in := NewInspector(fk)
	err := in.EnsureWritableIndexAge(context.Background(), dir, false, 30*time.Second)
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
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lock, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * time.Minute)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	fk := &fakeExec{responses: map[string][]byte{
		"rev-parse --git-path index.lock": []byte(".git/index.lock\n"),
	}}
	in := NewInspector(fk)
	if err := in.EnsureWritableIndexAge(context.Background(), dir, true, 30*time.Second); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("lock should be gone, stat=%v", err)
	}
}

func TestIsIndexLockError(t *testing.T) {
	if !IsIndexLockError(errors.New(`fatal: Unable to create '/tmp/repo/.git/index.lock': File exists.`)) {
		t.Fatal("expected index.lock message match")
	}
	if IsIndexLockError(errors.New("merge conflict")) {
		t.Fatal("unrelated error")
	}
}
