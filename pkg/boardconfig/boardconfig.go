// Package boardconfig is public API: the gitboard configuration document
// (tracked projects, views, sync sources and TTL settings) that pkg/dashboard,
// pkg/remotegit and pkg/projectsync take as input. It holds types and pure
// rules only; reading, writing and environment overlays live in the gitboard
// binary's loader.
package boardconfig

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// Host identifies which forge CLI backs a project.
type Host string

// Supported forge hosts.
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
	Projects []string `yaml:"projects"` // optional project-name filter within orgs.
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

// DefaultViewID is the synthetic view used when config omits views; it is the
// ActiveView a dashboard reports for an unviewed config.
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

const (
	defaultHeadsSeconds  = 120 // upstream.heads_seconds when omitted.
	defaultMergedSeconds = 600 // upstream.merged_seconds when omitted.
	defaultFetchSeconds  = 120 // local.fetch_seconds when omitted.
	defaultScanSeconds   = 300 // local.scan_seconds when omitted.
)

// HasSyncSources reports whether any org, group, or workspace is configured.
func (s SyncSources) HasSyncSources() bool {
	return len(s.GitHub.Orgs) > 0 ||
		len(s.GitLab.Groups) > 0 ||
		len(s.AzureDevOps.Orgs) > 0 ||
		len(s.Bitbucket.Workspaces) > 0
}

// Normalize validates every project and view and trims list fields in place:
// hosts are lower-cased, duplicate sync sources and roots are dropped, and
// view members that are not tracked projects are removed. Loaders call it
// after reading and before writing a File.
func (f *File) Normalize() error {
	if f == nil {
		return errors.New("boardconfig.Normalize: nil file")
	}
	normalizeSync(&f.Sync)
	f.Local.Roots = trimNonEmpty(f.Local.Roots)
	if err := validateUI(f.UI); err != nil {
		return err
	}
	for i, p := range f.Projects {
		if err := validateProject(p); err != nil {
			return fmt.Errorf("projects[%d]: %w", i, err)
		}
		f.Projects[i].Host = Host(strings.ToLower(strings.TrimSpace(string(p.Host))))
		f.Projects[i].LocalPath = strings.TrimSpace(p.LocalPath)
	}
	return normalizeViews(f)
}

// validateUI rejects settings that would otherwise be ignored without a
// word: a malformed hide pattern hides nothing and a negative poll interval
// silently fell back to the default.
func validateUI(ui UI) error {
	if ui.PollSeconds != nil && *ui.PollSeconds < 0 {
		return fmt.Errorf("ui.poll_seconds %d: must be 0 or more", *ui.PollSeconds)
	}
	for i, pat := range ui.HideBranches {
		if _, err := path.Match(pat, ""); err != nil {
			return fmt.Errorf("ui.hide_branches[%d] %q: %w", i, pat, err)
		}
	}
	return nil
}

func normalizeSync(s *SyncSources) {
	s.GitHub.Orgs = trimNonEmpty(s.GitHub.Orgs)
	s.GitLab.Groups = trimNonEmpty(s.GitLab.Groups)
	s.AzureDevOps.Orgs = trimNonEmpty(s.AzureDevOps.Orgs)
	s.AzureDevOps.Projects = trimNonEmpty(s.AzureDevOps.Projects)
	s.Bitbucket.Workspaces = trimNonEmpty(s.Bitbucket.Workspaces)
}

func normalizeViews(doc *File) error {
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
				continue // drop orphans.
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

// resolveView returns the view for id (case-sensitive match on configured id).
// Empty viewID selects the first effective view.
func (f File) resolveView(viewID string) (View, error) {
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
// preserving config project order. Empty viewID selects the first effective view.
func (f File) ProjectsForView(viewID string) ([]Project, View, error) {
	view, err := f.resolveView(viewID)
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
		return errors.New("missing id")
	}
	if strings.TrimSpace(p.Label) == "" {
		return errors.New("missing label")
	}
	host, err := ParseHost(string(p.Host))
	if err != nil {
		return err
	}
	_, err = SplitRepoPath(host, p.Path)
	return err
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

// EffectiveHeadsSeconds returns the remote-heads cache TTL in seconds.
// Explicit 0 disables caching. Negative values fall back to the default.
func (f File) EffectiveHeadsSeconds() int {
	return secondsOrDefault(f.Upstream.HeadsSeconds, defaultHeadsSeconds)
}

// EffectiveMergedSeconds returns the merged PR/MR cache TTL in seconds.
// Explicit 0 disables caching. Negative values fall back to the default.
func (f File) EffectiveMergedSeconds() int {
	return secondsOrDefault(f.Upstream.MergedSeconds, defaultMergedSeconds)
}

// EffectiveFetchSeconds returns the git fetch origin TTL in seconds.
// Explicit 0 disables caching (always fetch). Negative values fall back to the default.
func (f File) EffectiveFetchSeconds() int {
	return secondsOrDefault(f.Local.FetchSeconds, defaultFetchSeconds)
}

// EffectiveScanSeconds returns the local roots ScanRoots TTL in seconds.
// Explicit 0 disables caching (always rescan). Negative values fall back to the default.
func (f File) EffectiveScanSeconds() int {
	return secondsOrDefault(f.Local.ScanSeconds, defaultScanSeconds)
}

func secondsOrDefault(v *int, def int) int {
	if v == nil || *v < 0 {
		return def
	}
	return *v
}
