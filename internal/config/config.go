package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Host identifies which forge CLI backs a project.
type Host string

const (
	HostGitHub Host = "github"
	HostGitLab Host = "gitlab"
)

// Project is one tracked repository.
type Project struct {
	ID        string `yaml:"id" json:"id"`
	Label     string `yaml:"label" json:"label"`
	Host      Host   `yaml:"host" json:"host"`
	Path      string `yaml:"path" json:"path"`
	LocalPath string `yaml:"local_path,omitempty" json:"local_path,omitempty"`
}

// Local holds optional disk discovery for checkout status.
type Local struct {
	// Roots are directories to scan for git checkouts (and worktrees).
	// Matching uses origin remote URL against project host+path.
	Roots []string `yaml:"roots"`
}

// LLM holds optional AI triage settings.
type LLM struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
	APIKey  string `yaml:"api_key"`
}

// GitHubSync lists GitHub orgs to discover.
type GitHubSync struct {
	Orgs []string `yaml:"orgs"`
}

// GitLabSync lists GitLab groups to discover.
type GitLabSync struct {
	Groups []string `yaml:"groups"`
}

// SyncSources names upstreams used by gitboard sync.
type SyncSources struct {
	GitHub GitHubSync `yaml:"github"`
	GitLab GitLabSync `yaml:"gitlab"`
}

// UI holds dashboard presentation settings.
type UI struct {
	// PollSeconds is the browser auto-refresh interval in seconds.
	// nil / omitted → default 30; 0 disables polling.
	PollSeconds *int `yaml:"poll_seconds"`
}

// File is the full user config on disk.
type File struct {
	LLM      LLM         `yaml:"llm"`
	UI       UI          `yaml:"ui"`
	Local    Local       `yaml:"local"`
	Sync     SyncSources `yaml:"sync"`
	Projects []Project   `yaml:"projects"`
}

// DefaultExample is the template written by init.
const DefaultExample = `llm:
  base_url: http://127.0.0.1:1320/v1
  model: cf_local/@cf/zai-org/glm-4.7-flash
  api_key: ""

ui:
  poll_seconds: 30

# Optional: scan these trees for local checkouts (incl. git worktrees).
# Match is via origin remote → project host/path. Override per project with local_path.
local:
  roots: []

sync:
  github:
    orgs: []
  gitlab:
    groups: []

# Curated by: gitboard sync
# Optional per project: local_path: ~/code/my-clone
projects: []
`

// DefaultPollSeconds is used when ui.poll_seconds is unset (zero means use default on load).
const DefaultPollSeconds = 30

// Dir returns ~/.config/gitboard (or $XDG_CONFIG_HOME/gitboard).
func Dir() string {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, "gitboard")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".config", "gitboard")
	}
	return filepath.Join(home, ".config", "gitboard")
}

// DefaultPath resolves the config file location.
// Order: GITBOARD_CONFIG, GITBOARD_PROJECTS (legacy), user config.yaml if present,
// cwd config.yaml, cwd projects.yaml, else user config.yaml path (may not exist yet).
func DefaultPath() string {
	if v := strings.TrimSpace(os.Getenv("GITBOARD_CONFIG")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GITBOARD_PROJECTS")); v != "" {
		return v
	}
	userPath := filepath.Join(Dir(), "config.yaml")
	if fileExists(userPath) {
		return userPath
	}
	cwd, _ := os.Getwd()
	if cwd != "" {
		for _, name := range []string{"config.yaml", "projects.yaml"} {
			p := filepath.Join(cwd, name)
			if fileExists(p) {
				return p
			}
		}
	}
	return userPath
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// Load reads a config YAML file (full document or legacy projects-only).
func Load(path string) (File, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("read config: %w", err)
	}
	var doc File
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return File{}, fmt.Errorf("parse config: %w", err)
	}
	for i, p := range doc.Projects {
		if err := validateProject(p); err != nil {
			return File{}, fmt.Errorf("projects[%d]: %w", i, err)
		}
		doc.Projects[i].Host = Host(strings.ToLower(string(p.Host)))
		doc.Projects[i].LocalPath = strings.TrimSpace(p.LocalPath)
	}
	normalizeSync(&doc.Sync)
	normalizeLocal(&doc.Local)
	return doc, nil
}

// Save writes the config file with mode 0600.
func Save(path string, doc File) error {
	normalizeSync(&doc.Sync)
	normalizeLocal(&doc.Local)
	for i, p := range doc.Projects {
		if err := validateProject(p); err != nil {
			return fmt.Errorf("projects[%d]: %w", i, err)
		}
		doc.Projects[i].Host = Host(strings.ToLower(string(p.Host)))
		doc.Projects[i].LocalPath = strings.TrimSpace(p.LocalPath)
	}
	raw, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// Init writes the default config if path does not exist.
func Init(path string) (created bool, err error) {
	if fileExists(path) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, fmt.Errorf("mkdir config: %w", err)
	}
	if err := os.WriteFile(path, []byte(DefaultExample), 0o600); err != nil {
		return false, fmt.Errorf("write config: %w", err)
	}
	return true, nil
}

// HasSyncSources reports whether any org or group is configured.
func (s SyncSources) HasSyncSources() bool {
	return len(s.GitHub.Orgs) > 0 || len(s.GitLab.Groups) > 0
}

func normalizeSync(s *SyncSources) {
	s.GitHub.Orgs = trimNonEmpty(s.GitHub.Orgs)
	s.GitLab.Groups = trimNonEmpty(s.GitLab.Groups)
}

func normalizeLocal(l *Local) {
	l.Roots = trimNonEmpty(l.Roots)
}

func trimNonEmpty(in []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

func validateProject(p Project) error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("missing id")
	}
	if strings.TrimSpace(p.Label) == "" {
		return fmt.Errorf("missing label")
	}
	switch Host(strings.ToLower(string(p.Host))) {
	case HostGitHub, HostGitLab:
	default:
		return fmt.Errorf("host must be github or gitlab")
	}
	parts := strings.Split(strings.Trim(p.Path, "/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("path must be owner/repo")
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("path must be owner/repo")
		}
	}
	return nil
}

// OpenURL returns the human web URL for a project.
func (p Project) OpenURL() string {
	path := strings.Trim(p.Path, "/")
	switch p.Host {
	case HostGitHub:
		return "https://github.com/" + path
	default:
		return "https://gitlab.com/" + path
	}
}

// OwnerRepo splits path into owner and repo (last two path segments).
func (p Project) OwnerRepo() (string, string) {
	parts := strings.Split(strings.Trim(p.Path, "/"), "/")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[len(parts)-2], parts[len(parts)-1]
}

// EffectiveLLM merges file LLM settings with environment overrides.
func (f File) EffectiveLLM() LLM {
	base := firstNonEmpty(
		os.Getenv("GITBOARD_LLM_BASE_URL"),
		os.Getenv("POLYPUS_BASE_URL"),
		f.LLM.BaseURL,
	)
	if base != "" && !strings.HasSuffix(base, "/v1") {
		base = strings.TrimRight(base, "/") + "/v1"
	}
	return LLM{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey: firstNonEmpty(
			os.Getenv("GITBOARD_LLM_API_KEY"),
			os.Getenv("OPENAI_API_KEY"),
			f.LLM.APIKey,
		),
		Model: firstNonEmpty(
			os.Getenv("GITBOARD_LLM_MODEL"),
			f.LLM.Model,
			"cf_local/@cf/zai-org/glm-4.7-flash",
		),
	}
}

// EffectivePollSeconds returns the UI poll interval in seconds.
// Env GITBOARD_POLL_SECONDS overrides the file. Omitted file value defaults to 30.
// Explicit 0 disables polling.
func (f File) EffectivePollSeconds() int {
	if v := strings.TrimSpace(os.Getenv("GITBOARD_POLL_SECONDS")); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n >= 0 {
			return n
		}
	}
	if f.UI.PollSeconds == nil {
		return DefaultPollSeconds
	}
	if *f.UI.PollSeconds < 0 {
		return DefaultPollSeconds
	}
	return *f.UI.PollSeconds
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// DefaultProjectsPath is retained for callers; prefer DefaultPath.
func DefaultProjectsPath() string {
	return DefaultPath()
}
