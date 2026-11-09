package remotegit

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
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

// ready reports a nil receiver or a literal AzureDevOps{} without Run as
// ErrClientMissing, so exported methods fail with an error instead of panicking.
func (a *AzureDevOps) ready() error {
	if a == nil || a.Run == nil {
		return clientMissing("remotegit.AzureDevOps", "azuredevops")
	}
	return nil
}

func (a *AzureDevOps) AuthStatus(ctx context.Context) (bool, bool, string) {
	if err := a.ready(); err != nil {
		return false, false, err.Error()
	}
	return cliAuthStatus(ctx, a.Run, "az not on PATH (install Azure CLI)", "az", "account", "show")
}

func (a *AzureDevOps) host() boardconfig.Host { return boardconfig.HostAzureDevOps }

func (a *AzureDevOps) unauthMsg(detail string) string {
	return "az not authenticated: " + detail
}

func (a *AzureDevOps) ProjectSummary(ctx context.Context, p boardconfig.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, a.ready(), a)
}

// azureRepo is an org/project/repo path split for az arguments.
type azureRepo struct {
	orgURL, project, name string
}

// splitAzurePath splits an org/project/repo path; with more than three
// segments the last one names the repository.
func splitAzurePath(op, repo string) (azureRepo, error) {
	parts, err := splitRepo(boardconfig.HostAzureDevOps, op, repo)
	if err != nil {
		return azureRepo{}, err
	}
	return azureRepo{orgURL: azureOrgURL(parts[0]), project: parts[1], name: parts[len(parts)-1]}, nil
}

// args returns the az flags that address the repository.
func (r azureRepo) args() []string {
	return []string{"--organization", r.orgURL, "--project", r.project, "--repository", r.name}
}

// runJSON runs the az subcommand sub (space separated, for example
// "repos pr list") against the repository with extra flags and -o json.
func (a *AzureDevOps) runJSON(ctx context.Context, r azureRepo, sub string, extra ...string) ([]byte, error) {
	args := append(strings.Fields(sub), r.args()...)
	args = append(args, extra...)
	return a.Run.RunJSON(ctx, "az", append(args, "-o", "json")...)
}

// seedHeads loads the default branch and a bounded list of refs/heads.
func (a *AzureDevOps) seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	const op = "remotegit.AzureDevOps.seedHeads"
	var snap HeadsSnapshot
	r, err := splitAzurePath(op, repo)
	if err != nil {
		return snap, err
	}
	raw, err := a.runJSON(ctx, r, "repos show")
	if err != nil {
		return snap, forgeErr(err, op, "read azure devops default branch failed", repo)
	}
	var meta struct {
		DefaultBranch string `json:"defaultBranch"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return snap, forgeErr(err, op, "parse azure devops repository failed", repo)
	}
	snap.DefaultBranch = strings.TrimPrefix(meta.DefaultBranch, "refs/heads/")

	raw, err = a.runJSON(ctx, r, "repos ref list", "--filter", "heads/")
	if err != nil {
		return snap, forgeErr(err, op, "list azure devops branches failed", repo)
	}
	var refs []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &refs); err != nil {
		return snap, forgeErr(err, op, "parse azure devops branches failed", repo)
	}
	for i, ref := range refs {
		if i >= listPageSize {
			break
		}
		branch := strings.TrimPrefix(ref.Name, "refs/heads/")
		if branch == "" || branch == ref.Name {
			continue
		}
		snap.Heads = append(snap.Heads, RemoteHead{Name: branch})
	}
	return snap, nil
}

func (a *AzureDevOps) loadCI(ctx context.Context, repo string) (CISnapshot, error) {
	const op = "remotegit.AzureDevOps.loadCI"
	r, err := splitAzurePath(op, repo)
	if err != nil {
		return CISnapshot{}, err
	}
	raw, err := a.runJSON(ctx, r, "pipelines runs list", "--top", "20")
	if err != nil {
		// The pipelines extension may be missing. The loader shows this in
		// the row's error and keeps heads and reviews, so the board stays usable.
		return CISnapshot{}, forgeErr(err, op, "list azure devops pipeline runs failed", repo)
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
		return CISnapshot{}, forgeErr(err, op, "parse azure devops pipeline runs failed", repo)
	}
	rows := make([]ciRow, 0, len(runs))
	for _, run := range runs {
		rows = append(rows, ciRow{
			run: CachedCIRun{
				Branch:    strings.TrimPrefix(run.SourceBranch, "refs/heads/"),
				Status:    firstNonEmpty(run.Result, run.State),
				URL:       firstNonEmpty(run.Links.Web.Href, run.Pipeline.URL),
				UpdatedAt: firstNonEmpty(run.FinishedDate, run.CreatedDate),
				RunID:     strconv.Itoa(run.ID),
			},
			conclusion: run.Result,
			name:       firstNonEmpty(run.Pipeline.Name, "pipeline"),
		})
	}
	return newCISnapshot(rows), nil
}

func (a *AzureDevOps) loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error) {
	const op = "remotegit.AzureDevOps.loadOpenReviews"
	r, err := splitAzurePath(op, repo)
	if err != nil {
		return OpenReviewsSnapshot{}, err
	}
	raw, err := a.runJSON(ctx, r, "repos pr list", "--status", "active")
	if err != nil {
		return OpenReviewsSnapshot{}, forgeErr(err, op, "list azure devops open pull requests failed", repo)
	}
	prs, err := decodeAzurePRs(raw)
	if err != nil {
		return OpenReviewsSnapshot{}, forgeErr(err, op, "parse azure devops open pull requests failed", repo)
	}
	return newOpenReviews(boardconfig.HostAzureDevOps, prs), nil
}

func (a *AzureDevOps) loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error) {
	const op = "remotegit.AzureDevOps.loadMerged"
	r, err := splitAzurePath(op, repo)
	if err != nil {
		return nil, err
	}
	raw, err := a.runJSON(ctx, r, "repos pr list", "--status", "completed", "--top", "50")
	if err != nil {
		return nil, forgeErr(err, op, "list azure devops completed pull requests failed", repo)
	}
	prs, err := decodeAzurePRs(raw)
	if err != nil {
		return nil, forgeErr(err, op, "parse azure devops completed pull requests failed", repo)
	}
	return mergedReviews(prs), nil
}

// MergedForBranch looks up completed PRs whose source branch matches branch.
func (a *AzureDevOps) MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error) {
	const op = "remotegit.AzureDevOps.MergedForBranch"
	if err := a.ready(); err != nil {
		return nil, err
	}
	branch = trimBranch(branch)
	if err := forgeBranchNameOK(branch); err != nil {
		return nil, err
	}
	r, err := splitAzurePath(op, repo)
	if err != nil {
		return nil, err
	}
	raw, err := a.runJSON(ctx, r, "repos pr list",
		"--status", "completed", "--source-branch", branch, "--top", "1")
	if err != nil {
		return nil, forgeErr(err, op, "look up azure devops completed pull request failed", repo).With("branch", branch)
	}
	prs, err := decodeAzurePRs(raw)
	if err != nil {
		return nil, forgeErr(err, op, "parse azure devops completed pull request failed", repo).With("branch", branch)
	}
	return mergedReviews(prs), nil
}

func (a *AzureDevOps) FailedJobs(_ context.Context, _ boardconfig.Project, runID string) ([]board.FailedJob, error) {
	const op = "remotegit.AzureDevOps.FailedJobs"
	if err := a.ready(); err != nil {
		return nil, err
	}
	if err := requireArg(op, "run id", runID); err != nil {
		return nil, err
	}
	// Detailed job listing via az varies by pipeline type; keep empty rather than panic.
	return nil, nil
}

func (a *AzureDevOps) JobLog(_ context.Context, _ boardconfig.Project, runID, _ string) (string, error) {
	const op = "remotegit.AzureDevOps.JobLog"
	if err := a.ready(); err != nil {
		return "", err
	}
	if err := requireArg(op, "run id", runID); err != nil {
		return "", err
	}
	return "", gberr.New(gberr.CodeInvalid, op, "azure devops job logs are not supported yet")
}

// ListOrgRepos lists repositories under an Azure DevOps organization.
// Optional project names may be passed as org/project; bare org lists every project.
func (a *AzureDevOps) ListOrgRepos(ctx context.Context, org string) ([]RepoRef, error) {
	const op = "remotegit.AzureDevOps.ListOrgRepos"
	if err := a.ready(); err != nil {
		return nil, err
	}
	org = strings.TrimSpace(org)
	if err := requireArg(op, "azure devops org", org); err != nil {
		return nil, err
	}
	orgURL := azureOrgURL(org)
	projects, err := a.listProjects(ctx, op, org, orgURL)
	if err != nil {
		return nil, err
	}
	var out []RepoRef
	for _, project := range projects {
		if err := ctx.Err(); err != nil {
			return nil, gberr.Wrap(err, gberr.CodeCanceled, op, "repository listing cancelled")
		}
		repos, err := a.listReposInProject(ctx, op, org, orgURL, project)
		if err != nil {
			return nil, err
		}
		out = append(out, repos...)
	}
	return out, nil
}

func (a *AzureDevOps) listProjects(ctx context.Context, op, org, orgURL string) ([]string, error) {
	raw, err := a.Run.RunJSON(ctx, "az", "devops", "project", "list",
		"--organization", orgURL,
		"-o", "json",
	)
	if err != nil {
		return nil, forgeErr(err, op, "list azure devops projects failed", org)
	}
	type projectRow struct {
		Name string `json:"name"`
	}
	var payload struct {
		Value []projectRow `json:"value"`
	}
	var rows []projectRow
	if err := json.Unmarshal(raw, &payload); err == nil {
		rows = payload.Value
	} else if err2 := json.Unmarshal(raw, &rows); err2 != nil {
		// Some az versions return a bare array rather than {"value": [...]}.
		return nil, forgeErr(errors.Join(err, err2), op, "parse azure devops projects failed", org)
	}
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		if n := strings.TrimSpace(r.Name); n != "" {
			names = append(names, n)
		}
	}
	return names, nil
}

func (a *AzureDevOps) listReposInProject(ctx context.Context, op, org, orgURL, project string) ([]RepoRef, error) {
	raw, err := a.Run.RunJSON(ctx, "az", "repos", "list",
		"--organization", orgURL,
		"--project", project,
		"-o", "json",
	)
	if err != nil {
		return nil, forgeErr(err, op, "list azure devops repositories failed", org).With("project", project)
	}
	var rows []struct {
		Name       string `json:"name"`
		IsDisabled bool   `json:"isDisabled"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, forgeErr(err, op, "parse azure devops repositories failed", org).With("project", project)
	}
	out := make([]RepoRef, 0, len(rows))
	for _, r := range rows {
		name := strings.TrimSpace(r.Name)
		if r.IsDisabled || name == "" {
			continue
		}
		if ref, ok := newRepoRef(boardconfig.HostAzureDevOps, org+"/"+project+"/"+name, name); ok {
			out = append(out, ref)
		}
	}
	return out, nil
}

func decodeAzurePRs(raw []byte) ([]forgePR, error) {
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
	out := make([]forgePR, 0, len(rows))
	for _, r := range rows {
		merge := strings.ToLower(strings.TrimSpace(r.MergeStatus))
		out = append(out, forgePR{
			Branch:    strings.TrimPrefix(r.SourceRefName, "refs/heads/"),
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
