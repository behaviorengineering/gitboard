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
	branches := newBranchAccum()

	encoded := strings.ReplaceAll(repo, "/", "%2F")
	if raw, err := g.Run.RunJSON(ctx, "glab", "api", "projects/"+encoded+"?simple=true"); err == nil {
		var meta struct {
			DefaultBranch string `json:"default_branch"`
		}
		if json.Unmarshal(raw, &meta) == nil {
			branches.setDefault(meta.DefaultBranch)
		}
	}

	if raw, err := g.Run.RunJSON(ctx, "glab", "api", "projects/"+encoded+"/repository/branches?per_page=100"); err == nil {
		var heads []struct {
			Name   string `json:"name"`
			WebURL string `json:"web_url"`
			Commit struct {
				CommittedDate string `json:"committed_date"`
				AuthoredDate  string `json:"authored_date"`
			} `json:"commit"`
		}
		if json.Unmarshal(raw, &heads) == nil {
			for _, h := range heads {
				updated := firstNonEmpty(h.Commit.CommittedDate, h.Commit.AuthoredDate)
				branches.addRemote(h.Name, updated, h.WebURL)
			}
		}
	}

	raw, err := g.Run.RunJSON(ctx, "glab", "ci", "list",
		"-R", repo,
		"-P", "20",
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
	for i, pl := range pipelines {
		branches.setCI(pl.Ref, pl.Status, pl.WebURL, pl.UpdatedAt, fmt.Sprintf("%d", pl.ID))
		if i == 0 {
			summary.CI = &CIStatus{
				Status:    pl.Status,
				Ref:       pl.Ref,
				Name:      "pipeline",
				WebURL:    pl.WebURL,
				UpdatedAt: pl.UpdatedAt,
				RunID:     fmt.Sprintf("%d", pl.ID),
			}
		}
	}

	mrRaw, err := g.Run.RunJSON(ctx, "glab", "mr", "list",
		"-R", repo,
		"--output", "json",
	)
	if err != nil {
		if summary.Error == "" {
			summary.Error = err.Error()
		}
		summary.Branches = branches.list()
		return summary, nil
	}
	var mrs []struct {
		IID          int    `json:"iid"`
		SourceBranch string `json:"source_branch"`
		WebURL       string `json:"web_url"`
		UpdatedAt    string `json:"updated_at"`
		HasConflicts bool   `json:"has_conflicts"`
		MergeStatus  string `json:"merge_status"`
		Draft        bool   `json:"draft"`
		WorkInProg   bool   `json:"work_in_progress"`
	}
	if err := json.Unmarshal(mrRaw, &mrs); err != nil {
		if summary.Error == "" {
			summary.Error = err.Error()
		}
		summary.Branches = branches.list()
		return summary, nil
	}
	summary.OpenItems.MergeRequests = len(mrs)
	for _, mr := range mrs {
		branches.setOpenReview(mr.SourceBranch, reviewInfo{
			ID:        mr.IID,
			URL:       mr.WebURL,
			Conflict:  gitlabHasConflict(mr.HasConflicts, mr.MergeStatus),
			UpdatedAt: mr.UpdatedAt,
			Draft:     mr.Draft || mr.WorkInProg,
		})
	}
	summary.Branches = branches.list()
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

// ListGroupRepos lists non-archived projects in a GitLab group (including subgroups).
func (g *GitLab) ListGroupRepos(ctx context.Context, group string) ([]RepoRef, error) {
	group = strings.TrimSpace(group)
	if group == "" {
		return nil, fmt.Errorf("missing gitlab group")
	}
	var out []RepoRef
	for page := 1; page <= 50; page++ {
		raw, err := g.Run.RunJSON(ctx, "glab", "repo", "list",
			"--group", group,
			"--include-subgroups",
			"--archived=false",
			"--per-page", "100",
			"--page", fmt.Sprintf("%d", page),
			"--output", "json",
		)
		if err != nil {
			return nil, err
		}
		var rows []struct {
			PathWithNamespace string `json:"path_with_namespace"`
			Name              string `json:"name"`
			Archived          bool   `json:"archived"`
		}
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, fmt.Errorf("parse glab repo list: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		for _, r := range rows {
			if r.Archived {
				continue
			}
			path := strings.Trim(r.PathWithNamespace, "/")
			parts := strings.Split(path, "/")
			if len(parts) < 2 {
				continue
			}
			name := r.Name
			if name == "" {
				name = parts[len(parts)-1]
			}
			out = append(out, RepoRef{
				Host: config.HostGitLab,
				Path: path,
				Name: name,
			})
		}
		if len(rows) < 100 {
			break
		}
	}
	return out, nil
}
