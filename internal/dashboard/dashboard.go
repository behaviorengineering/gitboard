package dashboard

import (
	"context"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/forge"
)

// Service aggregates project rows via forge CLIs.
type Service struct {
	GitHub *forge.GitHub
	GitLab *forge.GitLab
}

// New returns a dashboard service.
func New(gh *forge.GitHub, gl *forge.GitLab) *Service {
	return &Service{GitHub: gh, GitLab: gl}
}

// Collect builds the dashboard for all configured projects.
func (s *Service) Collect(ctx context.Context, projects []config.Project) forge.Dashboard {
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

	rows := make([]forge.ProjectSummary, len(projects))
	var wg sync.WaitGroup
	for i, p := range projects {
		wg.Add(1)
		go func(i int, p config.Project) {
			defer wg.Done()
			rows[i] = s.summarize(ctx, p)
		}(i, p)
	}
	wg.Wait()
	out.Projects = rows
	return out
}

func (s *Service) summarize(ctx context.Context, p config.Project) forge.ProjectSummary {
	switch p.Host {
	case config.HostGitHub:
		if s.GitHub == nil {
			return forge.ProjectSummary{ID: p.ID, Label: p.Label, Host: string(p.Host), OpenURL: p.OpenURL(), Error: "github client missing"}
		}
		row, _ := s.GitHub.ProjectSummary(ctx, p)
		return row
	default:
		if s.GitLab == nil {
			return forge.ProjectSummary{ID: p.ID, Label: p.Label, Host: string(p.Host), OpenURL: p.OpenURL(), Error: "gitlab client missing"}
		}
		row, _ := s.GitLab.ProjectSummary(ctx, p)
		return row
	}
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
