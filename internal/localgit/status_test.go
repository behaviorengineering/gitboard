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

func TestInspectPathWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	if err := os.MkdirAll(mainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
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
	run(mainDir, "branch", "feature")
	wt := filepath.Join(root, "feature-wt")
	run(mainDir, "worktree", "add", wt, "feature")

	in := localgit.NewInspector(cliexec.New())
	st := in.InspectPath(context.Background(), mainDir)
	if !st.Mapped || st.Error != "" {
		t.Fatalf("status: %+v", st)
	}
	if len(st.Worktrees) < 2 {
		t.Fatalf("worktrees=%d want >=2: %+v", len(st.Worktrees), st.Worktrees)
	}
	foundFeature := false
	for _, w := range st.Worktrees {
		if w.Branch == "feature" && !w.Main {
			foundFeature = true
		}
	}
	if !foundFeature {
		t.Fatalf("missing feature worktree: %+v", st.Worktrees)
	}
}
