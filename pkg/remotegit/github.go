package remotegit

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
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

// ready reports a nil receiver or a literal GitHub{} without Run as
// ErrClientMissing, so exported methods fail with an error instead of panicking.
func (g *GitHub) ready() error {
	if g == nil || g.Run == nil {
		return clientMissing("remotegit.GitHub", "github")
	}
	return nil
}

func (g *GitHub) AuthStatus(ctx context.Context) (bool, bool, string) {
	if err := g.ready(); err != nil {
		return false, false, err.Error()
	}
	return cliAuthStatus(ctx, g.Run, "gh not on PATH", "gh", "auth", "status")
}

func (g *GitHub) host() boardconfig.Host { return boardconfig.HostGitHub }

func (g *GitHub) unauthMsg(detail string) string {
	return "gh not authenticated: " + detail
}

func (g *GitHub) ProjectSummary(ctx context.Context, p boardconfig.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, g.ready(), g)
}

// seedHeads loads the default branch and the first page of remote branches.
// This replaces the old 50-page census; open PR branches are added separately
// by loadOpenReviews.
func (g *GitHub) seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	const op = "remotegit.GitHub.seedHeads"
	var snap HeadsSnapshot
	raw, err := g.Run.RunJSON(ctx, "gh", "api", "repos/"+repo, "--jq", "{default: .default_branch}")
	if err != nil {
		return snap, forgeErr(err, op, "read github default branch failed", repo)
	}
	var meta struct {
		Default string `json:"default"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return snap, forgeErr(err, op, "parse github default branch failed", repo)
	}
	snap.DefaultBranch = meta.Default

	path := fmt.Sprintf("repos/%s/branches?per_page=%d&page=1", repo, listPageSize)
	raw, err = g.Run.RunJSON(ctx, "gh", "api", path)
	if err != nil {
		return snap, forgeErr(err, op, "list github branches failed", repo)
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
		return snap, forgeErr(err, op, "parse github branches failed", repo)
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
	const op = "remotegit.GitHub.loadCI"
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "list",
		"--repo", repo,
		"--limit", "20",
		"--json", "databaseId,status,conclusion,displayTitle,url,headBranch,updatedAt,workflowName",
	)
	if err != nil {
		return CISnapshot{}, forgeErr(err, op, "list github workflow runs failed", repo)
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
		return CISnapshot{}, forgeErr(err, op, "parse github workflow runs failed", repo)
	}
	rows := make([]ciRow, 0, len(runs))
	for _, r := range runs {
		rows = append(rows, ciRow{
			run: CachedCIRun{
				Branch: r.Branch, Status: firstNonEmpty(r.Conclusion, r.Status), URL: r.URL,
				UpdatedAt: r.UpdatedAt, RunID: strconv.FormatInt(r.ID, 10),
			},
			conclusion: r.Conclusion,
			name:       firstNonEmpty(r.Workflow, r.Title),
		})
	}
	return newCISnapshot(rows), nil
}

func (g *GitHub) loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error) {
	const op = "remotegit.GitHub.loadOpenReviews"
	prRaw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "open",
		"--json", "number,headRefName,url,mergeable,mergeStateStatus,updatedAt,isDraft",
	)
	if err != nil {
		return OpenReviewsSnapshot{}, forgeErr(err, op, "list github open pull requests failed", repo)
	}
	var rows []struct {
		Number           int    `json:"number"`
		Head             string `json:"headRefName"`
		URL              string `json:"url"`
		Mergeable        string `json:"mergeable"`
		MergeStateStatus string `json:"mergeStateStatus"`
		UpdatedAt        string `json:"updatedAt"`
		IsDraft          bool   `json:"isDraft"`
	}
	if err := json.Unmarshal(prRaw, &rows); err != nil {
		return OpenReviewsSnapshot{}, forgeErr(err, op, "parse github open pull requests failed", repo)
	}
	prs := make([]forgePR, 0, len(rows))
	for _, pr := range rows {
		prs = append(prs, forgePR{
			Branch:    pr.Head,
			ID:        pr.Number,
			URL:       pr.URL,
			Conflict:  githubHasConflict(pr.Mergeable, pr.MergeStateStatus),
			UpdatedAt: pr.UpdatedAt,
			Draft:     pr.IsDraft,
		})
	}
	return newOpenReviews(boardconfig.HostGitHub, prs), nil
}

func (g *GitHub) loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error) {
	const op = "remotegit.GitHub.loadMerged"
	raw, err := g.Run.RunJSON(ctx, "gh", "pr", "list",
		"--repo", repo,
		"--state", "merged",
		"--limit", "50",
		"--json", "number,headRefName,url,mergedAt",
	)
	if err != nil {
		return nil, forgeErr(err, op, "list github merged pull requests failed", repo)
	}
	prs, err := decodeGitHubMerged(raw)
	if err != nil {
		return nil, forgeErr(err, op, "parse github merged pull requests failed", repo)
	}
	return mergedReviews(prs), nil
}

// MergedForBranch looks up merged PRs whose head branch matches branch.
func (g *GitHub) MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error) {
	const op = "remotegit.GitHub.MergedForBranch"
	if err := g.ready(); err != nil {
		return nil, err
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
		return nil, forgeErr(err, op, "look up github merged pull request failed", repo).With("branch", branch)
	}
	prs, err := decodeGitHubMerged(raw)
	if err != nil {
		return nil, forgeErr(err, op, "parse github merged pull request failed", repo).With("branch", branch)
	}
	return mergedReviews(prs), nil
}

func decodeGitHubMerged(raw []byte) ([]forgePR, error) {
	var rows []struct {
		Number   int    `json:"number"`
		Head     string `json:"headRefName"`
		URL      string `json:"url"`
		MergedAt string `json:"mergedAt"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	out := make([]forgePR, 0, len(rows))
	for _, pr := range rows {
		out = append(out, forgePR{Branch: pr.Head, ID: pr.Number, URL: pr.URL, MergedAt: pr.MergedAt})
	}
	return out, nil
}

// githubHasConflict reads gh's mergeable and merge-state fields; either one
// reporting a conflict marks the PR.
func githubHasConflict(mergeable, mergeState string) bool {
	m := strings.ToUpper(strings.TrimSpace(mergeable))
	s := strings.ToUpper(strings.TrimSpace(mergeState))
	return m == "CONFLICTING" || s == "DIRTY" || s == "CONFLICTING"
}

func (g *GitHub) FailedJobs(ctx context.Context, p boardconfig.Project, runID string) ([]board.FailedJob, error) {
	const op = "remotegit.GitHub.FailedJobs"
	if err := g.ready(); err != nil {
		return nil, err
	}
	if err := requireArg(op, "run id", runID); err != nil {
		return nil, err
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "view", runID,
		"--repo", p.Path,
		"--json", "jobs",
	)
	if err != nil {
		return nil, forgeErr(err, op, "list github jobs failed", p.Path).With("run_id", runID)
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
		return nil, forgeErr(err, op, "parse github jobs failed", p.Path).With("run_id", runID)
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

func (g *GitHub) JobLog(ctx context.Context, p boardconfig.Project, runID, jobID string) (string, error) {
	const op = "remotegit.GitHub.JobLog"
	if err := g.ready(); err != nil {
		return "", err
	}
	runID = strings.TrimSpace(runID)
	jobID = strings.TrimPrefix(strings.TrimSpace(jobID), "github:")
	if err := requireArg(op, "run id", runID); err != nil {
		return "", err
	}
	args := []string{"run", "view", runID, "--repo", p.Path}
	if jobID != "" {
		args = append(args, "--job", jobID, "--log")
	} else {
		args = append(args, "--log-failed")
	}
	out, err := g.Run.Run(ctx, "gh", args...)
	if err != nil {
		return "", forgeErr(err, op, "read github job log failed", p.Path).With("run_id", runID)
	}
	return truncateLog(string(out)), nil
}

// ListOrgRepos lists non-archived repositories for a GitHub org or user.
func (g *GitHub) ListOrgRepos(ctx context.Context, org string) ([]RepoRef, error) {
	const op = "remotegit.GitHub.ListOrgRepos"
	if err := g.ready(); err != nil {
		return nil, err
	}
	org = strings.TrimSpace(org)
	if err := requireArg(op, "github org", org); err != nil {
		return nil, err
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "repo", "list", org,
		"--limit", strconv.Itoa(maxListRepos),
		"--no-archived",
		"--json", "nameWithOwner,name,isArchived",
	)
	if err != nil {
		return nil, forgeErr(err, op, "list github repositories failed", org)
	}
	var rows []struct {
		NameWithOwner string `json:"nameWithOwner"`
		Name          string `json:"name"`
		IsArchived    bool   `json:"isArchived"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, forgeErr(err, op, "parse github repositories failed", org)
	}
	out := make([]RepoRef, 0, len(rows))
	for _, r := range rows {
		if r.IsArchived {
			continue
		}
		if ref, ok := newRepoRef(boardconfig.HostGitHub, r.NameWithOwner, r.Name); ok {
			out = append(out, ref)
		}
	}
	return out, nil
}
