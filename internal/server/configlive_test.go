package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/remotegit"
)

func TestConfigLiveReloadsProjectsAndClearsCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	writeConfig(t, path, `projects:
  - id: a
    label: A
    host: github
    path: org/a
`)
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 8, 26, 6, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, base, base); err != nil {
		t.Fatal(err)
	}
	_ = st

	cache := remotegit.NewTTLCache()
	loads := 0
	_, err = cache.GetOrLoadHeads("github/org/a", time.Minute, false, func() (remotegit.HeadsSnapshot, error) {
		loads++
		return remotegit.HeadsSnapshot{DefaultBranch: "main"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	doc, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	live := newConfigLive(path, doc, doc.EffectivePollSeconds(), cache.Clear)
	got, _ := live.snapshot()
	if len(got.Projects) != 1 || got.Projects[0].ID != "a" {
		t.Fatalf("initial: %+v", got.Projects)
	}

	writeConfig(t, path, `projects:
  - id: a
    label: A
    host: github
    path: org/a
  - id: b
    label: B
    host: github
    path: org/b
`)
	newer := base.Add(2 * time.Second)
	if err := os.Chtimes(path, newer, newer); err != nil {
		t.Fatal(err)
	}

	got, _ = live.snapshot()
	if len(got.Projects) != 2 {
		t.Fatalf("after sync reload: %+v", got.Projects)
	}

	_, err = cache.GetOrLoadHeads("github/org/a", time.Minute, false, func() (remotegit.HeadsSnapshot, error) {
		loads++
		return remotegit.HeadsSnapshot{DefaultBranch: "main"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loads != 2 {
		t.Fatalf("project set change should clear forge cache, loads=%d", loads)
	}
}

func TestConfigLiveKeepsLastGoodOnBadReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	writeConfig(t, path, `projects:
  - id: a
    label: A
    host: github
    path: org/a
`)
	base := time.Date(2026, 8, 26, 7, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, base, base); err != nil {
		t.Fatal(err)
	}
	doc, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	live := newConfigLive(path, doc, 30, nil)

	if err := os.WriteFile(path, []byte("projects: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, base.Add(time.Second), base.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	got, _ := live.snapshot()
	if len(got.Projects) != 1 || got.Projects[0].ID != "a" {
		t.Fatalf("want last good snapshot, got %+v", got.Projects)
	}
}

func TestProjectsChanged(t *testing.T) {
	a := []config.Project{{ID: "a", Host: config.HostGitHub, Path: "o/a"}}
	b := []config.Project{{ID: "a", Host: config.HostGitHub, Path: "o/a"}}
	if projectsChanged(a, b) {
		t.Fatal("same set")
	}
	b = append(b, config.Project{ID: "b", Host: config.HostGitHub, Path: "o/b"})
	if !projectsChanged(a, b) {
		t.Fatal("added project")
	}
}

func writeConfig(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
