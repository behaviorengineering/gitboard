package dashboard

import (
	"context"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
)

// ViewSummaries builds switcher metadata from effective views.
func ViewSummaries(doc boardconfig.File) []board.ViewSummary {
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

func (c *Commands) forgeLister() projectsync.ForgeLister {
	if c == nil || c.Service == nil {
		return projectsync.ForgeLister{}
	}
	return projectsync.ForgeLister{
		GitHub:      c.gitHub,
		GitLab:      c.gitLab,
		AzureDevOps: c.azureDevOps,
		Bitbucket:   c.bitbucket,
	}
}

// DiscoverCandidates lists forge repos from configured sync sources.
// Per-source forge failures are returned in DiscoverResult.Warnings.
func (c *Commands) DiscoverCandidates(ctx context.Context, doc boardconfig.File, hostFilter string) (projectsync.DiscoverResult, error) {
	if err := c.requireService(); err != nil {
		return projectsync.DiscoverResult{}, err
	}
	return c.discover(ctx, "dashboard.DiscoverCandidates", doc, hostFilter)
}

// discover runs projectsync.Discover for op. No sync sources or an unknown
// host filter is CodeInvalid; a discovery failure is CodeUnavailable.
func (c *Commands) discover(ctx context.Context, op string, doc boardconfig.File, hostFilter string) (projectsync.DiscoverResult, error) {
	if !doc.Sync.HasSyncSources() {
		return projectsync.DiscoverResult{}, invalid(op, "no sync sources configured")
	}
	if strings.TrimSpace(hostFilter) != "" {
		if _, err := boardconfig.ParseHost(hostFilter); err != nil {
			return projectsync.DiscoverResult{}, invalidCause(op, err.Error(), err)
		}
	}
	res, err := projectsync.Discover(ctx, c.forgeLister(), doc, hostFilter)
	if err != nil {
		return projectsync.DiscoverResult{}, gberr.Wrap(err, gberr.CodeUnavailable, op, "repository discovery failed")
	}
	return res, nil
}

// AddTrackedProject appends or updates a project by host+path and prunes view orphans.
func (c *Commands) AddTrackedProject(doc boardconfig.File, host boardconfig.Host, path string) (boardconfig.File, error) {
	if err := c.requireService(); err != nil {
		return doc, err
	}
	host, err := boardconfig.ParseHost(string(host))
	if err != nil {
		return doc, invalidCause("dashboard.AddTrackedProject", err.Error(), err)
	}
	updated, err := projectsync.AddProject(doc.Projects, host, path)
	if err != nil {
		return doc, invalidCause("dashboard.AddTrackedProject", err.Error(), err)
	}
	doc.Projects = updated
	doc.Views = boardconfig.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// RemoveTrackedProject drops a project by id and prunes view membership.
func (c *Commands) RemoveTrackedProject(doc boardconfig.File, id string) (boardconfig.File, error) {
	if err := c.requireService(); err != nil {
		return doc, err
	}
	updated, found := projectsync.RemoveProject(doc.Projects, id)
	if !found {
		return doc, invalid("dashboard.RemoveTrackedProject", "unknown project")
	}
	doc.Projects = updated
	doc.Views = boardconfig.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// ApplySyncSelection replaces the tracked set from host/path refs after rediscovery.
func (c *Commands) ApplySyncSelection(ctx context.Context, doc boardconfig.File, hostFilter string, refs []projectsync.RepoRef) (boardconfig.File, error) {
	const op = "dashboard.ApplySyncSelection"
	if err := c.requireService(); err != nil {
		return doc, err
	}
	res, err := c.discover(ctx, op, doc, hostFilter)
	if err != nil {
		return doc, err
	}
	// A failed source leaves its repositories out of the candidates, so a
	// valid selection would be rejected as unknown. Report the outage instead.
	if len(res.Warnings) > 0 {
		return doc, gberr.New(gberr.CodeUnavailable, op, "a sync source failed during discovery; retry when the forge is reachable").
			With("warnings", strings.Join(res.Warnings, "; "))
	}
	return c.applySelection(op, doc, res.Candidates, refs)
}

// ApplySelection replaces the tracked set with refs, chosen from cands that
// the caller already discovered, and prunes view membership. A ref that is
// not among cands is CodeInvalid.
func (c *Commands) ApplySelection(doc boardconfig.File, cands []projectsync.Candidate, refs []projectsync.RepoRef) (boardconfig.File, error) {
	if err := c.requireService(); err != nil {
		return doc, err
	}
	return c.applySelection("dashboard.ApplySelection", doc, cands, refs)
}

func (c *Commands) applySelection(op string, doc boardconfig.File, cands []projectsync.Candidate, refs []projectsync.RepoRef) (boardconfig.File, error) {
	projects, err := projectsync.ApplySelectionByRefs(cands, refs, doc.Projects)
	if err != nil {
		return doc, invalidCause(op, err.Error(), err)
	}
	doc.Projects = projects
	doc.Views = boardconfig.PruneViewMembership(doc.Views, doc.Projects)
	return doc, nil
}

// SyncSourceLists is the discovery source list for each forge, as the
// sync sources API sends it.
type SyncSourceLists struct {
	GitHubOrgs          []string `json:"github_orgs"`
	GitLabGroups        []string `json:"gitlab_groups"`
	AzureDevOpsOrgs     []string `json:"azuredevops_orgs"`
	BitbucketWorkspaces []string `json:"bitbucket_workspaces"`
}

// SetSyncSources replaces the forge discovery orgs, groups and workspaces.
// Azure DevOps project filters are kept.
func (c *Commands) SetSyncSources(doc boardconfig.File, src SyncSourceLists) (boardconfig.File, error) {
	if err := c.requireService(); err != nil {
		return doc, err
	}
	doc.Sync.GitHub.Orgs = src.GitHubOrgs
	doc.Sync.GitLab.Groups = src.GitLabGroups
	doc.Sync.AzureDevOps.Orgs = src.AzureDevOpsOrgs
	doc.Sync.Bitbucket.Workspaces = src.BitbucketWorkspaces
	if !doc.Sync.HasSyncSources() {
		return doc, invalid("dashboard.SetSyncSources", "need at least one sync source (github org, gitlab group, azuredevops org, or bitbucket workspace)")
	}
	return doc, nil
}

// SetViews replaces named view definitions (must be non-empty).
func (c *Commands) SetViews(doc boardconfig.File, views []boardconfig.View) (boardconfig.File, error) {
	if err := c.requireService(); err != nil {
		return doc, err
	}
	if len(views) == 0 {
		return doc, invalid("dashboard.SetViews", "views must not be empty")
	}
	doc.Views = boardconfig.PruneViewMembership(views, doc.Projects)
	return doc, nil
}

// SetLocalRoots replaces directories scanned for git checkouts.
// Empty roots are allowed (local mapping then relies on per-project local_path only).
func (c *Commands) SetLocalRoots(doc boardconfig.File, roots []string) (boardconfig.File, error) {
	if err := c.requireService(); err != nil {
		return doc, err
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
func (c *Commands) SetProjectLocalPath(doc boardconfig.File, id, localPath string) (boardconfig.File, error) {
	if err := c.requireService(); err != nil {
		return doc, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return doc, invalid("dashboard.SetProjectLocalPath", "project id required")
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
		return doc, invalid("dashboard.SetProjectLocalPath", "unknown project")
	}
	return doc, nil
}
