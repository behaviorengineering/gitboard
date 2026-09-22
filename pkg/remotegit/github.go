package remotegit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
)

// GitHub uses the gh CLI.
type GitHub struct {
	Run cliexec.Exec
}

func NewGitHub(run cliexec.Exec) *GitHub {
	if run == nil {
		panic("remotegit.NewGitHub: Exec is required")
	}
	return &GitHub{Run: run}
}

func (g *GitHub) AuthStatus(ctx context.Context) (bool, bool, string) {
	if _, err := g.Run.LookPath("gh"); err != nil {
		return false, false, "gh not on PATH"
	}
	_, err := g.Run.Run(ctx, "gh", "auth", "status")
	if err != nil {
		return true, false, strings.TrimPrefix(err.Error(), "gh auth status: ")
	}
	return true, true, ""
}

func (g *GitHub) authCacheID() string { return "github" }

func (g *GitHub) unauthMsg(detail string) string {
	return "gh not authenticated: " + detail
}

func (g *GitHub) ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, g)
}

// seedHeads loads the default branch and the first page of remote branches.
// This replaces the old 50-page census; open PR branches are added separately
// by loadOpenReviews.
func (g *GitHub) seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	var snap HeadsSnapshot
	raw, err := g.Run.RunJSON(ctx, "gh", "api", "repos/"+repo, "--jq", "{default: .default_branch}")
	if err != nil {
		return snap, fmt.Errorf("github default branch: %w", err)
	}
	var meta struct {
		Default string `json:"default"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return snap, fmt.Errorf("parse github default branch: %w", err)
	}
	snap.DefaultBranch = meta.Default

	const perPage = 100
	path := fmt.Sprintf("repos/%s/branches?per_page=%d&page=1", repo, perPage)
	raw, err = g.Run.RunJSON(ctx, "gh", "api", path)
	if err != nil {
		return snap, fmt.Errorf("github branches page 1: %w", err)
	}
	var heads []struct {
		Name   string `json:"name"`
		Commit struct {
			Commit struct {
				Committer struct {
					Date string `json:"date"`
				} `json:"committer"`
				Author struct {
					Date string `json:"date"`
				} `json:"author"`
			} `json:"commit"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(raw, &heads); err != nil {
		return snap, fmt.Errorf("parse github branches: %w", err)
	}
	for _, h := range heads {
		snap.Heads = append(snap.Heads, RemoteHead{
			Name:      h.Name,
			UpdatedAt: firstNonEmpty(h.Commit.Commit.Committer.Date, h.Commit.Commit.Author.Date),
		})
	}
	return snap, nil
}

func (g *GitHub) loadCI(ctx context.Context, repo string) (CISnapshot, error) {
	var snap CISnapshot
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "list",
		"--repo", repo,
		"--limit", "20",
		"--json", "databaseId,status,conclusion,displayTitle,url,headBranch,updatedAt,workflowName",
	)
	if err != nil {
		return snap, err
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
		return snap, err
	}
	for i, r := range runs {
		status := firstNonEmpty(r.Conclusion, r.Status)
		runID := fmt.Sprintf("%d", r.ID)
		snap.Runs = append(snap.Runs, CachedCIRun{
			Branch: r.Branch, Status: status, URL: r.URL, UpdatedAt: r.UpdatedAt, RunID: runID,
		})
		if i == 0 {
			snap.Latest = &board.CIStatus{
				Status:     status,
				Conclusion: r.Conclusion,
				Ref:        r.Branch,
				Name:       firstNonEmpty(r.Workflow, r.Title),
				WebURL:     r.URL,
				UpdatedAt:  r.UpdatedAt,
				RunID:      runID,
			}
		}
	}
	return snap, nil
}

func (g *GitHub) loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error) {
	var snap OpenReviewsSnapshot
	prRaw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "open",
		"--json", "number,headRefName,url,mergeable,mergeStateStatus,updatedAt,isDraft",
	)
	if err != nil {
		return snap, err
	}
	var prs []struct {
		Number           int    `json:"number"`
		Head             string `json:"headRefName"`
		URL              string `json:"url"`
		Mergeable        string `json:"mergeable"`
		MergeStateStatus string `json:"mergeStateStatus"`
		UpdatedAt        string `json:"updatedAt"`
		IsDraft          bool   `json:"isDraft"`
	}
	if err := json.Unmarshal(prRaw, &prs); err != nil {
		return snap, err
	}
	snap.PullRequests = len(prs)
	for _, pr := range prs {
		snap.Reviews = append(snap.Reviews, CachedOpenReview{
			Branch:    pr.Head,
			ID:        pr.Number,
			URL:       pr.URL,
			Conflict:  githubHasConflict(pr.Mergeable, pr.MergeStateStatus),
			UpdatedAt: pr.UpdatedAt,
			Draft:     pr.IsDraft,
		})
	}
	return snap, nil
}

func (g *GitHub) loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error) {
	raw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "merged",
		"--limit", "50",
		"--json", "number,headRefName,url,mergedAt",
	)
	if err != nil {
		return nil, err
	}
	return decodeGitHubMerged(raw)
}

// MergedForBranch looks up merged PRs whose head branch matches branch.
func (g *GitHub) MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error) {
	if g == nil {
		return nil, fmt.Errorf("github client missing")
	}
	branch = trimBranch(branch)
	if err := forgeBranchNameOK(branch); err != nil {
		return nil, err
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "merged",
		"--head", branch,
		"--limit", "1",
		"--json", "number,headRefName,url,mergedAt",
	)
	if err != nil {
		return nil, fmt.Errorf("github merged for %s: %w", branch, err)
	}
	return decodeGitHubMerged(raw)
}

func decodeGitHubMerged(raw []byte) ([]board.MergedReview, error) {
	var prs []struct {
		Number   int    `json:"number"`
		Head     string `json:"headRefName"`
		URL      string `json:"url"`
		MergedAt string `json:"mergedAt"`
	}
	if err := json.Unmarshal(raw, &prs); err != nil {
		return nil, err
	}
	out := make([]board.MergedReview, 0, len(prs))
	for _, pr := range prs {
		name := trimBranch(pr.Head)
		if name == "" {
			continue
		}
		out = append(out, board.MergedReview{
			Branch:   name,
			ID:       pr.Number,
			URL:      pr.URL,
			MergedAt: pr.MergedAt,
		})
	}
	return out, nil
}

func (g *GitHub) FailedJobs(ctx context.Context, p config.Project, runID string) ([]board.FailedJob, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("missing run id")
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "view", runID,
		"--repo", p.Path,
		"--json", "jobs",
	)
	if err != nil {
		return nil, fmt.Errorf("github failed jobs: %w", err)
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
		return nil, fmt.Errorf("parse github jobs: %w", err)
	}
	var out []board.FailedJob
	for _, j := range payload.Jobs {
		if j.Conclusion != "failure" {
			continue
		}
		out = append(out, board.FailedJob{
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
		return "", fmt.Errorf("github job log: %w", err)
	}
	return truncateLog(string(out)), nil
}

// ListOrgRepos lists non-archived repositories for a GitHub org or user.
func (g *GitHub) ListOrgRepos(ctx context.Context, org string) ([]RepoRef, error) {
	org = strings.TrimSpace(org)
	if org == "" {
		return nil, fmt.Errorf("missing github org")
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "repo", "list", org,
		"--limit", "1000",
		"--no-archived",
		"--json", "nameWithOwner,name,isArchived",
	)
	if err != nil {
		return nil, fmt.Errorf("github list repos: %w", err)
	}
	var rows []struct {
		NameWithOwner string `json:"nameWithOwner"`
		Name          string `json:"name"`
		IsArchived    bool   `json:"isArchived"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse gh repo list: %w", err)
	}
	out := make([]RepoRef, 0, len(rows))
	for _, r := range rows {
		if r.IsArchived {
			continue
		}
		path := strings.Trim(r.NameWithOwner, "/")
		name := r.Name
		if name == "" {
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}
		out = append(out, RepoRef{
			Host: config.HostGitHub,
			Path: path,
			Name: name,
		})
	}
	return out, nil
}
