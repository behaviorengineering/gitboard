//go:build integration

package localgit_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

func TestRemoveSafeCheckoutLinkedWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	wtDir := filepath.Join(root, "feature-wt")
	if err := os.MkdirAll(mainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run(mainDir, "init", "-b", "main")
	run(mainDir, "config", "user.email", "t@example.com")
	run(mainDir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(mainDir, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(mainDir, "add", "README")
	run(mainDir, "commit", "-m", "init")
	run(mainDir, "worktree", "add", "-b", "feature", wtDir)

	in := localgit.NewInspector(cliexec.New())
	if err := in.RemoveSafeCheckout(context.Background(), wtDir, "feature", "main"); err != nil {
		t.Fatalf("RemoveSafeCheckout: %v", err)
	}
	if _, err := os.Stat(wtDir); !os.IsNotExist(err) {
		t.Fatalf("worktree path still exists: %v", err)
	}
	cmd := exec.Command("git", "branch", "--list", "feature")
	cmd.Dir = mainDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("branch list: %v\n%s", err, out)
	}
	if len(out) != 0 {
		t.Fatalf("feature branch still listed: %q", out)
	}
}

func TestRemoveSafeCheckoutMainWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupPullRepos(t)
	run := gitRunner(t)

	run(local, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(local, "feat"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(local, "add", "feat")
	run(local, "commit", "-m", "feat")

	// Default branch falls behind origin while we sit on the feature branch.
	run(origin, "commit", "--allow-empty", "-m", "origin-ahead-1")
	run(origin, "commit", "--allow-empty", "-m", "origin-ahead-2")

	in := localgit.NewInspector(cliexec.New())
	if err := in.RemoveSafeCheckout(context.Background(), local, "feature", "main"); err != nil {
		t.Fatalf("RemoveSafeCheckout: %v", err)
	}
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = local
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("current branch: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "main" {
		t.Fatalf("current branch=%q want main", got)
	}

	count, err := exec.Command("git", "-C", local, "rev-list", "--left-right", "--count", "origin/main...main").Output()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Fields(strings.TrimSpace(string(count)))
	if len(parts) != 2 || parts[0] != "0" || parts[1] != "0" {
		t.Fatalf("main vs origin after remove: %q (want 0 0)", string(count))
	}
}

func TestRemoveSafeCheckoutMainWorktreeDivergedDefault(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupPullRepos(t)
	run := gitRunner(t)

	run(local, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(local, "feat"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(local, "add", "feat")
	run(local, "commit", "-m", "feat")

	// Local main gains a unique commit while we sit on feature (diverged tip).
	run(local, "branch", "tmp-main-edit", "main")
	run(local, "checkout", "tmp-main-edit")
	run(local, "commit", "--allow-empty", "-m", "local-main-only")
	run(local, "branch", "-f", "main", "tmp-main-edit")
	run(local, "checkout", "feature")
	run(local, "branch", "-D", "tmp-main-edit")

	// Origin main also advances so local main is ahead and behind.
	run(origin, "commit", "--allow-empty", "-m", "origin-ahead-1")
	run(origin, "commit", "--allow-empty", "-m", "origin-ahead-2")

	countBefore, err := exec.Command("git", "-C", local, "fetch", "origin", "main").CombinedOutput()
	if err != nil {
		t.Fatalf("fetch: %v\n%s", err, countBefore)
	}
	count, err := exec.Command("git", "-C", local, "rev-list", "--left-right", "--count", "main...origin/main").Output()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Fields(strings.TrimSpace(string(count)))
	if len(parts) != 2 || parts[0] == "0" || parts[1] == "0" {
		t.Fatalf("setup want diverged main, got %q (ahead behind vs origin)", string(count))
	}

	in := localgit.NewInspector(cliexec.New())
	if err := in.RemoveSafeCheckout(context.Background(), local, "feature", "main"); err != nil {
		t.Fatalf("RemoveSafeCheckout: %v", err)
	}
	cur, err := exec.Command("git", "-C", local, "branch", "--show-current").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(cur)); got != "main" {
		t.Fatalf("current branch=%q want main", got)
	}
	after, err := exec.Command("git", "-C", local, "rev-list", "--left-right", "--count", "origin/main...main").Output()
	if err != nil {
		t.Fatal(err)
	}
	afterParts := strings.Fields(strings.TrimSpace(string(after)))
	if len(afterParts) != 2 || afterParts[0] != "0" || afterParts[1] != "0" {
		t.Fatalf("main vs origin after remove: %q (want 0 0)", string(after))
	}
	list, err := exec.Command("git", "-C", local, "branch", "--list", "feature").Output()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("feature branch still listed: %q", list)
	}
}
