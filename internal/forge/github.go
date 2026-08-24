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

func (g *GitHub) ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (ProjectSummary, error) {
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
	key := cacheKey(string(p.Host), repo)
	branches := newBranchAccum()

	heads, err := opts.Cache.GetOrLoadHeads(key, opts.HeadsTTL, opts.Fresh, func() (HeadsSnapshot, error) {
		return g.loadHeads(ctx, repo)
	})
	if err != nil {
		if summary.Error == "" {
			summary.Error = err.Error()
		}
	} else {
		if heads.DefaultBranch != "" {
			branches.setDefault(heads.DefaultBranch)
		}
		names := make([]string, 0, len(heads.Heads))
		for _, h := range heads.Heads {
			branches.addRemote(h.Name, h.UpdatedAt, h.WebURL)
			if n := trimBranch(h.Name); n != "" {
				names = append(names, n)
			}
		}
		summary.RemoteNames = names
	}

	if err := g.loadCI(ctx, repo, &summary, branches); err != nil && summary.Error == "" {
		summary.Error = err.Error()
	}
	if err := g.loadOpenReviews(ctx, repo, &summary, branches); err != nil && summary.Error == "" {
		summary.Error = err.Error()
	}

	merged, mergedOK := opts.Cache.GetOrLoadMerged(key, opts.MergedTTL, opts.Fresh, func() ([]MergedReview, error) {
		return g.loadMerged(ctx, repo)
	})
	summary.Merged = merged
	summary.MergedOK = mergedOK
	summary.Branches = branches.list()
	return summary, nil
}

func (g *GitHub) loadHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	var snap HeadsSnapshot
	if raw, err := g.Run.RunJSON(ctx, "gh", "api", "repos/"+repo, "--jq", "{default: .default_branch}"); err == nil {
		var meta struct {
			Default string `json:"default"`
		}
		if json.Unmarshal(raw, &meta) == nil {
			snap.DefaultBranch = meta.Default
		}
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "api", "repos/"+repo+"/branches?per_page=100")
	if err != nil {
		return snap, err
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
		return snap, err
	}
	for _, h := range heads {
		snap.Heads = append(snap.Heads, RemoteHead{
			Name:      h.Name,
			UpdatedAt: firstNonEmpty(h.Commit.Commit.Committer.Date, h.Commit.Commit.Author.Date),
		})
	}
	return snap, nil
}

func (g *GitHub) loadCI(ctx context.Context, repo string, summary *ProjectSummary, branches *branchAccum) error {
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "list",
		"--repo", repo,
		"--limit", "20",
		"--json", "databaseId,status,conclusion,displayTitle,url,headBranch,updatedAt,workflowName",
	)
	if err != nil {
		return err
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
		return err
	}
	for i, r := range runs {
		status := firstNonEmpty(r.Conclusion, r.Status)
		branches.setCI(r.Branch, status, r.URL, r.UpdatedAt, fmt.Sprintf("%d", r.ID))
		if i == 0 {
			summary.CI = &CIStatus{
				Status:     status,
				Conclusion: r.Conclusion,
				Ref:        r.Branch,
				Name:       firstNonEmpty(r.Workflow, r.Title),
				WebURL:     r.URL,
				UpdatedAt:  r.UpdatedAt,
				RunID:      fmt.Sprintf("%d", r.ID),
			}
		}
	}
	return nil
}

func (g *GitHub) loadOpenReviews(ctx context.Context, repo string, summary *ProjectSummary, branches *branchAccum) error {
	prRaw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "open",
		"--json", "number,headRefName,url,mergeable,mergeStateStatus,updatedAt,isDraft",
	)
	if err != nil {
		return err
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
		return err
	}
	summary.OpenItems.PullRequests = len(prs)
	for _, pr := range prs {
		branches.setOpenReview(pr.Head, reviewInfo{
			ID:        pr.Number,
			URL:       pr.URL,
			Conflict:  githubHasConflict(pr.Mergeable, pr.MergeStateStatus),
			UpdatedAt: pr.UpdatedAt,
			Draft:     pr.IsDraft,
		})
	}
	return nil
}

func (g *GitHub) loadMerged(ctx context.Context, repo string) ([]MergedReview, error) {
	raw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "merged",
		"--limit", "50",
		"--json", "number,headRefName,url,mergedAt",
	)
	if err != nil {
		return nil, err
	}
	var prs []struct {
		Number   int    `json:"number"`
		Head     string `json:"headRefName"`
		URL      string `json:"url"`
		MergedAt string `json:"mergedAt"`
	}
	if err := json.Unmarshal(raw, &prs); err != nil {
		return nil, err
	}
	out := make([]MergedReview, 0, len(prs))
	for _, pr := range prs {
		name := trimBranch(pr.Head)
		if name == "" {
			continue
		}
		out = append(out, MergedReview{
			Branch:   name,
			ID:       pr.Number,
			URL:      pr.URL,
			MergedAt: pr.MergedAt,
		})
	}
	return out, nil
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
		return nil, err
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

func baseSummary(p config.Project) ProjectSummary {
	return ProjectSummary{
		ID:      p.ID,
		Label:   p.Label,
		Host:    string(p.Host),
		Path:    strings.Trim(p.Path, "/"),
		Org:     pathOrg(p.Path),
		OpenURL: p.OpenURL(),
	}
}

func pathOrg(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
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
