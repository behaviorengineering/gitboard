package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
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
	live := newConfigLive(path, doc, doc.EffectivePollSeconds(), cache.Clear, nil)
	got, _, err := live.snapshot()
	if err != nil {
		t.Fatal(err)
	}
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

	got, _, err = live.snapshot()
	if err != nil {
		t.Fatal(err)
	}
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

func TestConfigLiveFailsFastOnBadReload(t *testing.T) {
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
	live := newConfigLive(path, doc, 30, nil, nil)

	if err := os.WriteFile(path, []byte("projects: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, base.Add(time.Second), base.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	_, _, err = live.snapshot()
	if err == nil {
		t.Fatal("expected reload error on corrupt config")
	}
}

func TestViewsChanged(t *testing.T) {
	a := []config.View{{ID: "work", Label: "Work", Projects: []string{"a"}}}
	b := []config.View{{ID: "work", Label: "Work", Projects: []string{"a"}}}
	if viewsChanged(a, b) {
		t.Fatal("same views")
	}
	b[0].Projects = []string{"a", "b"}
	if !viewsChanged(a, b) {
		t.Fatal("membership change")
	}
}

func TestLocalRootsChanged(t *testing.T) {
	if localRootsChanged([]string{"a"}, []string{"a"}) {
		t.Fatal("same roots")
	}
	if !localRootsChanged([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("len change")
	}
	if !localRootsChanged([]string{"a"}, []string{"b"}) {
		t.Fatal("membership change")
	}
}

func writeConfig(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
