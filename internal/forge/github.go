package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/config"
)

// GitHub uses the gh CLI.
type GitHub struct {
	Run *cliexec.Runner
}

func NewGitHub(run *cliexec.Runner) *GitHub {
	if run == nil {
		run = cliexec.New()
	}
	return &GitHub{Run: run}
}

func (g *GitHub) AuthStatus(ctx context.Context) (bool, bool, string) {
	if _, err := cliexec.LookPath("gh"); err != nil {
		return false, false, "gh not on PATH"
	}
	_, err := g.Run.Run(ctx, "gh", "auth", "status")
	if err != nil {
		return true, false, strings.TrimPrefix(err.Error(), "gh auth status: ")
	}
	return true, true, ""
}

func (g *GitHub) ProjectSummary(ctx context.Context, p config.Project) (ProjectSummary, error) {
	summary := baseSummary(p)
	installed, authed, detail := g.AuthStatus(ctx)
	if !installed {
		summary.Error = detail
		return summary, nil
	}
	if !authed {
		summary.Error = "gh not authenticated: " + detail
		return summary, nil
	}
	repo := p.Path
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "list",
		"--repo", repo,
		"--limit", "1",
		"--json", "databaseId,status,conclusion,displayTitle,url,headBranch,updatedAt,workflowName",
	)
	if err != nil {
		summary.Error = err.Error()
		return summary, nil
	}
	var runs []struct {
		ID         int64  `json:"databaseId"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		Title      string `json:"displayTitle"`
		URL        string `json:"url"`
		Branch     string `json:"headBranch"`
		UpdatedAt  string `json:"updatedAt"`
		Workflow   string `json:"workflowName"`
	}
	if err := json.Unmarshal(raw, &runs); err != nil {
		summary.Error = err.Error()
		return summary, nil
	}
	if len(runs) > 0 {
		r := runs[0]
		summary.CI = &CIStatus{
			Status:     firstNonEmpty(r.Conclusion, r.Status),
			Conclusion: r.Conclusion,
			Ref:        r.Branch,
			Name:       firstNonEmpty(r.Workflow, r.Title),
			WebURL:     r.URL,
			UpdatedAt:  r.UpdatedAt,
			RunID:      fmt.Sprintf("%d", r.ID),
		}
	}
	prRaw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "open",
		"--json", "number",
	)
	if err != nil {
		if summary.Error == "" {
			summary.Error = err.Error()
		}
		return summary, nil
	}
	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(prRaw, &prs); err != nil {
		if summary.Error == "" {
			summary.Error = err.Error()
		}
		return summary, nil
	}
	summary.OpenItems.PullRequests = len(prs)
	return summary, nil
}

func (g *GitHub) FailedJobs(ctx context.Context, p config.Project, runID string) ([]FailedJob, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("missing run id")
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "view", runID,
		"--repo", p.Path,
		"--json", "jobs",
	)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Jobs []struct {
			DatabaseID int64  `json:"databaseId"`
			Name       string `json:"name"`
			Conclusion string `json:"conclusion"`
			URL        string `json:"url"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	var out []FailedJob
	for _, j := range payload.Jobs {
		if j.Conclusion != "failure" {
			continue
		}
		out = append(out, FailedJob{
			ID:     fmt.Sprintf("github:%d", j.DatabaseID),
			Name:   j.Name,
			WebURL: j.URL,
		})
	}
	return out, nil
}

func (g *GitHub) JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error) {
	runID = strings.TrimSpace(runID)
	jobID = strings.TrimPrefix(strings.TrimSpace(jobID), "github:")
	if runID == "" {
		return "", fmt.Errorf("missing run id")
	}
	args := []string{"run", "view", runID, "--repo", p.Path}
	if jobID != "" {
		args = append(args, "--job", jobID, "--log")
	} else {
		args = append(args, "--log-failed")
	}
	out, err := g.Run.Run(ctx, "gh", args...)
	if err != nil {
		return "", err
	}
	return truncateLog(string(out)), nil
}

func baseSummary(p config.Project) ProjectSummary {
	return ProjectSummary{
		ID:      p.ID,
		Label:   p.Label,
		Host:    string(p.Host),
		OpenURL: p.OpenURL(),
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

func truncateLog(s string) string {
	const max = 96 * 1024
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}
