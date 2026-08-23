package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/config"
)

func TestLoadValidProjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.yaml")
	if err := os.WriteFile(path, []byte(`projects:
  - id: demo
    label: Demo
    host: github
    path: org/repo
`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Projects) != 1 || doc.Projects[0].ID != "demo" {
		t.Fatalf("unexpected projects: %+v", doc.Projects)
	}
	if doc.Projects[0].OpenURL() != "https://github.com/org/repo" {
		t.Fatalf("open url: %s", doc.Projects[0].OpenURL())
	}
}

func TestLoadRejectsBadHost(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.yaml")
	if err := os.WriteFile(path, []byte(`projects:
  - id: demo
    label: Demo
    host: bitbucket
    path: org/repo
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error")
	}
}
