package dashboard

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/forge"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

// Service aggregates project rows via forge CLIs and optional local git.
type Service struct {
	GitHub *forge.GitHub
	GitLab *forge.GitLab
	Local  *localgit.Inspector
	Cache  *forge.TTLCache
}

// New returns a dashboard service with an in-memory upstream cache.
func New(gh *forge.GitHub, gl *forge.GitLab, local *localgit.Inspector) *Service {
	return &Service{GitHub: gh, GitLab: gl, Local: local, Cache: forge.NewTTLCache()}
}

// Collect builds the dashboard for all configured projects.
// When fresh is true, upstream TTL caches are bypassed for this request.
func (s *Service) Collect(ctx context.Context, doc config.File, fresh bool) forge.Dashboard {
	projects := doc.Projects
	out := forge.Dashboard{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if s.GitHub != nil {
		installed, authed, detail := s.GitHub.AuthStatus(ctx)
		out.Tooling.GitHub.Installed = installed
		out.Tooling.GitHub.Authed = authed
		out.Tooling.GitHub.Detail = detail
	}
	if s.GitLab != nil {
		installed, authed, detail := s.GitLab.AuthStatus(ctx)
		out.Tooling.GitLab.Installed = installed
		out.Tooling.GitLab.Authed = authed
		out.Tooling.GitLab.Detail = detail
	}

	var disc localgit.Discovery
	if s.Local != nil && len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}

	opts := forge.SummaryOpts{
		Fresh:     fresh,
		Cache:     s.Cache,
		HeadsTTL:  time.Duration(doc.EffectiveHeadsSeconds()) * time.Second,
		MergedTTL: time.Duration(doc.EffectiveMergedSeconds()) * time.Second,
	}

	rows := make([]forge.ProjectSummary, len(projects))
	var wg sync.WaitGroup
	for i, p := range projects {
		wg.Add(1)
		go func(i int, p config.Project) {
			defer wg.Done()
			row := s.summarize(ctx, p, opts)
			row.Local = s.attachLocal(ctx, p, disc)
			forge.EnrichPruneHints(&row)
			rows[i] = row
		}(i, p)
	}
	wg.Wait()
	out.Projects = rows
	return out
}

func (s *Service) attachLocal(ctx context.Context, p config.Project, disc localgit.Discovery) *forge.LocalStatus {
	if s.Local == nil {
		return nil
	}
	path, ok := localgit.ResolvePath(p.LocalPath, string(p.Host), p.Path, disc)
	if !ok {
		if len(disc.ByKey) == 0 && strings.TrimSpace(p.LocalPath) == "" {
			return nil
		}
		return &forge.LocalStatus{Mapped: false}
	}
	st := s.Local.InspectPath(ctx, path)
	s.Local.EnrichDefault(ctx, &st)
	return toForgeLocal(st)
}

func toForgeLocal(st localgit.Status) *forge.LocalStatus {
	out := &forge.LocalStatus{
		Mapped:        st.Mapped,
		Path:          st.Path,
		Error:         st.Error,
		Branch:        st.Branch,
		Detached:      st.Detached,
		Dirty:         st.Dirty,
		Ahead:         st.Ahead,
		Behind:        st.Behind,
		Upstream:      st.Upstream,
		DefaultBranch: st.DefaultBranch,
		DefaultBehind: st.DefaultBehind,
		DefaultAhead:  st.DefaultAhead,
	}
	for _, wt := range st.Worktrees {
		if wt.Bare {
			continue
		}
		out.Worktrees = append(out.Worktrees, forge.LocalWorktree{
			Path:     wt.Path,
			Branch:   wt.Branch,
			Detached: wt.Detached,
			Bare:     wt.Bare,
			Main:     wt.Main,
			Dirty:    wt.Dirty,
			Ahead:    wt.Ahead,
			Behind:   wt.Behind,
			Upstream: wt.Upstream,
		})
	}
	return out
}

func (s *Service) summarize(ctx context.Context, p config.Project, opts forge.SummaryOpts) forge.ProjectSummary {
	switch p.Host {
	case config.HostGitHub:
		if s.GitHub == nil {
			return forge.ProjectSummary{
				ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
				OpenURL: p.OpenURL(), Error: "github client missing",
			}
		}
		row, _ := s.GitHub.ProjectSummary(ctx, p, opts)
		return row
	default:
		if s.GitLab == nil {
			return forge.ProjectSummary{
				ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
				OpenURL: p.OpenURL(), Error: "gitlab client missing",
			}
		}
		row, _ := s.GitLab.ProjectSummary(ctx, p, opts)
		return row
	}
}

func forgeOrg(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// FindProject returns a project by id.
func FindProject(projects []config.Project, id string) (config.Project, bool) {
	for _, p := range projects {
		if p.ID == id {
			return p, true
		}
	}
	return config.Project{}, false
}

// ClientFor returns the forge client for a project host.
func ClientFor(s *Service, p config.Project) forge.Client {
	switch p.Host {
	case config.HostGitHub:
		return s.GitHub
	default:
		return s.GitLab
	}
}
