// Package projectsync is public API: forge discovery of repositories from a
// config's sync sources and the rules that turn a selection into the tracked
// project list (add, remove, replace by host and path).
package projectsync

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// Lister discovers repositories from configured forges.
type Lister interface {
	ListGitHub(ctx context.Context, org string) ([]remotegit.RepoRef, error)
	ListGitLab(ctx context.Context, group string) ([]remotegit.RepoRef, error)
	ListAzureDevOps(ctx context.Context, org string) ([]remotegit.RepoRef, error)
	ListBitbucket(ctx context.Context, workspace string) ([]remotegit.RepoRef, error)
}

// ForgeLister adapts forge clients for discovery.
type ForgeLister struct {
	GitHub      *remotegit.GitHub
	GitLab      *remotegit.GitLab
	AzureDevOps *remotegit.AzureDevOps
	Bitbucket   *remotegit.Bitbucket
}

// listerMissing reports a ForgeLister built without the named client.
func listerMissing(forge string) error {
	return gberr.Wrap(remotegit.ErrClientMissing, gberr.CodeInternal, "projectsync.ForgeLister", forge+" client missing")
}

func (f ForgeLister) ListGitHub(ctx context.Context, org string) ([]remotegit.RepoRef, error) {
	if f.GitHub == nil {
		return nil, listerMissing("GitHub")
	}
	return f.GitHub.ListOrgRepos(ctx, org)
}

func (f ForgeLister) ListGitLab(ctx context.Context, group string) ([]remotegit.RepoRef, error) {
	if f.GitLab == nil {
		return nil, listerMissing("GitLab")
	}
	return f.GitLab.ListGroupRepos(ctx, group)
}

func (f ForgeLister) ListAzureDevOps(ctx context.Context, org string) ([]remotegit.RepoRef, error) {
	if f.AzureDevOps == nil {
		return nil, listerMissing("AzureDevOps")
	}
	return f.AzureDevOps.ListOrgRepos(ctx, org)
}

func (f ForgeLister) ListBitbucket(ctx context.Context, workspace string) ([]remotegit.RepoRef, error) {
	if f.Bitbucket == nil {
		return nil, listerMissing("Bitbucket")
	}
	return f.Bitbucket.ListWorkspaceRepos(ctx, workspace)
}

// Candidate is one selectable repository.
type Candidate struct {
	Host    boardconfig.Host `json:"host"`
	Path    string           `json:"path"`
	Name    string           `json:"name"`
	Tracked bool             `json:"tracked"`
	Index   int              `json:"index"` // 1-based display index.
}

// DiscoverResult is forge discovery output with optional per-source warnings.
type DiscoverResult struct {
	Candidates []Candidate
	Warnings   []string
}

// Discover lists unique repos from sync sources, marking already tracked ones.
// A failing org or group is recorded in Warnings; other sources still contribute.
// hostFilter limits discovery to one forge; empty means every forge, and any
// other value outside boardconfig's hosts is an error.
func Discover(ctx context.Context, lister Lister, doc boardconfig.File, hostFilter string) (DiscoverResult, error) {
	if lister == nil {
		return DiscoverResult{}, errors.New("projectsync.Discover: lister missing")
	}
	want, err := parseHostFilter(hostFilter)
	if err != nil {
		return DiscoverResult{}, err
	}
	var refs []remotegit.RepoRef
	var warnings []string
	for _, src := range discoverySources(lister, doc) {
		if want != "" && src.host != want {
			continue
		}
		got, warns, err := collectSource(ctx, src)
		if err != nil {
			return DiscoverResult{}, err
		}
		refs, warnings = append(refs, got...), append(warnings, warns...)
	}
	return DiscoverResult{Candidates: dedupeCandidates(refs, trackedKeys(doc.Projects)), Warnings: warnings}, nil
}

// discoverySource is one forge's sync sources and how to list them.
type discoverySource struct {
	host  boardconfig.Host
	label string   // names one source in a warning, for example "github org".
	names []string // orgs, groups or workspaces from the config.
	list  func(context.Context, string) ([]remotegit.RepoRef, error)
	// filter keeps a subset of a source's repos; nil keeps all.
	filter func([]remotegit.RepoRef) []remotegit.RepoRef
}

// discoverySources lists the forges in discovery order.
func discoverySources(lister Lister, doc boardconfig.File) []discoverySource {
	return []discoverySource{
		{host: boardconfig.HostGitHub, label: "github org", names: doc.Sync.GitHub.Orgs, list: lister.ListGitHub},
		{host: boardconfig.HostGitLab, label: "gitlab group", names: doc.Sync.GitLab.Groups, list: lister.ListGitLab},
		{
			host: boardconfig.HostAzureDevOps, label: "azuredevops org", names: doc.Sync.AzureDevOps.Orgs,
			list: lister.ListAzureDevOps, filter: azureProjectFilter(doc.Sync.AzureDevOps.Projects),
		},
		{host: boardconfig.HostBitbucket, label: "bitbucket workspace", names: doc.Sync.Bitbucket.Workspaces, list: lister.ListBitbucket},
	}
}

// parseHostFilter maps "" to "" (every forge) and anything else to a Host.
func parseHostFilter(hostFilter string) (boardconfig.Host, error) {
	if strings.TrimSpace(hostFilter) == "" {
		return "", nil
	}
	h, err := boardconfig.ParseHost(hostFilter)
	if err != nil {
		return "", fmt.Errorf("projectsync.Discover: host filter %q: %w", hostFilter, err)
	}
	return h, nil
}

// collectSource lists every name in src. A failing name becomes a warning;
// only a cancelled ctx stops the walk with an error.
func collectSource(ctx context.Context, src discoverySource) ([]remotegit.RepoRef, []string, error) {
	var refs []remotegit.RepoRef
	var warnings []string
	for _, name := range src.names {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		list, err := src.list(ctx, name)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s %s: %v", src.label, name, err))
			continue
		}
		if src.filter != nil {
			list = src.filter(list)
		}
		refs = append(refs, list...)
	}
	return refs, warnings, nil
}

// azureProjectFilter keeps repos whose project (the second path segment) is
// in projects, compared case-insensitively. No projects keeps every repo.
func azureProjectFilter(projects []string) func([]remotegit.RepoRef) []remotegit.RepoRef {
	if len(projects) == 0 {
		return nil
	}
	allowed := map[string]struct{}{}
	for _, p := range projects {
		allowed[strings.ToLower(p)] = struct{}{}
	}
	return func(refs []remotegit.RepoRef) []remotegit.RepoRef {
		var out []remotegit.RepoRef
		for _, r := range refs {
			parts, err := boardconfig.SplitRepoPath(boardconfig.HostAzureDevOps, r.Path)
			if err != nil {
				continue
			}
			if _, ok := allowed[strings.ToLower(parts[1])]; ok {
				out = append(out, r)
			}
		}
		return out
	}
}

// trackedKeys indexes the configured projects by host and path.
func trackedKeys(projects []boardconfig.Project) map[string]struct{} {
	tracked := make(map[string]struct{}, len(projects))
	for _, p := range projects {
		tracked[localgit.TrackKey(p.Host, p.Path)] = struct{}{}
	}
	return tracked
}

// dedupeCandidates sorts refs by host and path, keeps the first of each
// host and path, marks tracked ones and numbers them from 1.
func dedupeCandidates(refs []remotegit.RepoRef, tracked map[string]struct{}) []Candidate {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Host != refs[j].Host {
			return refs[i].Host < refs[j].Host
		}
		return refs[i].Path < refs[j].Path
	})
	seen := map[string]struct{}{}
	var out []Candidate
	for _, r := range refs {
		key := localgit.TrackKey(r.Host, r.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		_, isTracked := tracked[key]
		out = append(out, Candidate{
			Host:    r.Host,
			Path:    r.Path,
			Name:    r.Name,
			Tracked: isTracked,
			Index:   len(out) + 1,
		})
	}
	return out
}

// projectFromRef builds a config project from a forge repo.
func projectFromRef(host boardconfig.Host, path, name string) boardconfig.Project {
	path = strings.Trim(path, "/")
	id := slugID(name, path)
	label := name
	if label == "" {
		parts := strings.Split(path, "/")
		label = parts[len(parts)-1]
	}
	return boardconfig.Project{
		ID:    id,
		Label: label,
		Host:  host,
		Path:  path,
	}
}

// applySelection replaces projects with the selected candidates.
// Existing local_path values are kept when host+path still match.
func applySelection(cands []Candidate, selected []int, existing []boardconfig.Project) ([]boardconfig.Project, error) {
	byIndex := map[int]Candidate{}
	for _, c := range cands {
		byIndex[c.Index] = c
	}
	keepLocal := map[string]string{}
	for _, p := range existing {
		if strings.TrimSpace(p.LocalPath) == "" {
			continue
		}
		keepLocal[localgit.TrackKey(p.Host, p.Path)] = p.LocalPath
	}
	var projects []boardconfig.Project
	usedIDs := map[string]struct{}{}
	for _, n := range selected {
		c, ok := byIndex[n]
		if !ok {
			return nil, fmt.Errorf("unknown selection index %d", n)
		}
		p := projectFromRef(c.Host, c.Path, c.Name)
		if lp, ok := keepLocal[localgit.TrackKey(p.Host, p.Path)]; ok {
			p.LocalPath = lp
		}
		p.ID = uniqueID(p.ID, usedIDs)
		usedIDs[p.ID] = struct{}{}
		projects = append(projects, p)
	}
	return projects, nil
}

// ApplySelectionByRefs replaces projects with candidates matching host+path refs.
// Unknown refs return an error. Empty refs yields an empty project list.
func ApplySelectionByRefs(cands []Candidate, refs []RepoRef, existing []boardconfig.Project) ([]boardconfig.Project, error) {
	byKey := map[string]Candidate{}
	for _, c := range cands {
		byKey[localgit.TrackKey(c.Host, c.Path)] = c
	}
	var indices []int
	seen := map[int]struct{}{}
	for i, ref := range refs {
		host, err := boardconfig.ParseHost(string(ref.Host))
		if err != nil {
			return nil, fmt.Errorf("refs[%d]: %w", i, err)
		}
		path := strings.Trim(ref.Path, "/")
		if path == "" {
			return nil, fmt.Errorf("refs[%d]: path required", i)
		}
		c, ok := byKey[localgit.TrackKey(host, path)]
		if !ok {
			return nil, fmt.Errorf("refs[%d]: unknown candidate %s %s", i, host, path)
		}
		if _, dup := seen[c.Index]; dup {
			continue
		}
		seen[c.Index] = struct{}{}
		indices = append(indices, c.Index)
	}
	return applySelection(cands, indices, existing)
}

// RepoRef is a host+path pair used by ApplySelectionByRefs.
type RepoRef struct {
	Host boardconfig.Host `json:"host"`
	Path string           `json:"path"`
}

// AddProject appends or updates a project by host+path.
func AddProject(projects []boardconfig.Project, host boardconfig.Host, path string) ([]boardconfig.Project, error) {
	path = strings.Trim(path, "/")
	parts, err := boardconfig.SplitRepoPath(host, path)
	if err != nil {
		return nil, err
	}
	name := parts[len(parts)-1]
	p := projectFromRef(host, path, name)
	key := localgit.TrackKey(host, path)
	for i, existing := range projects {
		if localgit.TrackKey(existing.Host, existing.Path) == key {
			p.ID = existing.ID
			p.LocalPath = existing.LocalPath
			projects[i] = p
			return projects, nil
		}
	}
	used := map[string]struct{}{}
	for _, existing := range projects {
		used[existing.ID] = struct{}{}
	}
	p.ID = uniqueID(p.ID, used)
	return append(projects, p), nil
}

func uniqueID(base string, used map[string]struct{}) string {
	if _, ok := used[base]; !ok {
		return base
	}
	for n := 2; ; n++ {
		id := fmt.Sprintf("%s-%d", base, n)
		if _, ok := used[id]; !ok {
			return id
		}
	}
}

// RemoveProject drops a project by id.
func RemoveProject(projects []boardconfig.Project, id string) ([]boardconfig.Project, bool) {
	id = strings.TrimSpace(id)
	var out []boardconfig.Project
	found := false
	for _, p := range projects {
		if p.ID == id {
			found = true
			continue
		}
		out = append(out, p)
	}
	return out, found
}

func slugID(name, path string) string {
	base := name
	if base == "" {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		base = parts[len(parts)-1]
	}
	base = strings.ToLower(base)
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = "repo"
	}
	return id
}
