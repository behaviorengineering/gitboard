package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
)

func TestLoadValidProjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`projects:
  - id: demo
    label: Demo
    host: github
    path: org/repo
`), 0o600); err != nil {
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
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`projects:
  - id: demo
    label: Demo
    host: gitea
    path: org/repo
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	doc := boardconfig.File{
		LLM: boardconfig.LLM{
			BaseURL: "http://127.0.0.1:1320/v1",
			Model:   "test-model",
			APIKey:  "secret",
		},
		Sync: boardconfig.SyncSources{
			GitHub: boardconfig.GitHubSync{Orgs: []string{"behaviorengineering"}},
			GitLab: boardconfig.GitLabSync{Groups: []string{"behaviorengineering"}},
		},
		Projects: []boardconfig.Project{{
			ID: "demo", Label: "Demo", Host: boardconfig.HostGitHub, Path: "org/repo",
		}},
	}
	if err := config.Save(path, doc); err != nil {
		t.Fatal(err)
	}
	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.LLM.APIKey != "secret" || got.LLM.Model != "test-model" {
		t.Fatalf("llm: %+v", got.LLM)
	}
	if len(got.Sync.GitHub.Orgs) != 1 || got.Sync.GitHub.Orgs[0] != "behaviorengineering" {
		t.Fatalf("sync: %+v", got.Sync)
	}
	if len(got.Projects) != 1 || got.Projects[0].Path != "org/repo" {
		t.Fatalf("projects: %+v", got.Projects)
	}
}

func TestInitDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	created, err := config.Init(path)
	if err != nil || !created {
		t.Fatalf("first init: created=%v err=%v", created, err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	created, err = config.Init(path)
	if err != nil || created {
		t.Fatalf("second init: created=%v err=%v", created, err)
	}
	raw2, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(raw2) {
		t.Fatal("init overwrote existing config")
	}
}

// The init template spells out the dump retention; it must match the
// defaults EffectiveOpenInference applies when the file omits them.
func TestInitTemplateMatchesDumpDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if _, err := config.Init(path); err != nil {
		t.Fatal(err)
	}
	doc, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	dump := doc.OpenInference.FailureDump
	if dump.MaxAgeHours != config.DefaultDumpMaxAgeHours || dump.MaxFiles != config.DefaultDumpMaxFiles {
		t.Fatalf("template max_age_hours=%d max_files=%d, defaults %d and %d",
			dump.MaxAgeHours, dump.MaxFiles, config.DefaultDumpMaxAgeHours, config.DefaultDumpMaxFiles)
	}
}

func TestDefaultPathUsesConfigEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.yaml")
	t.Setenv("GITBOARD_CONFIG", path)
	t.Setenv("GITBOARD_PROJECTS", "")
	if got := config.DefaultPath(); got != path {
		t.Fatalf("DefaultPath=%s want %s", got, path)
	}
}

func TestDefaultPathPrefersUserConfig(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("GITBOARD_CONFIG", "")
	t.Setenv("GITBOARD_PROJECTS", "")
	userPath := filepath.Join(xdg, "gitboard", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userPath, []byte("projects: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := config.DefaultPath(); got != userPath {
		t.Fatalf("DefaultPath=%s want %s", got, userPath)
	}
}

func TestEffectiveLLMEnvOverridesFile(t *testing.T) {
	doc := boardconfig.File{
		LLM: boardconfig.LLM{
			BaseURL: "http://file/v1",
			Model:   "file-model",
			APIKey:  "file-key",
		},
	}
	t.Setenv("GITBOARD_LLM_BASE_URL", "http://env/v1")
	t.Setenv("GITBOARD_LLM_MODEL", "env-model")
	t.Setenv("GITBOARD_LLM_API_KEY", "env-key")
	got := config.EffectiveLLM(doc)
	if got.BaseURL != "http://env/v1" || got.Model != "env-model" || got.APIKey != "env-key" {
		t.Fatalf("effective: %+v", got)
	}
}

func TestAgentsDirUsesXDG(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	want := filepath.Join(xdg, "gitboard", "agents")
	if got := config.AgentsDir(); got != want {
		t.Fatalf("AgentsDir=%s want %s", got, want)
	}
}

func TestLoadLocalPathAndRoots(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`local:
  roots:
    - ~/code
projects:
  - id: demo
    label: Demo
    host: github
    path: org/repo
    local_path: ~/code/repo
`), 0o600); err != nil {
		t.Fatal(err)
	}
	doc, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Local.Roots) != 1 || doc.Local.Roots[0] != "~/code" {
		t.Fatalf("local: %+v", doc.Local)
	}
	if doc.Projects[0].LocalPath != "~/code/repo" {
		t.Fatalf("local_path: %q", doc.Projects[0].LocalPath)
	}
}

func TestEffectivePollSeconds(t *testing.T) {
	t.Setenv("GITBOARD_POLL_SECONDS", "")
	doc := boardconfig.File{}
	if got := config.EffectivePollSeconds(doc); got != config.DefaultPollSeconds {
		t.Fatalf("default=%d want %d", got, config.DefaultPollSeconds)
	}
	zero := 0
	doc.UI.PollSeconds = &zero
	if got := config.EffectivePollSeconds(doc); got != 0 {
		t.Fatalf("explicit zero=%d", got)
	}
	forty := 40
	doc.UI.PollSeconds = &forty
	t.Setenv("GITBOARD_POLL_SECONDS", "15")
	if got := config.EffectivePollSeconds(doc); got != 15 {
		t.Fatalf("env override=%d", got)
	}
}

func TestProjectsForViewFilters(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	doc := boardconfig.File{
		Projects: []boardconfig.Project{
			{ID: "a", Label: "A", Host: boardconfig.HostGitHub, Path: "o/a"},
			{ID: "b", Label: "B", Host: boardconfig.HostGitHub, Path: "o/b"},
			{ID: "c", Label: "C", Host: boardconfig.HostGitHub, Path: "o/c"},
		},
		Views: []boardconfig.View{
			{ID: "work", Label: "Work", Projects: []string{"a", "c", "missing"}},
		},
	}
	if err := config.Save(path, doc); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Views[0].Projects) != 2 {
		t.Fatalf("orphans should drop: %+v", loaded.Views[0].Projects)
	}
	projects, view, err := loaded.ProjectsForView("work")
	if err != nil {
		t.Fatal(err)
	}
	if view.ID != "work" || len(projects) != 2 || projects[0].ID != "a" || projects[1].ID != "c" {
		t.Fatalf("got projects=%+v view=%+v", projects, view)
	}
	if _, _, err := loaded.ProjectsForView("nope"); err == nil {
		t.Fatal("expected unknown view error")
	}
}

func TestLoadRejectsDuplicateViewIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`projects:
  - id: a
    label: A
    host: github
    path: o/a
views:
  - id: work
    label: Work
    projects: [a]
  - id: work
    label: Other
    projects: [a]
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected duplicate view id error")
	}
}

func TestLoadRejectsBadPollEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("projects: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITBOARD_POLL_SECONDS", "30s")
	if _, err := config.Load(path); err == nil {
		t.Fatal("a GITBOARD_POLL_SECONDS typo must fail Load instead of falling back")
	}
	t.Setenv("GITBOARD_POLL_SECONDS", "15")
	doc, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := config.EffectivePollSeconds(doc); got != 15 {
		t.Fatalf("poll = %d", got)
	}
}
