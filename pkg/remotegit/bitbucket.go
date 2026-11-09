package remotegit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// DefaultBitbucketBaseURL is the Bitbucket Cloud REST API root.
const DefaultBitbucketBaseURL = "https://api.bitbucket.org/2.0"

// bitbucketCredentialsHint tells the operator how to authenticate.
const bitbucketCredentialsHint = "set BITBUCKET_TOKEN or BITBUCKET_USERNAME and BITBUCKET_APP_PASSWORD"

// BitbucketCredentials authenticate Bitbucket API requests: an access token
// (sent as Bearer), or a username with an app password (sent as Basic).
// They are set on each request's headers only: never logged, never passed in
// argv and never put in an error. String and GoString redact them.
type BitbucketCredentials struct {
	Token       string
	Username    string
	AppPassword string
}

// BitbucketCredentialsFromEnv reads BITBUCKET_TOKEN, BITBUCKET_USERNAME and
// BITBUCKET_APP_PASSWORD.
func BitbucketCredentialsFromEnv() BitbucketCredentials {
	return BitbucketCredentials{
		Token:       strings.TrimSpace(os.Getenv("BITBUCKET_TOKEN")),
		Username:    strings.TrimSpace(os.Getenv("BITBUCKET_USERNAME")),
		AppPassword: strings.TrimSpace(os.Getenv("BITBUCKET_APP_PASSWORD")),
	}
}

// String redacts the credentials.
func (c BitbucketCredentials) String() string { return "BitbucketCredentials{redacted}" }

// GoString redacts the credentials for %#v.
func (c BitbucketCredentials) GoString() string { return c.String() }

func (c BitbucketCredentials) usable() bool {
	return c.Token != "" || (c.Username != "" && c.AppPassword != "")
}

// apply sets the Authorization header; a token wins over username and password.
func (c BitbucketCredentials) apply(req *http.Request) {
	switch {
	case c.Token != "":
		req.Header.Set("Authorization", "Bearer "+c.Token)
	case c.Username != "" && c.AppPassword != "":
		req.SetBasicAuth(c.Username, c.AppPassword)
	}
}

// BitbucketConfig holds what a Bitbucket client needs at construction time.
type BitbucketConfig struct {
	// HTTP sends API requests. Required. Leave Timeout zero: the caller's
	// context deadline bounds each request.
	HTTP *http.Client
	// BaseURL is the API root. Optional: empty uses DefaultBitbucketBaseURL.
	// Tests point it at an httptest server.
	BaseURL string
	// Credentials returns the credentials for each request. Required.
	// BitbucketCredentialsFromEnv reads the environment on every call.
	Credentials func() BitbucketCredentials
}

// DefaultBitbucketConfig talks to Bitbucket Cloud with environment credentials.
func DefaultBitbucketConfig() BitbucketConfig {
	return BitbucketConfig{
		HTTP:        &http.Client{},
		BaseURL:     DefaultBitbucketBaseURL,
		Credentials: BitbucketCredentialsFromEnv,
	}
}

// CreateBitbucket builds the client. Panics if HTTP or Credentials is nil
// (required dependencies at wire-up).
func (cfg BitbucketConfig) CreateBitbucket() *Bitbucket {
	if cfg.HTTP == nil || cfg.Credentials == nil {
		panic("remotegit.CreateBitbucket: HTTP and Credentials are required")
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = DefaultBitbucketBaseURL
	}
	return &Bitbucket{
		api:     newForgeHTTP("bitbucket", cfg.HTTP),
		baseURL: base,
		creds:   cfg.Credentials,
	}
}

// Bitbucket talks to the Bitbucket Cloud REST API over net/http.
// Project Path format is workspace/repo. Build it with
// BitbucketConfig.CreateBitbucket.
type Bitbucket struct {
	api     *forgeHTTP
	baseURL string
	creds   func() BitbucketCredentials
}

// ready reports a nil receiver or a literal Bitbucket{} as ErrClientMissing,
// so exported methods fail with an error instead of panicking.
func (b *Bitbucket) ready() error {
	if b == nil || b.api == nil || b.creds == nil {
		return clientMissing("remotegit.Bitbucket", "bitbucket")
	}
	return nil
}

// AuthStatus reports whether credentials are set. It makes no request: the
// HTTP client is built in, so the forge always counts as installed.
func (b *Bitbucket) AuthStatus(_ context.Context) (bool, bool, string) {
	if err := b.ready(); err != nil {
		return false, false, err.Error()
	}
	if !b.creds().usable() {
		return true, false, bitbucketCredentialsHint
	}
	return true, true, ""
}

func (b *Bitbucket) host() boardconfig.Host { return boardconfig.HostBitbucket }

func (b *Bitbucket) unauthMsg(detail string) string {
	return "bitbucket not authenticated: " + detail
}

func (b *Bitbucket) ProjectSummary(ctx context.Context, p boardconfig.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, b.ready(), b)
}

// apiGET fetches apiPath (relative to the base URL, query included) with the
// current credentials.
func (b *Bitbucket) apiGET(ctx context.Context, apiPath string) ([]byte, error) {
	const op = "remotegit.Bitbucket.apiGET"
	creds := b.creds()
	if !creds.usable() {
		return nil, gberr.New(gberr.CodeUnavailable, op, "bitbucket credentials are not set")
	}
	apiPath = strings.TrimSpace(apiPath)
	if !strings.HasPrefix(apiPath, "/") {
		apiPath = "/" + apiPath
	}
	body, err := b.api.get(ctx, b.baseURL+apiPath, creds.apply)
	if err != nil {
		return nil, httpErr(err, op, "bitbucket")
	}
	return body, nil
}

// repoAPIPath is the API path of a repository resource; suffix starts with
// "/" or "?" or is empty.
func repoAPIPath(workspace, name, suffix string) string {
	return "/repositories/" + url.PathEscape(workspace) + "/" + url.PathEscape(name) + suffix
}

// splitBitbucketPath splits workspace/repo; with more than two segments the
// last one names the repository.
func splitBitbucketPath(op, repo string) (workspace, name string, err error) {
	parts, err := splitRepo(boardconfig.HostBitbucket, op, repo)
	if err != nil {
		return "", "", err
	}
	return parts[0], parts[len(parts)-1], nil
}

func (b *Bitbucket) seedHeads(ctx context.Context, repo string) (HeadsSnapshot, error) {
	const op = "remotegit.Bitbucket.seedHeads"
	var snap HeadsSnapshot
	workspace, name, err := splitBitbucketPath(op, repo)
	if err != nil {
		return snap, err
	}
	raw, err := b.apiGET(ctx, repoAPIPath(workspace, name, ""))
	if err != nil {
		return snap, forgeErr(err, op, "read bitbucket default branch failed", repo)
	}
	var meta struct {
		MainBranch struct {
			Name string `json:"name"`
		} `json:"mainbranch"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return snap, forgeErr(err, op, "parse bitbucket repository failed", repo)
	}
	snap.DefaultBranch = meta.MainBranch.Name

	raw, err = b.apiGET(ctx, repoAPIPath(workspace, name, "/refs/branches?pagelen="+strconv.Itoa(listPageSize)))
	if err != nil {
		return snap, forgeErr(err, op, "list bitbucket branches failed", repo)
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
		return snap, forgeErr(err, op, "parse bitbucket branches failed", repo)
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
	const op = "remotegit.Bitbucket.loadCI"
	workspace, name, err := splitBitbucketPath(op, repo)
	if err != nil {
		return CISnapshot{}, err
	}
	raw, err := b.apiGET(ctx, repoAPIPath(workspace, name, "/pipelines/?pagelen=20&sort=-created_on"))
	if err != nil {
		return CISnapshot{}, forgeErr(err, op, "list bitbucket pipelines failed", repo)
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
		return CISnapshot{}, forgeErr(err, op, "parse bitbucket pipelines failed", repo)
	}
	rows := make([]ciRow, 0, len(page.Values))
	for _, r := range page.Values {
		rows = append(rows, ciRow{
			run: CachedCIRun{
				Branch:    r.Target.RefName,
				Status:    firstNonEmpty(r.State.Result.Name, r.State.Name),
				URL:       r.Links.HTML.Href,
				UpdatedAt: firstNonEmpty(r.CompletedOn, r.CreatedOn),
				RunID:     firstNonEmpty(strings.Trim(r.UUID, "{}"), strconv.Itoa(r.BuildNumber)),
			},
			conclusion: r.State.Result.Name,
			name:       "pipeline",
		})
	}
	return newCISnapshot(rows), nil
}

func (b *Bitbucket) loadOpenReviews(ctx context.Context, repo string) (OpenReviewsSnapshot, error) {
	const op = "remotegit.Bitbucket.loadOpenReviews"
	prs, err := b.listPRs(ctx, op, repo, "OPEN", "open")
	if err != nil {
		return OpenReviewsSnapshot{}, err
	}
	return newOpenReviews(boardconfig.HostBitbucket, prs), nil
}

func (b *Bitbucket) loadMerged(ctx context.Context, repo string) ([]board.MergedReview, error) {
	const op = "remotegit.Bitbucket.loadMerged"
	prs, err := b.listPRs(ctx, op, repo, "MERGED", "merged")
	if err != nil {
		return nil, err
	}
	return mergedReviews(prs), nil
}

// listPRs reads the first page of pull requests in state (OPEN or MERGED);
// noun names the state in client-safe error messages.
func (b *Bitbucket) listPRs(ctx context.Context, op, repo, state, noun string) ([]forgePR, error) {
	workspace, name, err := splitBitbucketPath(op, repo)
	if err != nil {
		return nil, err
	}
	raw, err := b.apiGET(ctx, repoAPIPath(workspace, name, "/pullrequests?state="+state+"&pagelen=50"))
	if err != nil {
		return nil, forgeErr(err, op, "list bitbucket "+noun+" pull requests failed", repo)
	}
	prs, err := decodeBitbucketPRs(raw)
	if err != nil {
		return nil, forgeErr(err, op, "parse bitbucket "+noun+" pull requests failed", repo)
	}
	return prs, nil
}

// MergedForBranch looks up merged PRs whose source branch matches branch.
func (b *Bitbucket) MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error) {
	const op = "remotegit.Bitbucket.MergedForBranch"
	if err := b.ready(); err != nil {
		return nil, err
	}
	branch = trimBranch(branch)
	if err := forgeBranchNameOK(branch); err != nil {
		return nil, err
	}
	merged, err := b.loadMerged(ctx, repo)
	if err != nil {
		return nil, gberr.Wrap(err, gberr.CodeOf(err), op, "look up bitbucket merged pull request failed").With("branch", branch)
	}
	var out []board.MergedReview
	for _, m := range merged {
		if trimBranch(m.Branch) == branch {
			out = append(out, m)
		}
	}
	return out, nil
}

func (b *Bitbucket) FailedJobs(_ context.Context, _ boardconfig.Project, runID string) ([]board.FailedJob, error) {
	const op = "remotegit.Bitbucket.FailedJobs"
	if err := b.ready(); err != nil {
		return nil, err
	}
	if err := requireArg(op, "run id", runID); err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *Bitbucket) JobLog(_ context.Context, _ boardconfig.Project, runID, _ string) (string, error) {
	const op = "remotegit.Bitbucket.JobLog"
	if err := b.ready(); err != nil {
		return "", err
	}
	if err := requireArg(op, "run id", runID); err != nil {
		return "", err
	}
	return "", gberr.New(gberr.CodeInvalid, op, "bitbucket job logs are not supported yet")
}

// ListWorkspaceRepos lists repositories in a Bitbucket workspace, following
// the API's next links up to maxListPages.
func (b *Bitbucket) ListWorkspaceRepos(ctx context.Context, workspace string) ([]RepoRef, error) {
	const op = "remotegit.Bitbucket.ListWorkspaceRepos"
	if err := b.ready(); err != nil {
		return nil, err
	}
	workspace = strings.TrimSpace(workspace)
	if err := requireArg(op, "bitbucket workspace", workspace); err != nil {
		return nil, err
	}
	var out []RepoRef
	path := fmt.Sprintf("/repositories/%s?pagelen=%d", url.PathEscape(workspace), listPageSize)
	err := pageLoop(ctx, func(int) (bool, error) {
		refs, next, err := b.listWorkspacePage(ctx, op, workspace, path)
		if err != nil {
			return false, err
		}
		out = append(out, refs...)
		path = next
		return next == "", nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// listWorkspacePage reads one page of a workspace listing and returns its
// refs and the API path of the next page ("" on the last page).
func (b *Bitbucket) listWorkspacePage(ctx context.Context, op, workspace, path string) ([]RepoRef, string, error) {
	raw, err := b.apiGET(ctx, path)
	if err != nil {
		return nil, "", forgeErr(err, op, "list bitbucket repositories failed", workspace)
	}
	var payload struct {
		Next   string `json:"next"`
		Values []struct {
			Slug     string `json:"slug"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
		} `json:"values"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, "", forgeErr(err, op, "parse bitbucket repositories failed", workspace)
	}
	refs := make([]RepoRef, 0, len(payload.Values))
	for _, r := range payload.Values {
		full := firstNonEmpty(r.FullName, workspace+"/"+firstNonEmpty(r.Slug, r.Name))
		if ref, ok := newRepoRef(boardconfig.HostBitbucket, full, firstNonEmpty(r.Name, r.Slug)); ok {
			refs = append(refs, ref)
		}
	}
	if payload.Next == "" {
		return refs, "", nil
	}
	next, err := b.nextAPIPath(payload.Next)
	if err != nil {
		return nil, "", forgeErr(err, op, "bitbucket returned an unusable next page link", workspace)
	}
	return refs, next, nil
}

// nextAPIPath turns an absolute "next" link into a path under the base URL.
// A link to any other scheme, host or path prefix is refused, so the
// credentials are never sent off the API host.
func (b *Bitbucket) nextAPIPath(next string) (string, error) {
	base, err := url.Parse(b.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base url: %w", err)
	}
	u, err := url.Parse(next)
	if err != nil {
		return "", fmt.Errorf("parse next link: %w", err)
	}
	basePath := strings.TrimRight(base.EscapedPath(), "/")
	if u.Scheme != base.Scheme || u.Host != base.Host || !strings.HasPrefix(u.EscapedPath(), basePath+"/") {
		return "", fmt.Errorf("next link leaves the api root")
	}
	path := strings.TrimPrefix(u.EscapedPath(), basePath)
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	return path, nil
}

func decodeBitbucketPRs(raw []byte) ([]forgePR, error) {
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
	out := make([]forgePR, 0, len(page.Values))
	for _, r := range page.Values {
		updated := firstNonEmpty(r.UpdatedOn, r.CreatedOn)
		out = append(out, forgePR{
			Branch:    r.Source.Branch.Name,
			ID:        r.ID,
			URL:       r.Links.HTML.Href,
			UpdatedAt: updated,
			// Bitbucket has no merged_at; the last update of a merged PR is its merge.
			MergedAt: updated,
			Draft:    r.Draft,
		})
	}
	return out, nil
}
