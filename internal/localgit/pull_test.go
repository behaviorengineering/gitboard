//go:build integration

package localgit_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

func TestPullFFOnlyCheckedOut(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupPullRepos(t)

	run := gitRunner(t)
	run(origin, "commit", "--allow-empty", "-m", "origin-ahead")
	run(local, "fetch", "origin")

	in := localgit.NewInspector(cliexec.New())
	st := in.InspectPath(context.Background(), local)
	in.EnrichOriginSync(context.Background(), &st)
	if st.DefaultBehind != 1 {
		t.Fatalf("want behind=1 got %d", st.DefaultBehind)
	}

	if err := in.PullFFOnly(context.Background(), local, "main"); err != nil {
		t.Fatalf("PullFFOnly: %v", err)
	}
	st = in.InspectPath(context.Background(), local)
	in.EnrichOriginSync(context.Background(), &st)
	if st.DefaultBehind != 0 || st.DefaultAhead != 0 {
		t.Fatalf("after pull ahead=%d behind=%d", st.DefaultAhead, st.DefaultBehind)
	}
}

func TestPullFFOnlyNotCheckedOut(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupPullRepos(t)

	run := gitRunner(t)
	run(local, "checkout", "-b", "feature")
	run(origin, "commit", "--allow-empty", "-m", "origin-ahead")
	run(local, "fetch", "origin")

	in := localgit.NewInspector(cliexec.New())
	if err := in.PullFFOnly(context.Background(), local, "main"); err != nil {
		t.Fatalf("PullFFOnly: %v", err)
	}

	out, err := exec.Command("git", "-C", local, "rev-list", "--left-right", "--count", "origin/main...main").Output()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 || parts[0] != "0" || parts[1] != "0" {
		t.Fatalf("main vs origin after pull: %q", string(out))
	}
	cur, _ := exec.Command("git", "-C", local, "branch", "--show-current").Output()
	if got := strings.TrimSpace(string(cur)); got != "feature" {
		t.Fatalf("current branch=%q want feature", got)
	}
}

func TestPullFFOnlyRefusesDirty(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupPullRepos(t)
	run := gitRunner(t)
	run(origin, "commit", "--allow-empty", "-m", "origin-ahead")
	run(local, "fetch", "origin")
	if err := os.WriteFile(filepath.Join(local, "README"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	in := localgit.NewInspector(cliexec.New())
	err := in.PullFFOnly(context.Background(), local, "main")
	if !errors.Is(err, localgit.ErrDirtyTree) {
		t.Fatalf("want ErrDirtyTree, got %v", err)
	}
}

func TestPullFFOnlyUpdatesSubmodules(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	run := gitRunner(t)

	subOrigin := filepath.Join(root, "sub-origin")
	if err := os.MkdirAll(subOrigin, 0o755); err != nil {
		t.Fatal(err)
	}
	run(subOrigin, "init", "-b", "main")
	run(subOrigin, "config", "user.email", "t@example.com")
	run(subOrigin, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(subOrigin, "pack.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(subOrigin, "add", "pack.txt")
	run(subOrigin, "commit", "-m", "pack-v1")

	origin := filepath.Join(root, "origin")
	if err := os.MkdirAll(origin, 0o755); err != nil {
		t.Fatal(err)
	}
	run(origin, "init", "-b", "main")
	run(origin, "config", "user.email", "t@example.com")
	run(origin, "config", "user.name", "t")
	run(origin, "config", "protocol.file.allow", "always")
	run(origin, "-c", "protocol.file.allow=always", "submodule", "add", subOrigin, "packs")
	run(origin, "commit", "-m", "add-packs")

	local := filepath.Join(root, "local")
	run(root, "-c", "protocol.file.allow=always", "clone", "--recurse-submodules", origin, local)
	run(local, "config", "user.email", "t@example.com")
	run(local, "config", "user.name", "t")
	run(local, "config", "protocol.file.allow", "always")

	if err := os.WriteFile(filepath.Join(subOrigin, "pack.txt"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(subOrigin, "add", "pack.txt")
	run(subOrigin, "commit", "-m", "pack-v2")
	run(filepath.Join(origin, "packs"), "fetch", "origin")
	run(filepath.Join(origin, "packs"), "checkout", "main")
	run(filepath.Join(origin, "packs"), "pull", "--ff-only", "origin", "main")
	run(origin, "add", "packs")
	run(origin, "commit", "-m", "bump-packs")

	wantSHA, err := exec.Command("git", "-C", subOrigin, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSpace(string(wantSHA))

	run(local, "fetch", "origin")
	in := localgit.NewInspector(cliexec.New())
	if err := in.PullFFOnly(context.Background(), local, "main"); err != nil {
		t.Fatalf("PullFFOnly: %v", err)
	}

	gotSHA, err := exec.Command("git", "-C", filepath.Join(local, "packs"), "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(gotSHA))
	if got != want {
		t.Fatalf("submodule HEAD=%s want %s", got, want)
	}
	status, err := exec.Command("git", "-C", local, "status", "--porcelain").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(status)) != "" {
		t.Fatalf("want clean tree after pull, got %q", string(status))
	}
}

func TestPullFFOnlyClearsStaleSubmoduleDirtBeforePull(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	run := gitRunner(t)

	subOrigin := filepath.Join(root, "sub-origin")
	if err := os.MkdirAll(subOrigin, 0o755); err != nil {
		t.Fatal(err)
	}
	run(subOrigin, "init", "-b", "main")
	run(subOrigin, "config", "user.email", "t@example.com")
	run(subOrigin, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(subOrigin, "pack.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(subOrigin, "add", "pack.txt")
	run(subOrigin, "commit", "-m", "pack-v1")

	origin := filepath.Join(root, "origin")
	if err := os.MkdirAll(origin, 0o755); err != nil {
		t.Fatal(err)
	}
	run(origin, "init", "-b", "main")
	run(origin, "config", "user.email", "t@example.com")
	run(origin, "config", "user.name", "t")
	run(origin, "config", "protocol.file.allow", "always")
	run(origin, "-c", "protocol.file.allow=always", "submodule", "add", subOrigin, "packs")
	run(origin, "commit", "-m", "add-packs")

	local := filepath.Join(root, "local")
	run(root, "-c", "protocol.file.allow=always", "clone", "--recurse-submodules", origin, local)
	run(local, "config", "user.email", "t@example.com")
	run(local, "config", "user.name", "t")
	run(local, "config", "protocol.file.allow", "always")

	if err := os.WriteFile(filepath.Join(subOrigin, "pack.txt"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(subOrigin, "add", "pack.txt")
	run(subOrigin, "commit", "-m", "pack-v2")
	run(filepath.Join(origin, "packs"), "fetch", "origin")
	run(filepath.Join(origin, "packs"), "checkout", "main")
	run(filepath.Join(origin, "packs"), "pull", "--ff-only", "origin", "main")
	run(origin, "add", "packs")
	run(origin, "commit", "-m", "bump-packs")

	// Fast-forward the parent without updating the nested checkout: dirty submodule.
	run(local, "fetch", "origin")
	run(local, "merge", "--ff-only", "origin/main")
	status, err := exec.Command("git", "-C", local, "status", "--porcelain").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(status), "packs") {
		t.Fatalf("setup want dirty submodule, got %q", string(status))
	}

	run(origin, "commit", "--allow-empty", "-m", "origin-ahead-again")
	run(local, "fetch", "origin")

	in := localgit.NewInspector(cliexec.New())
	if err := in.PullFFOnly(context.Background(), local, "main"); err != nil {
		t.Fatalf("PullFFOnly: %v", err)
	}
	status, err = exec.Command("git", "-C", local, "status", "--porcelain").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(status)) != "" {
		t.Fatalf("want clean tree after pull, got %q", string(status))
	}
}

func TestPullFFOnlyRefusesDiverged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	origin, local := setupPullRepos(t)
	run := gitRunner(t)
	run(origin, "commit", "--allow-empty", "-m", "origin-only")
	run(local, "commit", "--allow-empty", "-m", "local-only")
	run(local, "fetch", "origin")

	in := localgit.NewInspector(cliexec.New())
	err := in.PullFFOnly(context.Background(), local, "main")
	if !errors.Is(err, localgit.ErrDiverged) {
		t.Fatalf("want ErrDiverged, got %v", err)
	}
}

func TestPullFFOnlyRefusesInvalidBranch(t *testing.T) {
	in := localgit.NewInspector(cliexec.New())
	err := in.PullFFOnly(context.Background(), "/tmp", "evil:ref")
	if !errors.Is(err, localgit.ErrInvalidBranch) {
		t.Fatalf("want ErrInvalidBranch, got %v", err)
	}
}

func TestValidateBranchName(t *testing.T) {
	cases := []struct {
		name string
		ok   bool
	}{
		{"HEAD", false},
		{"main", true},
		{"feat/x", true},
		{"feat/config-sync-dashboard", true},
		{"", false},
		{"-bad", false},
		{"evil:ref", false},
		{"has space", false},
		{"a..b", false},
		{".hidden", false},
	}
	for _, tc := range cases {
		err := localgit.ValidateBranchName(tc.name)
		if tc.ok && err != nil {
			t.Fatalf("%q: unexpected %v", tc.name, err)
		}
		if !tc.ok && !errors.Is(err, localgit.ErrInvalidBranch) {
			t.Fatalf("%q: want ErrInvalidBranch, got %v", tc.name, err)
		}
	}
}

func setupPullRepos(t *testing.T) (origin, local string) {
	t.Helper()
	root := t.TempDir()
	origin = filepath.Join(root, "origin")
	local = filepath.Join(root, "local")
	if err := os.MkdirAll(origin, 0o755); err != nil {
		t.Fatal(err)
	}
	run := gitRunner(t)
	run(origin, "init", "-b", "main")
	run(origin, "config", "user.email", "t@example.com")
	run(origin, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(origin, "README"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(origin, "add", "README")
	run(origin, "commit", "-m", "init")
	run(root, "clone", origin, local)
	run(local, "config", "user.email", "t@example.com")
	run(local, "config", "user.name", "t")
	return origin, local
}

func gitRunner(t *testing.T) func(dir string, args ...string) {
	t.Helper()
	return func(dir string, args ...string) {
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
}
