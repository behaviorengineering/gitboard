package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/config"
)

// GitLab uses the glab CLI.
type GitLab struct {
	Run *cliexec.Runner
}

func NewGitLab(run *cliexec.Runner) *GitLab {
	if run == nil {
		run = cliexec.New()
	}
	return &GitLab{Run: run}
}

func (g *GitLab) AuthStatus(ctx context.Context) (bool, bool, string) {
	if _, err := cliexec.LookPath("glab"); err != nil {
		return false, false, "glab not on PATH (brew install glab)"
	}
	_, err := g.Run.Run(ctx, "glab", "auth", "status")
	if err != nil {
		return true, false, strings.TrimPrefix(err.Error(), "glab auth status: ")
	}
	return true, true, ""
}

func (g *GitLab) ProjectSummary(ctx context.Context, p config.Project) (ProjectSummary, error) {
	summary := baseSummary(p)
	installed, authed, detail := g.AuthStatus(ctx)
	if !installed {
		summary.Error = detail
		return summary, nil
	}
	if !authed {
		summary.Error = "glab not authenticated: " + detail
		return summary, nil
	}
	repo := p.Path
	raw, err := g.Run.RunJSON(ctx, "glab", "ci", "list",
		"-R", repo,
		"-P", "1",
		"--output", "json",
	)
	if err != nil {
		summary.Error = err.Error()
		return summary, nil
	}
	var pipelines []struct {
		ID        int    `json:"id"`
		Status    string `json:"status"`
		Ref       string `json:"ref"`
		SHA       string `json:"sha"`
		WebURL    string `json:"web_url"`
		UpdatedAt string `json:"updated_at"`
	}
	if err := json.Unmarshal(raw, &pipelines); err != nil {
		summary.Error = err.Error()
		return summary, nil
	}
	if len(pipelines) > 0 {
		pl := pipelines[0]
		sha := pl.SHA
		if len(sha) > 8 {
			sha = sha[:8]
		}
		summary.CI = &CIStatus{
			Status:    pl.Status,
			Ref:       pl.Ref,
			Name:      "pipeline",
			WebURL:    pl.WebURL,
			UpdatedAt: pl.UpdatedAt,
			RunID:     fmt.Sprintf("%d", pl.ID),
		}
		_ = sha
	}
	mrRaw, err := g.Run.RunJSON(ctx, "glab", "mr", "list",
		"-R", repo,
		"--state", "opened",
		"--output", "json",
	)
	if err != nil {
		if summary.Error == "" {
			summary.Error = err.Error()
		}
		return summary, nil
	}
	var mrs []json.RawMessage
	if err := json.Unmarshal(mrRaw, &mrs); err != nil {
		if summary.Error == "" {
			summary.Error = err.Error()
		}
		return summary, nil
	}
	summary.OpenItems.MergeRequests = len(mrs)
	return summary, nil
}

func (g *GitLab) FailedJobs(ctx context.Context, p config.Project, pipelineID string) ([]FailedJob, error) {
	if strings.TrimSpace(pipelineID) == "" {
		return nil, fmt.Errorf("missing pipeline id")
	}
	raw, err := g.Run.RunJSON(ctx, "glab", "ci", "view", pipelineID,
		"-R", p.Path,
		"--output", "json",
	)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Jobs []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Stage  string `json:"stage"`
			Status string `json:"status"`
			WebURL string `json:"web_url"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	var out []FailedJob
	for _, j := range payload.Jobs {
		if j.Status != "failed" {
			continue
		}
		out = append(out, FailedJob{
			ID:     fmt.Sprintf("gitlab:%d", j.ID),
			Name:   j.Name,
			Stage:  j.Stage,
			WebURL: j.WebURL,
		})
	}
	return out, nil
}

func (g *GitLab) JobLog(ctx context.Context, p config.Project, _, jobID string) (string, error) {
	jobID = strings.TrimPrefix(strings.TrimSpace(jobID), "gitlab:")
	if jobID == "" {
		return "", fmt.Errorf("missing job id")
	}
	out, err := g.Run.Run(ctx, "glab", "ci", "trace", jobID, "-R", p.Path)
	if err != nil {
		return "", err
	}
	return truncateLog(string(out)), nil
}
