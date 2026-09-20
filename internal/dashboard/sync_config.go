package dashboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/board"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/syncproj"
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
	return syncproj.ForgeLister{GitHub: c.GitHub, GitLab: c.GitLab}
}

// DiscoverCandidates lists forge repos from configured sync sources.
func (c *Commands) DiscoverCandidates(ctx context.Context, doc config.File, hostFilter string) ([]syncproj.Candidate, error) {
	if c == nil {
		return nil, badRequest("sync commands unavailable")
	}
	if !doc.Sync.HasSyncSources() {
		return nil, badRequest("no sync sources configured")
	}
	cands, err := syncproj.Discover(ctx, c.forgeLister(), doc, hostFilter)
	if err != nil {
		return nil, fmt.Errorf("dashboard.DiscoverCandidates: %w", err)
	}
	return cands, nil
}

// AddTrackedProject appends or updates a project by host+path and prunes view orphans.
func (c *Commands) AddTrackedProject(doc config.File, host config.Host, path string) (config.File, error) {
	if c == nil {
		return doc, badRequest("sync commands unavailable")
	}
	host = config.Host(strings.ToLower(strings.TrimSpace(string(host))))
	if host != config.HostGitHub && host != config.HostGitLab {
		return doc, badRequest("host must be github or gitlab")
	}
	updated, err := syncproj.AddProject(doc.Projects, host, path)
	if err != nil {
		return doc, badRequest(err.Error())
	}
	doc.Projects = updated
	doc.Views = config.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// RemoveTrackedProject drops a project by id and prunes view membership.
func (c *Commands) RemoveTrackedProject(doc config.File, id string) (config.File, error) {
	if c == nil {
		return doc, badRequest("sync commands unavailable")
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
	if c == nil {
		return doc, badRequest("sync commands unavailable")
	}
	if !doc.Sync.HasSyncSources() {
		return doc, badRequest("no sync sources configured")
	}
	cands, err := syncproj.Discover(ctx, c.forgeLister(), doc, hostFilter)
	if err != nil {
		return doc, fmt.Errorf("dashboard.ApplySyncSelection discover: %w", err)
	}
	projects, err := syncproj.ApplySelectionByRefs(cands, refs, doc.Projects)
	if err != nil {
		return doc, badRequest(err.Error())
	}
	doc.Projects = projects
	doc.Views = config.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// SetSyncSources updates forge discovery orgs/groups.
func (c *Commands) SetSyncSources(doc config.File, githubOrgs, gitlabGroups []string) (config.File, error) {
	if c == nil {
		return doc, badRequest("sync commands unavailable")
	}
	doc.Sync.GitHub.Orgs = githubOrgs
	doc.Sync.GitLab.Groups = gitlabGroups
	if !doc.Sync.HasSyncSources() {
		return doc, badRequest("need at least one github org or gitlab group")
	}
	return doc, nil
}

// SetViews replaces named view definitions (must be non-empty).
func (c *Commands) SetViews(doc config.File, views []config.View) (config.File, error) {
	if c == nil {
		return doc, badRequest("sync commands unavailable")
	}
	if len(views) == 0 {
		return doc, badRequest("views must not be empty")
	}
	doc.Views = config.PruneViewMembership(views, doc.Projects)
	return doc, nil
}
