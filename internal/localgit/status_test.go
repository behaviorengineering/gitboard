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

func TestEnrichOriginSync(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	origin := filepath.Join(root, "origin")
	local := filepath.Join(root, "local")
	if err := os.MkdirAll(origin, 0o755); err != nil {
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
			t.Fatalf("git %v (dir=%s): %v\n%s", args, dir, err, out)
		}
	}
	write := func(dir, name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run(origin, "init", "-b", "main")
	run(origin, "config", "user.email", "t@example.com")
	run(origin, "config", "user.name", "t")
	write(origin, "README", "a\n")
	run(origin, "add", "README")
	run(origin, "commit", "-m", "init")

	run(root, "clone", origin, local)
	run(local, "config", "user.email", "t@example.com")
	run(local, "config", "user.name", "t")

	write(origin, "README", "a\nb\n")
	run(origin, "add", "README")
	run(origin, "commit", "-m", "origin-main")
	run(local, "fetch", "origin")

	run(local, "checkout", "-b", "feature")
	write(local, "feat", "1\n")
	run(local, "add", "feat")
	run(local, "commit", "-m", "feature-1")
	run(local, "push", "-u", "origin", "feature")
	write(local, "feat", "1\n2\n")
	run(local, "add", "feat")
	run(local, "commit", "-m", "feature-2")

	run(local, "checkout", "-b", "wip")
	write(local, "wip", "x\n")
	run(local, "add", "wip")
	run(local, "commit", "-m", "wip-only")
	run(local, "checkout", "main")

	in := localgit.NewInspector(cliexec.New())
	st := in.InspectPath(context.Background(), local)
	in.EnrichOriginSync(context.Background(), &st)
	if st.Error != "" {
		t.Fatalf("status: %+v", st)
	}
	if st.DefaultBranch != "main" {
		t.Fatalf("default=%q", st.DefaultBranch)
	}
	if st.DefaultBehind != 1 || st.DefaultAhead != 0 {
		t.Fatalf("default ahead=%d behind=%d", st.DefaultAhead, st.DefaultBehind)
	}

	byName := map[string]localgit.BranchSync{}
	for _, s := range st.OriginSync {
		byName[s.Name] = s
	}
	if _, ok := byName["wip"]; ok {
		t.Fatalf("local-only wip should not have origin sync: %+v", st.OriginSync)
	}
	mainSync, ok := byName["main"]
	if !ok || mainSync.Behind != 1 || mainSync.Ahead != 0 {
		t.Fatalf("main sync=%+v ok=%v", mainSync, ok)
	}
	featSync, ok := byName["feature"]
	if !ok || featSync.Ahead != 1 || featSync.Behind != 0 {
		t.Fatalf("feature sync=%+v ok=%v", featSync, ok)
	}
}
