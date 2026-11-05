package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, interactive bool, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := NewRoot(Options{
		Version:     "v9.9.9",
		Interactive: func() bool { return interactive },
	})
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	err = root.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}

func TestBareInvokePrintsGuide(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, false)
	if err != nil {
		t.Fatalf("bare invoke: %v", err)
	}
	for _, want := range []string{"gitboard v9.9.9", "never starts serve", "Documentation for agents", "AGENTS.md", "gitboard help"} {
		if !strings.Contains(out, want) {
			t.Errorf("guide missing %q:\n%s", want, out)
		}
	}
	if n := strings.Count(out, "\n"); n < 15 {
		t.Errorf("guide has %d lines, want a multi-line layout", n)
	}
}

func TestGuideDocPathsExist(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"AGENTS.md", "README.md"} {
		if _, err := os.Stat(filepath.Join("..", "..", p)); err != nil {
			t.Errorf("guide names %s but it does not exist: %v", p, err)
		}
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"version"}, {"--version"}} {
		out, _, err := run(t, false, args...)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if out != "gitboard v9.9.9\n" {
			t.Errorf("%v printed %q", args, out)
		}
	}
}

func TestHelpListsCommands(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, false, "help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	for _, want := range []string{"serve", "sync", "init", "version"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q:\n%s", want, out)
		}
	}
}

func TestUnknownCommandFails(t *testing.T) {
	t.Parallel()
	_, stderr, err := run(t, false, "not-a-command")
	if err == nil {
		t.Fatal("unknown command succeeded")
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("stderr = %q, want unknown command message", stderr)
	}
}

func TestInitDryRunWritesNothing(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.yaml")
	out, _, err := run(t, false, "init", "--config", path, "--dry-run")
	if err != nil {
		t.Fatalf("init --dry-run: %v", err)
	}
	if !strings.Contains(out, "would create") {
		t.Errorf("out = %q", out)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run created %s (stat err %v)", path, err)
	}
}

func TestSyncWithoutTerminalRefuses(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.yaml")
	_, _, err := run(t, false, "sync", "--config", path)
	if !errors.Is(err, errNeedsTerminal) {
		t.Fatalf("err = %v, want errNeedsTerminal", err)
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("refused sync still wrote %s", path)
	}
}
