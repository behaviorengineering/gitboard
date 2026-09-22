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
	HostGitHub      Host = "github"
	HostGitLab      Host = "gitlab"
	HostAzureDevOps Host = "azuredevops"
	HostBitbucket   Host = "bitbucket"
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
	// FetchSeconds TTL-gates git fetch origin before local↔origin sync.
	// nil → default 120; explicit 0 always fetches; negative → default.
	FetchSeconds *int `yaml:"fetch_seconds"`
	// ScanSeconds TTL-gates ScanRoots discovery under local.roots.
	// nil → default 300; explicit 0 always rescans; negative → default.
	ScanSeconds *int `yaml:"scan_seconds"`
}

// LLM holds optional AI triage settings.
type LLM struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
	APIKey  string `yaml:"api_key"`
}

// DefaultLLMModel is the Polypus/cf_local model used when config omits llm.model.
const DefaultLLMModel = "cf_local/@cf/qwen/qwen3-30b-a3b-fp8"

// FailureDump holds error-only inference dump settings.
type FailureDump struct {
	Enabled     *bool  `yaml:"enabled"`
	Dir         string `yaml:"dir"`
	MaxAgeHours int    `yaml:"max_age_hours"`
	MaxFiles    int    `yaml:"max_files"`
}

// OpenInference holds OTEL / dump settings for agent and LLM work.
type OpenInference struct {
	Enabled     *bool       `yaml:"enabled"`
	Endpoint    string      `yaml:"endpoint"`
	ServiceName string      `yaml:"service_name"`
	FailureDump FailureDump `yaml:"failure_dump"`
}

// GitHubSync lists GitHub orgs to discover.
type GitHubSync struct {
	Orgs []string `yaml:"orgs"`
}

// GitLabSync lists GitLab groups to discover.
type GitLabSync struct {
	Groups []string `yaml:"groups"`
}

// AzureDevOpsSync lists Azure DevOps orgs (and optional projects) to discover.
type AzureDevOpsSync struct {
	Orgs     []string `yaml:"orgs"`
	Projects []string `yaml:"projects"` // optional project-name filter within orgs
}

// BitbucketSync lists Bitbucket workspaces to discover.
type BitbucketSync struct {
	Workspaces []string `yaml:"workspaces"`
}

// SyncSources names upstreams used by gitboard sync.
type SyncSources struct {
	GitHub      GitHubSync      `yaml:"github"`
	GitLab      GitLabSync      `yaml:"gitlab"`
	AzureDevOps AzureDevOpsSync `yaml:"azuredevops"`
	Bitbucket   BitbucketSync   `yaml:"bitbucket"`
}

// DefaultViewID is the synthetic view used when config omits views.
const DefaultViewID = "default"

// View is a named subset of tracked project ids for board loading.
type View struct {
	ID       string   `yaml:"id" json:"id"`
	Label    string   `yaml:"label" json:"label"`
	Projects []string `yaml:"projects" json:"projects"`
}

// UI holds dashboard presentation settings.
type UI struct {
	// PollSeconds is the browser auto-refresh interval in seconds.
	// nil / omitted → default 30; 0 disables polling.
	PollSeconds *int `yaml:"poll_seconds"`
	// HideBranches are path.Match patterns for branch names omitted from the
	// dashboard Branches column. Empty / omitted → show all (default off).
	HideBranches []string `yaml:"hide_branches,omitempty" json:"hide_branches,omitempty"`
}

// Upstream holds server-side TTL seconds for forge CLI slices.
// Explicit 0 disables caching for that slice (always refetch).
type Upstream struct {
	// HeadsSeconds caches remote branch list + default branch. nil → 120.
	HeadsSeconds *int `yaml:"heads_seconds"`
	// MergedSeconds caches merged PR/MR lists for prune hints. nil → 600.
	MergedSeconds *int `yaml:"merged_seconds"`
}

// File is the full user config on disk.
type File struct {
	LLM           LLM           `yaml:"llm"`
	OpenInference OpenInference `yaml:"openinference"`
	UI            UI            `yaml:"ui"`
	Upstream      Upstream      `yaml:"upstream"`
	Local         Local         `yaml:"local"`
	Sync          SyncSources   `yaml:"sync"`
	Views         []View        `yaml:"views,omitempty"`
	Projects      []Project     `yaml:"projects"`
}

// DefaultExample is the template written by init.
const DefaultExample = `llm:
  base_url: http://127.0.0.1:1320/v1
  model: cf_local/@cf/qwen/qwen3-30b-a3b-fp8
  api_key: ""

# OpenInference tracing + error-only dumps for agent/LLM work.
openinference:
  enabled: true
  endpoint: ""
  service_name: gitboard
  failure_dump:
    enabled: true
    dir: ""
    max_age_hours: 48
    max_files: 20

ui:
  poll_seconds: 30
  # Optional path.Match patterns; empty / omitted = show all.
  # Hides matching names from the Branches column only (prune data unchanged).
  # hide_branches:
  #   - majordomo-context/*
  #   - dependabot/*

# Server-side forge cache TTLs (0 = always refetch that slice).
upstream:
  heads_seconds: 120
  merged_seconds: 600

# Optional: scan these trees for local checkouts (incl. git worktrees).
# Match is via origin remote → project host/path. Override per project with local_path.
# fetch_seconds TTL-gates git fetch origin before ↑/↓ sync (0 = always fetch).
# scan_seconds TTL-gates root ScanRoots (0 = always rescan; default 300).
local:
  roots: []
  fetch_seconds: 120
  scan_seconds: 300

sync:
  github:
    orgs: []
  gitlab:
    groups: []
  azuredevops:
    orgs: []
    projects: []
  bitbucket:
    workspaces: []

# Named board views (subsets of projects). Empty / omitted → implicit "default"
# view containing every tracked project. Membership may overlap across views.
# views:
#   - id: work
#     label: Work
#     projects: [demo]

# Curated by: gitboard sync
# Optional per project: local_path: ~/code/my-clone
projects: []
`

// DefaultPollSeconds is used when ui.poll_seconds is unset (zero means use default on load).
const DefaultPollSeconds = 30

// DefaultHeadsSeconds caches remote heads when upstream.heads_seconds is omitted.
const DefaultHeadsSeconds = 120

// DefaultMergedSeconds caches merged reviews when upstream.merged_seconds is omitted.
const DefaultMergedSeconds = 600

// DefaultFetchSeconds caches git fetch origin when local.fetch_seconds is omitted.
const DefaultFetchSeconds = 120

// DefaultScanSeconds caches ScanRoots when local.scan_seconds is omitted.
const DefaultScanSeconds = 300

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
		return expandHome(v)
	}
	if v := strings.TrimSpace(os.Getenv("GITBOARD_PROJECTS")); v != "" {
		return expandHome(v)
	}
	userPath := filepath.Join(Dir(), "config.yaml")
	if fileExists(userPath) {
		return userPath
	}
	cwd, cwdErr := os.Getwd()
	if cwdErr == nil && cwd != "" {
		for _, name := range []string{"config.yaml", "projects.yaml"} {
			p := filepath.Join(cwd, name)
			if fileExists(p) {
				return p
			}
		}
	}
	return userPath
}

func expandHome(p string) string {
	p = strings.TrimSpace(p)
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return p
		}
		if p == "~" {
			return home
		}
		return filepath.Join(home, p[2:])
	}
	return p
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
	if err := normalizeViews(&doc); err != nil {
		return File{}, err
	}
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
	if err := normalizeViews(&doc); err != nil {
		return err
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

// HasSyncSources reports whether any org, group, or workspace is configured.
func (s SyncSources) HasSyncSources() bool {
	return len(s.GitHub.Orgs) > 0 ||
		len(s.GitLab.Groups) > 0 ||
		len(s.AzureDevOps.Orgs) > 0 ||
		len(s.Bitbucket.Workspaces) > 0
}

func normalizeSync(s *SyncSources) {
	s.GitHub.Orgs = trimNonEmpty(s.GitHub.Orgs)
	s.GitLab.Groups = trimNonEmpty(s.GitLab.Groups)
	s.AzureDevOps.Orgs = trimNonEmpty(s.AzureDevOps.Orgs)
	s.AzureDevOps.Projects = trimNonEmpty(s.AzureDevOps.Projects)
	s.Bitbucket.Workspaces = trimNonEmpty(s.Bitbucket.Workspaces)
}

func normalizeLocal(l *Local) {
	l.Roots = trimNonEmpty(l.Roots)
}

func normalizeViews(doc *File) error {
	if doc == nil {
		return nil
	}
	known := make(map[string]struct{}, len(doc.Projects))
	for _, p := range doc.Projects {
		known[p.ID] = struct{}{}
	}
	seen := map[string]struct{}{}
	var out []View
	for i, v := range doc.Views {
		id := strings.TrimSpace(v.ID)
		label := strings.TrimSpace(v.Label)
		if id == "" {
			return fmt.Errorf("views[%d]: missing id", i)
		}
		if label == "" {
			return fmt.Errorf("views[%d]: missing label", i)
		}
		key := strings.ToLower(id)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("views[%d]: duplicate id %q", i, id)
		}
		seen[key] = struct{}{}
		members := trimNonEmpty(v.Projects)
		var kept []string
		memberSeen := map[string]struct{}{}
		for _, pid := range members {
			if _, ok := known[pid]; !ok {
				continue // drop orphans
			}
			if _, ok := memberSeen[pid]; ok {
				continue
			}
			memberSeen[pid] = struct{}{}
			kept = append(kept, pid)
		}
		out = append(out, View{ID: id, Label: label, Projects: kept})
	}
	doc.Views = out
	return nil
}

// EffectiveViews returns configured views, or a single default view over all projects
// when views is empty/omitted.
func (f File) EffectiveViews() []View {
	if len(f.Views) > 0 {
		out := make([]View, len(f.Views))
		copy(out, f.Views)
		return out
	}
	ids := make([]string, 0, len(f.Projects))
	for _, p := range f.Projects {
		ids = append(ids, p.ID)
	}
	return []View{{
		ID:       DefaultViewID,
		Label:    "Default",
		Projects: ids,
	}}
}

// ResolveView returns the view for id (case-sensitive match on configured id).
// Empty viewID selects the first effective view.
func (f File) ResolveView(viewID string) (View, error) {
	views := f.EffectiveViews()
	viewID = strings.TrimSpace(viewID)
	if viewID == "" {
		return views[0], nil
	}
	for _, v := range views {
		if v.ID == viewID {
			return v, nil
		}
	}
	return View{}, fmt.Errorf("unknown view %q", viewID)
}

// ProjectsForView returns tracked projects that belong to the resolved view,
// preserving config project order.
func (f File) ProjectsForView(viewID string) ([]Project, View, error) {
	view, err := f.ResolveView(viewID)
	if err != nil {
		return nil, View{}, err
	}
	want := make(map[string]struct{}, len(view.Projects))
	for _, id := range view.Projects {
		want[id] = struct{}{}
	}
	var out []Project
	for _, p := range f.Projects {
		if _, ok := want[p.ID]; ok {
			out = append(out, p)
		}
	}
	return out, view, nil
}

// PruneViewMembership drops project ids that are no longer tracked.
func PruneViewMembership(views []View, projects []Project) []View {
	known := make(map[string]struct{}, len(projects))
	for _, p := range projects {
		known[p.ID] = struct{}{}
	}
	out := make([]View, 0, len(views))
	for _, v := range views {
		var kept []string
		seen := map[string]struct{}{}
		for _, id := range v.Projects {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := known[id]; !ok {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			kept = append(kept, id)
		}
		v.Projects = kept
		out = append(out, v)
	}
	return out
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
	host := Host(strings.ToLower(string(p.Host)))
	switch host {
	case HostGitHub, HostGitLab, HostAzureDevOps, HostBitbucket:
	default:
		return fmt.Errorf("host must be github, gitlab, azuredevops, or bitbucket")
	}
	parts := strings.Split(strings.Trim(p.Path, "/"), "/")
	minParts := 2
	pathHint := "owner/repo"
	if host == HostAzureDevOps {
		minParts = 3
		pathHint = "org/project/repo"
	}
	if len(parts) < minParts {
		return fmt.Errorf("path must be %s", pathHint)
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("path must be %s", pathHint)
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
	case HostGitLab:
		return "https://gitlab.com/" + path
	case HostAzureDevOps:
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			org, project, repo := parts[0], parts[1], parts[len(parts)-1]
			return fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", org, project, repo)
		}
		return "https://dev.azure.com/" + path
	case HostBitbucket:
		return "https://bitbucket.org/" + path
	default:
		return "https://gitlab.com/" + path
	}
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
			DefaultLLMModel,
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

// EffectiveHeadsSeconds returns the remote-heads cache TTL in seconds.
// Explicit 0 disables caching. Negative values fall back to the default.
func (f File) EffectiveHeadsSeconds() int {
	if f.Upstream.HeadsSeconds == nil {
		return DefaultHeadsSeconds
	}
	if *f.Upstream.HeadsSeconds < 0 {
		return DefaultHeadsSeconds
	}
	return *f.Upstream.HeadsSeconds
}

// EffectiveMergedSeconds returns the merged PR/MR cache TTL in seconds.
// Explicit 0 disables caching. Negative values fall back to the default.
func (f File) EffectiveMergedSeconds() int {
	if f.Upstream.MergedSeconds == nil {
		return DefaultMergedSeconds
	}
	if *f.Upstream.MergedSeconds < 0 {
		return DefaultMergedSeconds
	}
	return *f.Upstream.MergedSeconds
}

// EffectiveFetchSeconds returns the git fetch origin TTL in seconds.
// Explicit 0 disables caching (always fetch). Negative values fall back to the default.
func (f File) EffectiveFetchSeconds() int {
	if f.Local.FetchSeconds == nil {
		return DefaultFetchSeconds
	}
	if *f.Local.FetchSeconds < 0 {
		return DefaultFetchSeconds
	}
	return *f.Local.FetchSeconds
}

// EffectiveScanSeconds returns the local roots ScanRoots TTL in seconds.
// Explicit 0 disables caching (always rescan). Negative values fall back to the default.
func (f File) EffectiveScanSeconds() int {
	if f.Local.ScanSeconds == nil {
		return DefaultScanSeconds
	}
	if *f.Local.ScanSeconds < 0 {
		return DefaultScanSeconds
	}
	return *f.Local.ScanSeconds
}

// AgentsDir is where agentsession stores sessions (<config Dir>/agents).
func AgentsDir() string {
	return filepath.Join(Dir(), "agents")
}

// EffectiveOpenInference merges file settings with env for tracing dumps.
func (f File) EffectiveOpenInference() OpenInference {
	enabled := true
	if f.OpenInference.Enabled != nil {
		enabled = *f.OpenInference.Enabled
	}
	if v := strings.TrimSpace(os.Getenv("GITBOARD_OPENINFERENCE_ENABLED")); v != "" {
		enabled = v == "1" || strings.EqualFold(v, "true")
	}
	dumpEnabled := true
	if f.OpenInference.FailureDump.Enabled != nil {
		dumpEnabled = *f.OpenInference.FailureDump.Enabled
	}
	en := enabled
	den := dumpEnabled
	dir := strings.TrimSpace(f.OpenInference.FailureDump.Dir)
	if dir == "" {
		dir = filepath.Join(Dir(), "logs", "inference-failures")
	}
	maxAge := f.OpenInference.FailureDump.MaxAgeHours
	if maxAge <= 0 {
		maxAge = 48
	}
	maxFiles := f.OpenInference.FailureDump.MaxFiles
	if maxFiles <= 0 {
		maxFiles = 20
	}
	return OpenInference{
		Enabled:     &en,
		Endpoint:    firstNonEmpty(os.Getenv("GITBOARD_OPENINFERENCE_ENDPOINT"), f.OpenInference.Endpoint),
		ServiceName: firstNonEmpty(f.OpenInference.ServiceName, "gitboard"),
		FailureDump: FailureDump{
			Enabled:     &den,
			Dir:         dir,
			MaxAgeHours: maxAge,
			MaxFiles:    maxFiles,
		},
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
