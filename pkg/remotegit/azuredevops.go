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

// AzureDevOps uses the Azure CLI (`az`) with DevOps extensions.
// Project Path format is org/project/repo (three segments preferred).
type AzureDevOps struct {
	Run cliexec.Exec
}

// NewAzureDevOps returns an Azure DevOps forge client.
func NewAzureDevOps(run cliexec.Exec) *AzureDevOps {
	if run == nil {
		panic("remotegit.NewAzureDevOps: Exec is required")
	}
	return &AzureDevOps{Run: run}
}

func (a *AzureDevOps) AuthStatus(ctx context.Context) (bool, bool, string) {
	if _, err := a.Run.LookPath("az"); err != nil {
		return false, false, "az not on PATH (install Azure CLI)"
	}
	_, err := a.Run.Run(ctx, "az", "account", "show")
	if err != nil {
		return true, false, strings.TrimPrefix(err.Error(), "az account show: ")
	}
	return true, true, ""
}

func (a *AzureDevOps) authCacheID() string { return "azuredevops" }

func (a *AzureDevOps) unauthMsg(detail string) string {
	return "az not authenticated: " + detail
}

func (a *AzureDevOps) ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, a)
}

// seedHeads loads the default branch and a bounded list of refs/heads.
func (a *AzureDevOps) seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	var snap HeadsSnapshot
	org, project, name, err := splitAzurePath(repo)
	if err != nil {
		return snap, err
	}
	orgURL := azureOrgURL(org)

	raw, err := a.Run.RunJSON(ctx, "az", "repos", "show",
		"--organization", orgURL,
		"--project", project,
		"--repository", name,
		"-o", "json",
	)
	if err != nil {
		return snap, fmt.Errorf("azuredevops default branch: %w", err)
	}
	var meta struct {
		DefaultBranch string `json:"defaultBranch"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return snap, fmt.Errorf("parse azuredevops repo: %w", err)
	}
	snap.DefaultBranch = strings.TrimPrefix(meta.DefaultBranch, "refs/heads/")

	raw, err = a.Run.RunJSON(ctx, "az", "repos", "ref", "list",
		"--organization", orgURL,
		"--project", project,
		"--repository", name,
		"--filter", "heads/",
		"-o", "json",
	)
	if err != nil {
		return snap, fmt.Errorf("azuredevops branches: %w", err)
	}
	var refs []struct {
		Name    string `json:"name"`
		Creator struct {
			DisplayName string `json:"displayName"`
		} `json:"creator"`
	}
	if err := json.Unmarshal(raw, &refs); err != nil {
		return snap, fmt.Errorf("parse azuredevops refs: %w", err)
	}
	const maxHeads = 100
	for i, r := range refs {
		if i >= maxHeads {
			break
		}
		branch := strings.TrimPrefix(r.Name, "refs/heads/")
		if branch == "" || branch == r.Name {
			continue
		}
		snap.Heads = append(snap.Heads, RemoteHead{Name: branch})
	}
	return snap, nil
}

func (a *AzureDevOps) loadCI(ctx context.Context, repo string) (CISnapshot, error) {
	var snap CISnapshot
	org, project, name, err := splitAzurePath(repo)
	if err != nil {
		return snap, nil
	}
	raw, err := a.Run.RunJSON(ctx, "az", "pipelines", "runs", "list",
		"--organization", azureOrgURL(org),
		"--project", project,
		"--repository", name,
		"--top", "20",
		"-o", "json",
	)
	if err != nil {
		// Pipelines extension may be missing; keep board usable.
		return snap, nil
	}
	var runs []struct {
		ID           int    `json:"id"`
		State        string `json:"state"`
		Result       string `json:"result"`
		SourceBranch string `json:"sourceBranch"`
		FinishedDate string `json:"finishedDate"`
		CreatedDate  string `json:"createdDate"`
		Pipeline     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pipeline"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(raw, &runs); err != nil {
		return snap, nil
	}
	for i, r := range runs {
		status := firstNonEmpty(r.Result, r.State)
		branch := strings.TrimPrefix(r.SourceBranch, "refs/heads/")
		runID := fmt.Sprintf("%d", r.ID)
		url := firstNonEmpty(r.Links.Web.Href, r.Pipeline.URL)
		updated := firstNonEmpty(r.FinishedDate, r.CreatedDate)
		snap.Runs = append(snap.Runs, CachedCIRun{
			Branch: branch, Status: status, URL: url, UpdatedAt: updated, RunID: runID,
		})
		if i == 0 {
			snap.Latest = &board.CIStatus{
				Status:     status,
				Conclusion: r.Result,
				Ref:        branch,
				Name:       firstNonEmpty(r.Pipeline.Name, "pipeline"),
				WebURL:     url,
				UpdatedAt:  updated,
				RunID:      runID,
			}
		}
	}
	return snap, nil
}

func (a *AzureDevOps) loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error) {
	var snap OpenReviewsSnapshot
	org, project, name, err := splitAzurePath(repo)
	if err != nil {
		return snap, nil
	}
	raw, err := a.Run.RunJSON(ctx, "az", "repos", "pr", "list",
		"--organization", azureOrgURL(org),
		"--project", project,
		"--repository", name,
		"--status", "active",
		"-o", "json",
	)
	if err != nil {
		return snap, nil
	}
	prs, err := decodeAzurePRs(raw)
	if err != nil {
		return snap, nil
	}
	snap.PullRequests = len(prs)
	for _, pr := range prs {
		snap.Reviews = append(snap.Reviews, CachedOpenReview{
			Branch:    pr.Branch,
			ID:        pr.ID,
			URL:       pr.URL,
			Conflict:  pr.Conflict,
			UpdatedAt: pr.UpdatedAt,
			Draft:     pr.Draft,
		})
	}
	return snap, nil
}

func (a *AzureDevOps) loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error) {
	org, project, name, err := splitAzurePath(repo)
	if err != nil {
		return nil, nil
	}
	raw, err := a.Run.RunJSON(ctx, "az", "repos", "pr", "list",
		"--organization", azureOrgURL(org),
		"--project", project,
		"--repository", name,
		"--status", "completed",
		"--top", "50",
		"-o", "json",
	)
	if err != nil {
		return nil, nil
	}
	prs, err := decodeAzurePRs(raw)
	if err != nil {
		return nil, nil
	}
	out := make([]board.MergedReview, 0, len(prs))
	for _, pr := range prs {
		branch := trimBranch(pr.Branch)
		if branch == "" {
			continue
		}
		out = append(out, board.MergedReview{
			Branch:   branch,
			ID:       pr.ID,
			URL:      pr.URL,
			MergedAt: pr.MergedAt,
		})
	}
	return out, nil
}

// MergedForBranch looks up completed PRs whose source branch matches branch.
func (a *AzureDevOps) MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error) {
	if a == nil {
		return nil, fmt.Errorf("azuredevops client missing")
	}
	branch = trimBranch(branch)
	if err := forgeBranchNameOK(branch); err != nil {
		return nil, err
	}
	org, project, name, err := splitAzurePath(repo)
	if err != nil {
		return nil, err
	}
	raw, err := a.Run.RunJSON(ctx, "az", "repos", "pr", "list",
		"--organization", azureOrgURL(org),
		"--project", project,
		"--repository", name,
		"--status", "completed",
		"--source-branch", branch,
		"--top", "1",
		"-o", "json",
	)
	if err != nil {
		return nil, fmt.Errorf("azuredevops merged for %s: %w", branch, err)
	}
	prs, err := decodeAzurePRs(raw)
	if err != nil {
		return nil, fmt.Errorf("azuredevops merged for %s: %w", branch, err)
	}
	out := make([]board.MergedReview, 0, len(prs))
	for _, pr := range prs {
		b := trimBranch(pr.Branch)
		if b == "" {
			continue
		}
		out = append(out, board.MergedReview{
			Branch:   b,
			ID:       pr.ID,
			URL:      pr.URL,
			MergedAt: pr.MergedAt,
		})
	}
	return out, nil
}

func (a *AzureDevOps) FailedJobs(ctx context.Context, p config.Project, runID string) ([]board.FailedJob, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("missing run id")
	}
	// Detailed job listing via az varies by pipeline type; keep empty rather than panic.
	return nil, nil
}

func (a *AzureDevOps) JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error) {
	if strings.TrimSpace(runID) == "" {
		return "", fmt.Errorf("missing run id")
	}
	return "", fmt.Errorf("azuredevops job logs not supported yet")
}

// ListOrgRepos lists repositories under an Azure DevOps organization.
// Optional project names may be passed as org/project; bare org lists every project.
func (a *AzureDevOps) ListOrgRepos(ctx context.Context, org string) ([]RepoRef, error) {
	org = strings.TrimSpace(org)
	if org == "" {
		return nil, fmt.Errorf("missing azuredevops org")
	}
	orgURL := azureOrgURL(org)
	projects, err := a.listProjects(ctx, orgURL)
	if err != nil {
		return nil, err
	}
	var out []RepoRef
	for _, project := range projects {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		repos, err := a.listReposInProject(ctx, org, orgURL, project)
		if err != nil {
			return nil, err
		}
		out = append(out, repos...)
	}
	return out, nil
}

func (a *AzureDevOps) listProjects(ctx context.Context, orgURL string) ([]string, error) {
	raw, err := a.Run.RunJSON(ctx, "az", "devops", "project", "list",
		"--organization", orgURL,
		"-o", "json",
	)
	if err != nil {
		return nil, fmt.Errorf("azuredevops list projects: %w", err)
	}
	var payload struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		// Some az versions return a bare array.
		var rows []struct {
			Name string `json:"name"`
		}
		if err2 := json.Unmarshal(raw, &rows); err2 != nil {
			return nil, fmt.Errorf("parse azuredevops projects: %w", err)
		}
		names := make([]string, 0, len(rows))
		for _, r := range rows {
			if n := strings.TrimSpace(r.Name); n != "" {
				names = append(names, n)
			}
		}
		return names, nil
	}
	names := make([]string, 0, len(payload.Value))
	for _, r := range payload.Value {
		if n := strings.TrimSpace(r.Name); n != "" {
			names = append(names, n)
		}
	}
	return names, nil
}

func (a *AzureDevOps) listReposInProject(ctx context.Context, org, orgURL, project string) ([]RepoRef, error) {
	raw, err := a.Run.RunJSON(ctx, "az", "repos", "list",
		"--organization", orgURL,
		"--project", project,
		"-o", "json",
	)
	if err != nil {
		return nil, fmt.Errorf("azuredevops list repos: %w", err)
	}
	var rows []struct {
		Name       string `json:"name"`
		IsDisabled bool   `json:"isDisabled"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse azuredevops repos: %w", err)
	}
	out := make([]RepoRef, 0, len(rows))
	for _, r := range rows {
		if r.IsDisabled {
			continue
		}
		name := strings.TrimSpace(r.Name)
		if name == "" {
			continue
		}
		out = append(out, RepoRef{
			Host: config.HostAzureDevOps,
			Path: org + "/" + project + "/" + name,
			Name: name,
		})
	}
	return out, nil
}

type azurePR struct {
	Branch    string
	ID        int
	URL       string
	Conflict  bool
	UpdatedAt string
	MergedAt  string
	Draft     bool
}

func decodeAzurePRs(raw []byte) ([]azurePR, error) {
	var rows []struct {
		PullRequestID int    `json:"pullRequestId"`
		Status        string `json:"status"`
		IsDraft       bool   `json:"isDraft"`
		MergeStatus   string `json:"mergeStatus"`
		SourceRefName string `json:"sourceRefName"`
		ClosedDate    string `json:"closedDate"`
		CreationDate  string `json:"creationDate"`
		URL           string `json:"url"`
		Links         struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	out := make([]azurePR, 0, len(rows))
	for _, r := range rows {
		branch := strings.TrimPrefix(r.SourceRefName, "refs/heads/")
		merge := strings.ToLower(strings.TrimSpace(r.MergeStatus))
		out = append(out, azurePR{
			Branch:    branch,
			ID:        r.PullRequestID,
			URL:       firstNonEmpty(r.Links.Web.Href, r.URL),
			Conflict:  merge == "conflicts" || merge == "failure",
			UpdatedAt: firstNonEmpty(r.ClosedDate, r.CreationDate),
			MergedAt:  r.ClosedDate,
			Draft:     r.IsDraft,
		})
	}
	return out, nil
}

func azureOrgURL(org string) string {
	org = strings.Trim(strings.TrimSpace(org), "/")
	if strings.HasPrefix(org, "https://") || strings.HasPrefix(org, "http://") {
		return org
	}
	return "https://dev.azure.com/" + org
}

func splitAzurePath(repo string) (org, project, name string, err error) {
	parts := strings.Split(strings.Trim(repo, "/"), "/")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("azuredevops path must be org/project/repo")
	}
	for _, p := range parts {
		if p == "" {
			return "", "", "", fmt.Errorf("azuredevops path must be org/project/repo")
		}
	}
	return parts[0], parts[1], parts[len(parts)-1], nil
}
