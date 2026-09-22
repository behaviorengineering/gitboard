package dashboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/syncproj"
	"github.com/behaviorengineering/gitboard/pkg/board"
)

// ViewSummaries builds switcher metadata from effective views.
func ViewSummaries(doc config.File) []board.ViewSummary {
	implicit := len(doc.Views) == 0
	views := doc.EffectiveViews()
	out := make([]board.ViewSummary, 0, len(views))
	for _, v := range views {
		out = append(out, board.ViewSummary{
			ID:       v.ID,
			Label:    v.Label,
			Count:    len(v.Projects),
			Implicit: implicit,
		})
	}
	return out
}

func (c *Commands) forgeLister() syncproj.ForgeLister {
	if c == nil || c.Service == nil {
		return syncproj.ForgeLister{}
	}
	return syncproj.ForgeLister{
		GitHub:      c.GitHub,
		GitLab:      c.GitLab,
		AzureDevOps: c.AzureDevOps,
		Bitbucket:   c.Bitbucket,
	}
}

// DiscoverCandidates lists forge repos from configured sync sources.
// Per-source forge failures are returned in DiscoverResult.Warnings.
func (c *Commands) DiscoverCandidates(ctx context.Context, doc config.File, hostFilter string) (syncproj.DiscoverResult, error) {
	if err := c.requireService(); err != nil {
		return syncproj.DiscoverResult{}, badRequestCause("", err)
	}
	if !doc.Sync.HasSyncSources() {
		return syncproj.DiscoverResult{}, badRequest("no sync sources configured")
	}
	res, err := syncproj.Discover(ctx, c.forgeLister(), doc, hostFilter)
	if err != nil {
		return syncproj.DiscoverResult{}, fmt.Errorf("dashboard.DiscoverCandidates: %w", err)
	}
	return res, nil
}

// AddTrackedProject appends or updates a project by host+path and prunes view orphans.
func (c *Commands) AddTrackedProject(doc config.File, host config.Host, path string) (config.File, error) {
	if err := c.requireService(); err != nil {
		return doc, badRequestCause("", err)
	}
	host = config.Host(strings.ToLower(strings.TrimSpace(string(host))))
	switch host {
	case config.HostGitHub, config.HostGitLab, config.HostAzureDevOps, config.HostBitbucket:
	default:
		return doc, badRequest("host must be github, gitlab, azuredevops, or bitbucket")
	}
	updated, err := syncproj.AddProject(doc.Projects, host, path)
	if err != nil {
		return doc, badRequestCause("", err)
	}
	doc.Projects = updated
	doc.Views = config.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// RemoveTrackedProject drops a project by id and prunes view membership.
func (c *Commands) RemoveTrackedProject(doc config.File, id string) (config.File, error) {
	if err := c.requireService(); err != nil {
		return doc, badRequestCause("", err)
	}
	updated, found := syncproj.RemoveProject(doc.Projects, id)
	if !found {
		return doc, badRequest("unknown project")
	}
	doc.Projects = updated
	doc.Views = config.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// ApplySyncSelection replaces the tracked set from host/path refs after rediscovery.
func (c *Commands) ApplySyncSelection(ctx context.Context, doc config.File, hostFilter string, refs []syncproj.RepoRef) (config.File, error) {
	if err := c.requireService(); err != nil {
		return doc, badRequestCause("", err)
	}
	if !doc.Sync.HasSyncSources() {
		return doc, badRequest("no sync sources configured")
	}
	res, err := syncproj.Discover(ctx, c.forgeLister(), doc, hostFilter)
	if err != nil {
		return doc, fmt.Errorf("dashboard.ApplySyncSelection discover: %w", err)
	}
	projects, err := syncproj.ApplySelectionByRefs(res.Candidates, refs, doc.Projects)
	if err != nil {
		return doc, badRequestCause("", err)
	}
	doc.Projects = projects
	doc.Views = config.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// SetSyncSources updates forge discovery orgs/groups/workspaces.
func (c *Commands) SetSyncSources(doc config.File, githubOrgs, gitlabGroups, azureOrgs, bitbucketWorkspaces []string) (config.File, error) {
	if err := c.requireService(); err != nil {
		return doc, badRequestCause("", err)
	}
	doc.Sync.GitHub.Orgs = githubOrgs
	doc.Sync.GitLab.Groups = gitlabGroups
	doc.Sync.AzureDevOps.Orgs = azureOrgs
	doc.Sync.Bitbucket.Workspaces = bitbucketWorkspaces
	if !doc.Sync.HasSyncSources() {
		return doc, badRequest("need at least one sync source (github org, gitlab group, azuredevops org, or bitbucket workspace)")
	}
	return doc, nil
}

// SetViews replaces named view definitions (must be non-empty).
func (c *Commands) SetViews(doc config.File, views []config.View) (config.File, error) {
	if err := c.requireService(); err != nil {
		return doc, badRequestCause("", err)
	}
	if len(views) == 0 {
		return doc, badRequest("views must not be empty")
	}
	doc.Views = config.PruneViewMembership(views, doc.Projects)
	return doc, nil
}

// SetLocalRoots replaces directories scanned for git checkouts.
// Empty roots are allowed (local mapping then relies on per-project local_path only).
func (c *Commands) SetLocalRoots(doc config.File, roots []string) (config.File, error) {
	if err := c.requireService(); err != nil {
		return doc, badRequestCause("", err)
	}
	cleaned := make([]string, 0, len(roots))
	seen := map[string]struct{}{}
	for _, r := range roots {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		cleaned = append(cleaned, r)
	}
	doc.Local.Roots = cleaned
	return doc, nil
}

// SetProjectLocalPath sets or clears an explicit checkout path for a tracked project.
// Empty localPath clears the override so scan roots can map the repo again.
func (c *Commands) SetProjectLocalPath(doc config.File, id, localPath string) (config.File, error) {
	if err := c.requireService(); err != nil {
		return doc, badRequestCause("", err)
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return doc, badRequest("project id required")
	}
	localPath = strings.TrimSpace(localPath)
	found := false
	for i := range doc.Projects {
		if doc.Projects[i].ID != id {
			continue
		}
		doc.Projects[i].LocalPath = localPath
		found = true
		break
	}
	if !found {
		return doc, badRequest("unknown project")
	}
	return doc, nil
}
