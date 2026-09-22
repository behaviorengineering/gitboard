package remotegit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
)

// Bitbucket talks to api.bitbucket.org via curl and env credentials.
// Project Path format is workspace/repo.
type Bitbucket struct {
	Run cliexec.Exec
}

// NewBitbucket returns a Bitbucket forge client.
func NewBitbucket(run cliexec.Exec) *Bitbucket {
	if run == nil {
		panic("remotegit.NewBitbucket: Exec is required")
	}
	return &Bitbucket{Run: run}
}

func (b *Bitbucket) AuthStatus(ctx context.Context) (bool, bool, string) {
	_, curlErr := b.Run.LookPath("curl")
	token := strings.TrimSpace(os.Getenv("BITBUCKET_TOKEN"))
	user := strings.TrimSpace(os.Getenv("BITBUCKET_USERNAME"))
	pass := strings.TrimSpace(os.Getenv("BITBUCKET_APP_PASSWORD"))
	hasCreds := token != "" || (user != "" && pass != "")
	installed := curlErr == nil || hasCreds
	if !installed {
		return false, false, "curl not on PATH and no Bitbucket credentials"
	}
	if !hasCreds {
		return true, false, "set BITBUCKET_TOKEN or BITBUCKET_USERNAME+BITBUCKET_APP_PASSWORD"
	}
	return true, true, ""
}

func (b *Bitbucket) authCacheID() string { return "bitbucket" }

func (b *Bitbucket) unauthMsg(detail string) string {
	return "bitbucket not authenticated: " + detail
}

func (b *Bitbucket) ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, b)
}

func (b *Bitbucket) seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	var snap HeadsSnapshot
	workspace, name, err := splitBitbucketPath(repo)
	if err != nil {
		return snap, err
	}
	raw, err := b.apiGET(ctx, fmt.Sprintf("/repositories/%s/%s", url.PathEscape(workspace), url.PathEscape(name)))
	if err != nil {
		return snap, fmt.Errorf("bitbucket default branch: %w", err)
	}
	var meta struct {
		MainBranch struct {
			Name string `json:"name"`
		} `json:"mainbranch"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return snap, fmt.Errorf("parse bitbucket repo: %w", err)
	}
	snap.DefaultBranch = meta.MainBranch.Name

	raw, err = b.apiGET(ctx, fmt.Sprintf("/repositories/%s/%s/refs/branches?pagelen=100",
		url.PathEscape(workspace), url.PathEscape(name)))
	if err != nil {
		return snap, fmt.Errorf("bitbucket branches: %w", err)
	}
	var page struct {
		Values []struct {
			Name   string `json:"name"`
			Target struct {
				Date string `json:"date"`
			} `json:"target"`
		} `json:"values"`
	}
	if err := json.Unmarshal(raw, &page); err != nil {
		return snap, fmt.Errorf("parse bitbucket branches: %w", err)
	}
	for _, h := range page.Values {
		snap.Heads = append(snap.Heads, RemoteHead{
			Name:      h.Name,
			UpdatedAt: h.Target.Date,
		})
	}
	return snap, nil
}

func (b *Bitbucket) loadCI(ctx context.Context, repo string) (CISnapshot, error) {
	var snap CISnapshot
	workspace, name, err := splitBitbucketPath(repo)
	if err != nil {
		return snap, nil
	}
	raw, err := b.apiGET(ctx, fmt.Sprintf("/repositories/%s/%s/pipelines/?pagelen=20&sort=-created_on",
		url.PathEscape(workspace), url.PathEscape(name)))
	if err != nil {
		return snap, nil
	}
	var page struct {
		Values []struct {
			UUID  string `json:"uuid"`
			State struct {
				Name   string `json:"name"`
				Result struct {
					Name string `json:"name"`
				} `json:"result"`
			} `json:"state"`
			Target struct {
				RefName string `json:"ref_name"`
			} `json:"target"`
			CreatedOn   string `json:"created_on"`
			CompletedOn string `json:"completed_on"`
			BuildNumber int    `json:"build_number"`
			Links       struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"links"`
		} `json:"values"`
	}
	if err := json.Unmarshal(raw, &page); err != nil {
		return snap, nil
	}
	for i, r := range page.Values {
		status := firstNonEmpty(r.State.Result.Name, r.State.Name)
		runID := firstNonEmpty(strings.Trim(r.UUID, "{}"), fmt.Sprintf("%d", r.BuildNumber))
		updated := firstNonEmpty(r.CompletedOn, r.CreatedOn)
		snap.Runs = append(snap.Runs, CachedCIRun{
			Branch: r.Target.RefName, Status: status, URL: r.Links.HTML.Href,
			UpdatedAt: updated, RunID: runID,
		})
		if i == 0 {
			snap.Latest = &board.CIStatus{
				Status:     status,
				Conclusion: r.State.Result.Name,
				Ref:        r.Target.RefName,
				Name:       "pipeline",
				WebURL:     r.Links.HTML.Href,
				UpdatedAt:  updated,
				RunID:      runID,
			}
		}
	}
	return snap, nil
}

func (b *Bitbucket) loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error) {
	var snap OpenReviewsSnapshot
	workspace, name, err := splitBitbucketPath(repo)
	if err != nil {
		return snap, nil
	}
	raw, err := b.apiGET(ctx, fmt.Sprintf("/repositories/%s/%s/pullrequests?state=OPEN&pagelen=50",
		url.PathEscape(workspace), url.PathEscape(name)))
	if err != nil {
		return snap, nil
	}
	prs, err := decodeBitbucketPRs(raw)
	if err != nil {
		return snap, nil
	}
	snap.PullRequests = len(prs)
	for _, pr := range prs {
		snap.Reviews = append(snap.Reviews, CachedOpenReview{
			Branch:    pr.Branch,
			ID:        pr.ID,
			URL:       pr.URL,
			Conflict:  false,
			UpdatedAt: pr.UpdatedAt,
			Draft:     pr.Draft,
		})
	}
	return snap, nil
}

func (b *Bitbucket) loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error) {
	workspace, name, err := splitBitbucketPath(repo)
	if err != nil {
		return nil, nil
	}
	raw, err := b.apiGET(ctx, fmt.Sprintf("/repositories/%s/%s/pullrequests?state=MERGED&pagelen=50",
		url.PathEscape(workspace), url.PathEscape(name)))
	if err != nil {
		return nil, nil
	}
	prs, err := decodeBitbucketPRs(raw)
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
			MergedAt: firstNonEmpty(pr.MergedAt, pr.UpdatedAt),
		})
	}
	return out, nil
}

// MergedForBranch looks up merged PRs whose source branch matches branch.
func (b *Bitbucket) MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error) {
	if b == nil {
		return nil, fmt.Errorf("bitbucket client missing")
	}
	branch = trimBranch(branch)
	if err := forgeBranchNameOK(branch); err != nil {
		return nil, err
	}
	merged, err := b.loadMerged(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("bitbucket merged for %s: %w", branch, err)
	}
	var out []board.MergedReview
	for _, m := range merged {
		if trimBranch(m.Branch) == branch {
			out = append(out, m)
		}
	}
	return out, nil
}

func (b *Bitbucket) FailedJobs(ctx context.Context, p config.Project, runID string) ([]board.FailedJob, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("missing run id")
	}
	return nil, nil
}

func (b *Bitbucket) JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error) {
	if strings.TrimSpace(runID) == "" {
		return "", fmt.Errorf("missing run id")
	}
	return "", fmt.Errorf("bitbucket job logs not supported yet")
}

// ListWorkspaceRepos lists non-archived repositories in a Bitbucket workspace.
func (b *Bitbucket) ListWorkspaceRepos(ctx context.Context, workspace string) ([]RepoRef, error) {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return nil, fmt.Errorf("missing bitbucket workspace")
	}
	var out []RepoRef
	path := fmt.Sprintf("/repositories/%s?pagelen=100", url.PathEscape(workspace))
	for page := 0; page < 50; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		raw, err := b.apiGET(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("bitbucket list repos: %w", err)
		}
		var payload struct {
			Next   string `json:"next"`
			Values []struct {
				Slug      string `json:"slug"`
				Name      string `json:"name"`
				FullName  string `json:"full_name"`
				IsPrivate bool   `json:"is_private"`
			} `json:"values"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("parse bitbucket repos: %w", err)
		}
		for _, r := range payload.Values {
			pathName := strings.Trim(r.FullName, "/")
			if pathName == "" {
				pathName = workspace + "/" + firstNonEmpty(r.Slug, r.Name)
			}
			name := firstNonEmpty(r.Name, r.Slug)
			parts := strings.Split(pathName, "/")
			if name == "" && len(parts) > 0 {
				name = parts[len(parts)-1]
			}
			out = append(out, RepoRef{
				Host: config.HostBitbucket,
				Path: pathName,
				Name: name,
			})
		}
		if payload.Next == "" {
			break
		}
		u, err := url.Parse(payload.Next)
		if err != nil {
			break
		}
		path = strings.TrimPrefix(u.Path, "/2.0") + "?" + u.RawQuery
	}
	return out, nil
}

func (b *Bitbucket) apiGET(ctx context.Context, apiPath string) ([]byte, error) {
	if _, err := b.Run.LookPath("curl"); err != nil {
		return nil, fmt.Errorf("curl not on PATH: %w", err)
	}
	apiPath = strings.TrimSpace(apiPath)
	if !strings.HasPrefix(apiPath, "/") {
		apiPath = "/" + apiPath
	}
	endpoint := "https://api.bitbucket.org/2.0" + apiPath
	args := []string{"-sS"}
	token := strings.TrimSpace(os.Getenv("BITBUCKET_TOKEN"))
	user := strings.TrimSpace(os.Getenv("BITBUCKET_USERNAME"))
	pass := strings.TrimSpace(os.Getenv("BITBUCKET_APP_PASSWORD"))
	switch {
	case token != "":
		args = append(args, "-H", "Authorization: Bearer "+token)
	case user != "" && pass != "":
		args = append(args, "-u", user+":"+pass)
	default:
		return nil, fmt.Errorf("missing Bitbucket credentials")
	}
	args = append(args, endpoint)
	return b.Run.RunJSON(ctx, "curl", args...)
}

type bitbucketPR struct {
	Branch    string
	ID        int
	URL       string
	UpdatedAt string
	MergedAt  string
	Draft     bool
}

func decodeBitbucketPRs(raw []byte) ([]bitbucketPR, error) {
	var page struct {
		Values []struct {
			ID        int    `json:"id"`
			UpdatedOn string `json:"updated_on"`
			CreatedOn string `json:"created_on"`
			Draft     bool   `json:"draft"`
			Source    struct {
				Branch struct {
					Name string `json:"name"`
				} `json:"branch"`
			} `json:"source"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"links"`
		} `json:"values"`
	}
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}
	out := make([]bitbucketPR, 0, len(page.Values))
	for _, r := range page.Values {
		out = append(out, bitbucketPR{
			Branch:    r.Source.Branch.Name,
			ID:        r.ID,
			URL:       r.Links.HTML.Href,
			UpdatedAt: firstNonEmpty(r.UpdatedOn, r.CreatedOn),
			MergedAt:  r.UpdatedOn,
			Draft:     r.Draft,
		})
	}
	return out, nil
}

func splitBitbucketPath(repo string) (workspace, name string, err error) {
	parts := strings.Split(strings.Trim(repo, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("bitbucket path must be workspace/repo")
	}
	for _, p := range parts {
		if p == "" {
			return "", "", fmt.Errorf("bitbucket path must be workspace/repo")
		}
	}
	return parts[0], parts[len(parts)-1], nil
}
