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

// GitLab uses the glab CLI.
type GitLab struct {
	Run cliexec.Exec
}

func NewGitLab(run cliexec.Exec) *GitLab {
	if run == nil {
		panic("remotegit.NewGitLab: Exec is required")
	}
	return &GitLab{Run: run}
}

// ready reports a nil receiver or a literal GitLab{} without Run as
// ErrClientMissing, so exported methods fail with an error instead of panicking.
func (g *GitLab) ready() error {
	if g == nil || g.Run == nil {
		return clientMissing("remotegit.GitLab", "gitlab")
	}
	return nil
}

func (g *GitLab) AuthStatus(ctx context.Context) (bool, bool, string) {
	if err := g.ready(); err != nil {
		return false, false, err.Error()
	}
	return cliAuthStatus(ctx, g.Run, "glab not on PATH (brew install glab)", "glab", "auth", "status")
}

func (g *GitLab) host() boardconfig.Host { return boardconfig.HostGitLab }

func (g *GitLab) unauthMsg(detail string) string {
	return "glab not authenticated: " + detail
}

func (g *GitLab) ProjectSummary(ctx context.Context, p boardconfig.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, g.ready(), g)
}

// seedHeads loads the default branch and the first page of remote branches.
// This replaces the old 50-page census; open MR branches are added separately
// by loadOpenReviews.
func (g *GitLab) seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	const op = "remotegit.GitLab.seedHeads"
	var snap HeadsSnapshot
	encoded := strings.ReplaceAll(repo, "/", "%2F")
	raw, err := g.Run.RunJSON(ctx, "glab", "api", "projects/"+encoded+"?simple=true")
	if err != nil {
		return snap, forgeErr(err, op, "read gitlab default branch failed", repo)
	}
	var meta struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return snap, forgeErr(err, op, "parse gitlab default branch failed", repo)
	}
	snap.DefaultBranch = meta.DefaultBranch

	path := fmt.Sprintf("projects/%s/repository/branches?per_page=%d&page=1", encoded, listPageSize)
	raw, err = g.Run.RunJSON(ctx, "glab", "api", path)
	if err != nil {
		return snap, forgeErr(err, op, "list gitlab branches failed", repo)
	}
	var heads []struct {
		Name   string `json:"name"`
		WebURL string `json:"web_url"`
		Commit struct {
			CommittedDate string `json:"committed_date"`
			AuthoredDate  string `json:"authored_date"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(raw, &heads); err != nil {
		return snap, forgeErr(err, op, "parse gitlab branches failed", repo)
	}
	for _, h := range heads {
		snap.Heads = append(snap.Heads, RemoteHead{
			Name:      h.Name,
			UpdatedAt: firstNonEmpty(h.Commit.CommittedDate, h.Commit.AuthoredDate),
			WebURL:    h.WebURL,
		})
	}
	return snap, nil
}

func (g *GitLab) loadCI(ctx context.Context, repo string) (CISnapshot, error) {
	const op = "remotegit.GitLab.loadCI"
	raw, err := g.Run.RunJSON(ctx, "glab", "ci", "list",
		"-R", repo,
		"-P", "20",
		"--output", "json",
	)
	if err != nil {
		return CISnapshot{}, forgeErr(err, op, "list gitlab pipelines failed", repo)
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
		return CISnapshot{}, forgeErr(err, op, "parse gitlab pipelines failed", repo)
	}
	rows := make([]ciRow, 0, len(pipelines))
	for _, pl := range pipelines {
		rows = append(rows, ciRow{
			run: CachedCIRun{
				Branch: pl.Ref, Status: pl.Status, URL: pl.WebURL,
				UpdatedAt: pl.UpdatedAt, RunID: strconv.Itoa(pl.ID),
			},
			name: "pipeline",
		})
	}
	return newCISnapshot(rows), nil
}

func (g *GitLab) loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error) {
	const op = "remotegit.GitLab.loadOpenReviews"
	mrRaw, err := g.Run.RunJSON(ctx, "glab", "mr", "list",
		"-R", repo,
		"--output", "json",
	)
	if err != nil {
		return OpenReviewsSnapshot{}, forgeErr(err, op, "list gitlab open merge requests failed", repo)
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
		return OpenReviewsSnapshot{}, forgeErr(err, op, "parse gitlab open merge requests failed", repo)
	}
	prs := make([]forgePR, 0, len(mrs))
	for _, mr := range mrs {
		prs = append(prs, forgePR{
			Branch:    mr.SourceBranch,
			ID:        mr.IID,
			URL:       mr.WebURL,
			Conflict:  gitlabHasConflict(mr.HasConflicts, mr.MergeStatus),
			UpdatedAt: mr.UpdatedAt,
			Draft:     mr.Draft || mr.WorkInProg,
		})
	}
	return newOpenReviews(boardconfig.HostGitLab, prs), nil
}

func (g *GitLab) loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error) {
	const op = "remotegit.GitLab.loadMerged"
	raw, err := g.Run.RunJSON(ctx, "glab", "mr", "list",
		"-R", repo,
		"--merged",
		"--per-page", "50",
		"--output", "json",
	)
	if err != nil {
		return nil, forgeErr(err, op, "list gitlab merged merge requests failed", repo)
	}
	prs, err := decodeGitLabMerged(raw)
	if err != nil {
		return nil, forgeErr(err, op, "parse gitlab merged merge requests failed", repo)
	}
	return mergedReviews(prs), nil
}

// MergedForBranch looks up merged MRs whose source branch matches branch.
func (g *GitLab) MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error) {
	const op = "remotegit.GitLab.MergedForBranch"
	if err := g.ready(); err != nil {
		return nil, err
	}
	branch = trimBranch(branch)
	if err := forgeBranchNameOK(branch); err != nil {
		return nil, err
	}
	raw, err := g.Run.RunJSON(ctx, "glab", "mr", "list",
		"-R", repo,
		"--merged",
		"--source-branch", branch,
		"--per-page", "1",
		"--output", "json",
	)
	if err != nil {
		return nil, forgeErr(err, op, "look up gitlab merged merge request failed", repo).With("branch", branch)
	}
	prs, err := decodeGitLabMerged(raw)
	if err != nil {
		return nil, forgeErr(err, op, "parse gitlab merged merge request failed", repo).With("branch", branch)
	}
	return mergedReviews(prs), nil
}

func decodeGitLabMerged(raw []byte) ([]forgePR, error) {
	var mrs []struct {
		IID          int    `json:"iid"`
		SourceBranch string `json:"source_branch"`
		WebURL       string `json:"web_url"`
		MergedAt     string `json:"merged_at"`
	}
	if err := json.Unmarshal(raw, &mrs); err != nil {
		return nil, err
	}
	out := make([]forgePR, 0, len(mrs))
	for _, mr := range mrs {
		out = append(out, forgePR{Branch: mr.SourceBranch, ID: mr.IID, URL: mr.WebURL, MergedAt: mr.MergedAt})
	}
	return out, nil
}

// gitlabHasConflict reads glab's conflict flag and merge status; either one
// reporting a conflict marks the MR.
func gitlabHasConflict(hasConflicts bool, mergeStatus string) bool {
	if hasConflicts {
		return true
	}
	s := strings.ToLower(strings.TrimSpace(mergeStatus))
	return s == "cannot_be_merged" || s == "cannot_be_merged_recheck"
}

func (g *GitLab) FailedJobs(ctx context.Context, p boardconfig.Project, pipelineID string) ([]board.FailedJob, error) {
	const op = "remotegit.GitLab.FailedJobs"
	if err := g.ready(); err != nil {
		return nil, err
	}
	if err := requireArg(op, "pipeline id", pipelineID); err != nil {
		return nil, err
	}
	raw, err := g.Run.RunJSON(ctx, "glab", "ci", "view", pipelineID,
		"-R", p.Path,
		"--output", "json",
	)
	if err != nil {
		return nil, forgeErr(err, op, "list gitlab jobs failed", p.Path).With("pipeline_id", pipelineID)
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
		return nil, forgeErr(err, op, "parse gitlab jobs failed", p.Path).With("pipeline_id", pipelineID)
	}
	var out []board.FailedJob
	for _, j := range payload.Jobs {
		if j.Status != "failed" {
			continue
		}
		out = append(out, board.FailedJob{
			ID:     fmt.Sprintf("gitlab:%d", j.ID),
			Name:   j.Name,
			Stage:  j.Stage,
			WebURL: j.WebURL,
		})
	}
	return out, nil
}

func (g *GitLab) JobLog(ctx context.Context, p boardconfig.Project, _, jobID string) (string, error) {
	const op = "remotegit.GitLab.JobLog"
	if err := g.ready(); err != nil {
		return "", err
	}
	jobID = strings.TrimPrefix(strings.TrimSpace(jobID), "gitlab:")
	if err := requireArg(op, "job id", jobID); err != nil {
		return "", err
	}
	out, err := g.Run.Run(ctx, "glab", "ci", "trace", jobID, "-R", p.Path)
	if err != nil {
		return "", forgeErr(err, op, "read gitlab job log failed", p.Path).With("job_id", jobID)
	}
	return truncateLog(string(out)), nil
}

// ListGroupRepos lists non-archived projects in a GitLab group (including subgroups).
func (g *GitLab) ListGroupRepos(ctx context.Context, group string) ([]RepoRef, error) {
	const op = "remotegit.GitLab.ListGroupRepos"
	if err := g.ready(); err != nil {
		return nil, err
	}
	group = strings.TrimSpace(group)
	if err := requireArg(op, "gitlab group", group); err != nil {
		return nil, err
	}
	var out []RepoRef
	err := pageLoop(ctx, func(page int) (bool, error) {
		refs, n, err := g.listGroupPage(ctx, op, group, page)
		if err != nil {
			return false, err
		}
		out = append(out, refs...)
		return n < listPageSize, nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// listGroupPage reads one page of a group listing and returns its kept refs
// and the raw row count, which tells the caller whether a next page exists.
func (g *GitLab) listGroupPage(ctx context.Context, op, group string, page int) ([]RepoRef, int, error) {
	raw, err := g.Run.RunJSON(ctx, "glab", "repo", "list",
		"--group", group,
		"--include-subgroups",
		"--archived=false",
		"--per-page", strconv.Itoa(listPageSize),
		"--page", strconv.Itoa(page),
		"--output", "json",
	)
	if err != nil {
		return nil, 0, forgeErr(err, op, "list gitlab repositories failed", group)
	}
	var rows []struct {
		PathWithNamespace string `json:"path_with_namespace"`
		Name              string `json:"name"`
		Archived          bool   `json:"archived"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, 0, forgeErr(err, op, "parse gitlab repositories failed", group)
	}
	refs := make([]RepoRef, 0, len(rows))
	for _, r := range rows {
		if r.Archived {
			continue
		}
		if ref, ok := newRepoRef(boardconfig.HostGitLab, r.PathWithNamespace, r.Name); ok {
			refs = append(refs, ref)
		}
	}
	return refs, len(rows), nil
}
