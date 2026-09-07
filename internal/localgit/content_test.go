package localgit

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

type fakeExitError struct {
	code int
}

func (e *fakeExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func (e *fakeExitError) ExitCode() int {
	return e.code
}

func TestContentOnDefaultAncestor(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --verify refs/remotes/origin/main":               []byte("abc\n"),
			"merge-base refs/remotes/origin/main feature":               []byte("abc\n"),
			"merge-base --is-ancestor feature refs/remotes/origin/main": []byte(""),
		},
	}
	in := NewInspector(fk)
	ok, reason, err := in.ContentOnDefault(context.Background(), dir, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("want ok, reason=%q calls=%v", reason, fk.calls)
	}
	if reason == "" {
		t.Fatal("want reason")
	}
}

func TestContentOnDefaultIdenticalTrees(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --verify refs/remotes/origin/main":   []byte("abc\n"),
			"merge-base refs/remotes/origin/main feature":   []byte("abc\n"),
			"diff --quiet refs/remotes/origin/main feature": []byte(""),
		},
		errors: map[string]error{
			"merge-base --is-ancestor feature refs/remotes/origin/main": &fakeExitError{code: 1},
		},
	}
	in := NewInspector(fk)
	ok, reason, err := in.ContentOnDefault(context.Background(), dir, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("want ok, reason=%q calls=%v", reason, fk.calls)
	}
	if !strings.Contains(reason, "identical") {
		t.Fatalf("reason=%q", reason)
	}
}

func TestContentOnDefaultTreesDiffer(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --verify refs/remotes/origin/main": []byte("abc\n"),
			"merge-base refs/remotes/origin/main feature": []byte("abc\n"),
		},
		errors: map[string]error{
			"merge-base --is-ancestor feature refs/remotes/origin/main": &fakeExitError{code: 1},
			"diff --quiet refs/remotes/origin/main feature":             &fakeExitError{code: 1},
		},
	}
	in := NewInspector(fk)
	ok, reason, err := in.ContentOnDefault(context.Background(), dir, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("want not ok, reason=%q", reason)
	}
	if !strings.Contains(reason, "trees differ") {
		t.Fatalf("reason=%q", reason)
	}
}

func TestContentOnDefaultFallsBackToLocalDefault(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --verify refs/heads/main":   []byte("abc\n"),
			"merge-base refs/heads/main feature":   []byte("abc\n"),
			"diff --quiet refs/heads/main feature": []byte(""),
		},
		errors: map[string]error{
			"rev-parse --verify refs/remotes/origin/main":      fmtError("missing"),
			"merge-base --is-ancestor feature refs/heads/main": &fakeExitError{code: 1},
		},
	}
	in := NewInspector(fk)
	ok, reason, err := in.ContentOnDefault(context.Background(), dir, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("want ok, reason=%q calls=%v", reason, fk.calls)
	}
}

func TestContentOnDefaultMissingBaseline(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		errors: map[string]error{
			"rev-parse --verify refs/remotes/origin/main": fmtError("missing"),
			"rev-parse --verify refs/heads/main":          fmtError("missing"),
		},
	}
	in := NewInspector(fk)
	ok, reason, err := in.ContentOnDefault(context.Background(), dir, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("want not ok")
	}
	if !strings.Contains(reason, "missing baseline") {
		t.Fatalf("reason=%q", reason)
	}
}

func TestContentOnDefaultUnrelated(t *testing.T) {
	dir := t.TempDir()
	fk := &fakeExec{
		responses: map[string][]byte{
			"rev-parse --verify refs/remotes/origin/main": []byte("abc\n"),
		},
		errors: map[string]error{
			"merge-base refs/remotes/origin/main feature": fmtError("unrelated"),
		},
	}
	in := NewInspector(fk)
	ok, reason, err := in.ContentOnDefault(context.Background(), dir, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if ok || reason != "unrelated histories" {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
}

func TestContentOnDefaultNilInspector(t *testing.T) {
	var in *Inspector
	_, _, err := in.ContentOnDefault(context.Background(), filepath.Join(t.TempDir(), "r"), "feature", "main")
	if err == nil {
		t.Fatal("want error")
	}
}

func TestContentOnDefaultRejectsDefaultBranch(t *testing.T) {
	in := NewInspector(&fakeExec{responses: map[string][]byte{}})
	ok, reason, err := in.ContentOnDefault(context.Background(), t.TempDir(), "main", "main")
	if err != nil {
		t.Fatal(err)
	}
	if ok || reason != "branch is default" {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
}
