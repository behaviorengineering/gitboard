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

func TestFillCheckoutMetaAndDisplayID(t *testing.T) {
	c := localgit.Checkout{
		Path:         "/Users/me/Xynova/ai/n8n/providers/strop",
		Superproject: "/Users/me/Xynova/ai/n8n",
	}
	localgit.FillCheckoutMeta(&c)
	if c.Role != localgit.RoleSubmodule {
		t.Fatalf("role=%q", c.Role)
	}
	if c.RelPath != "providers/strop" {
		t.Fatalf("rel=%q", c.RelPath)
	}
	if c.ParentBasename != "n8n" {
		t.Fatalf("basename=%q", c.ParentBasename)
	}
	got := localgit.DisplayID(c, "content-pipelines")
	want := "content-pipelines → providers/strop"
	if got != want {
		t.Fatalf("display=%q want %q", got, want)
	}
	got = localgit.DisplayID(c, "")
	want = "n8n → providers/strop"
	if got != want {
		t.Fatalf("fallback display=%q want %q", got, want)
	}

	standalone := localgit.Checkout{Path: "/tmp/strop"}
	localgit.FillCheckoutMeta(&standalone)
	if standalone.Role != localgit.RoleStandalone {
		t.Fatalf("standalone role=%q", standalone.Role)
	}
	if localgit.DisplayID(standalone, "") != "/tmp/strop" {
		t.Fatalf("standalone display=%q", localgit.DisplayID(standalone, ""))
	}
}

func TestPickPrimaryOrder(t *testing.T) {
	checkouts := []localgit.Checkout{
		{Path: "/a/deep/parent/providers/repo", Role: localgit.RoleSubmodule, Superproject: "/a/deep/parent"},
		{Path: "/z/standalone", Role: localgit.RoleStandalone},
		{Path: "/b/shallow/providers/repo", Role: localgit.RoleSubmodule, Superproject: "/b/shallow"},
	}
	got, ok := localgit.PickPrimary("", checkouts)
	if !ok || got.Path != "/z/standalone" {
		t.Fatalf("prefer standalone: %+v ok=%v", got, ok)
	}

	onlySubs := []localgit.Checkout{
		{Path: "/a/deep/parent/providers/repo", Role: localgit.RoleSubmodule},
		{Path: "/b/shallow/providers/repo", Role: localgit.RoleSubmodule},
	}
	got, ok = localgit.PickPrimary("", onlySubs)
	if !ok || got.Path != "/b/shallow/providers/repo" {
		t.Fatalf("prefer shallow: %+v ok=%v", got, ok)
	}

	got, ok = localgit.PickPrimary("/explicit/path", checkouts)
	if !ok || got.Path != "/explicit/path" {
		t.Fatalf("explicit: %+v ok=%v", got, ok)
	}
}

func TestScanRootsMultipleClones(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	mkClone := func(dir string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		run := func(args ...string) {
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
		run("init", "-b", "main")
		run("config", "user.email", "t@example.com")
		run("config", "user.name", "t")
		run("remote", "add", "origin", "https://github.com/acme/widget.git")
		if err := os.WriteFile(filepath.Join(dir, "README"), []byte("hi\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		run("add", "README")
		run("commit", "-m", "init")
	}

	a := filepath.Join(root, "clone-a")
	b := filepath.Join(root, "clone-b")
	mkClone(a)
	mkClone(b)

	// Linked worktree of clone-a must not add a second appearance.
	wt := filepath.Join(root, "clone-a-wt")
	cmd := exec.Command("git", "worktree", "add", wt, "-b", "feature")
	cmd.Dir = a
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("worktree add: %v\n%s", err, out)
	}

	in := localgit.NewInspector(cliexec.New())
	disc := in.ScanRoots(context.Background(), []string{root})
	apps := disc.Appearances("github", "acme/widget")
	if len(apps) != 2 {
		t.Fatalf("appearances=%d want 2: %+v", len(apps), apps)
	}
	paths := map[string]bool{}
	for _, c := range apps {
		resolved, err := filepath.EvalSymlinks(c.Path)
		if err != nil {
			resolved = c.Path
		}
		paths[resolved] = true
		if c.Role != localgit.RoleStandalone {
			t.Fatalf("expected standalone: %+v", c)
		}
	}
	wantA, _ := filepath.EvalSymlinks(a)
	wantB, _ := filepath.EvalSymlinks(b)
	if !paths[wantA] || !paths[wantB] {
		t.Fatalf("paths=%v want %s and %s", paths, wantA, wantB)
	}
	wantWT, _ := filepath.EvalSymlinks(wt)
	if paths[wantWT] {
		t.Fatalf("linked worktree should not be a separate appearance")
	}
}

func TestScanRootsSubmoduleAppearance(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	childSrc := filepath.Join(root, "child-src")
	runGit := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
	for _, dir := range []string{parent, childSrc} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		runGit(dir, "init", "-b", "main")
		runGit(dir, "config", "user.email", "t@example.com")
		runGit(dir, "config", "user.name", "t")
		if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		runGit(dir, "add", "README")
		runGit(dir, "commit", "-m", "init")
	}
	runGit(childSrc, "remote", "add", "origin", "https://github.com/acme/child-src-only.git")
	runGit(parent, "remote", "add", "origin", "https://github.com/acme/parent.git")

	// File-protocol submodule from local path.
	runGit(parent, "-c", "protocol.file.allow=always", "submodule", "add", childSrc, "modules/child")
	runGit(parent, "commit", "-m", "add submodule")

	// Point submodule origin at the forge URL so ScanRoots can key it.
	childCheckout := filepath.Join(parent, "modules", "child")
	runGit(childCheckout, "remote", "set-url", "origin", "https://github.com/acme/child.git")

	in := localgit.NewInspector(cliexec.New())
	disc := in.ScanRoots(context.Background(), []string{root})
	apps := disc.Appearances("github", "acme/child")
	if len(apps) != 1 {
		t.Fatalf("child appearances=%d: %+v", len(apps), apps)
	}
	c := apps[0]
	if c.Role != localgit.RoleSubmodule {
		t.Fatalf("role=%q want submodule: %+v", c.Role, c)
	}
	if c.RelPath != "modules/child" {
		t.Fatalf("rel=%q path=%q super=%q", c.RelPath, c.Path, c.Superproject)
	}
	gotSuper, _ := filepath.EvalSymlinks(c.Superproject)
	wantSuper, _ := filepath.EvalSymlinks(parent)
	if gotSuper != wantSuper {
		t.Fatalf("super=%q want %q", c.Superproject, parent)
	}
	gotPath, _ := filepath.EvalSymlinks(c.Path)
	wantPath, _ := filepath.EvalSymlinks(childCheckout)
	if gotPath != wantPath {
		t.Fatalf("path=%q want %q", c.Path, childCheckout)
	}
	if got := localgit.DisplayID(c, "parent-label"); got != "parent-label → modules/child" {
		t.Fatalf("display=%q", got)
	}
}
