//go:build integration

package localgit_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

func TestContentOnDefaultSquashIdenticalTrees(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupPullRepos(t)
	run := gitRunner(t)

	run(local, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(local, "feat.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(local, "add", "feat.txt")
	run(local, "commit", "-m", "feat")
	run(local, "push", "-u", "origin", "feature")

	// Squash onto main on origin, then refresh local main.
	run(origin, "checkout", "main")
	run(origin, "merge", "--squash", "feature")
	run(origin, "commit", "-m", "feat (#1)")
	run(local, "fetch", "origin")
	run(local, "checkout", "main")
	run(local, "merge", "--ff-only", "origin/main")
	run(local, "checkout", "feature")

	in := localgit.NewInspector(cliexec.New())
	ok, reason, err := in.ContentOnDefault(context.Background(), local, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("want content on default after squash, reason=%q", reason)
	}
}

func TestContentOnDefaultLeftoverChange(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	_, local := setupPullRepos(t)
	run := gitRunner(t)

	run(local, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(local, "feat.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(local, "add", "feat.txt")
	run(local, "commit", "-m", "feat")

	in := localgit.NewInspector(cliexec.New())
	ok, reason, err := in.ContentOnDefault(context.Background(), local, "feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("want leftover content, reason=%q", reason)
	}
}
