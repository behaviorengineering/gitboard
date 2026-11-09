package remotegit

import (
	"context"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// Discovery limits shared by the org, group and workspace listers.
const (
	// maxListPages is how many pages a paged lister reads before it stops.
	maxListPages = 50
	// listPageSize is the rows per page asked of paged forge APIs.
	listPageSize = 100
	// maxListRepos is the row limit for a lister that asks in one call.
	maxListRepos = 1000
)

// ciRow is one forge CI run in the shared shape.
type ciRow struct {
	run        CachedCIRun
	conclusion string
	name       string
}

// newCISnapshot builds the cached CI slice from runs ordered newest first:
// every run attaches to its branch, and the first run is the project's Latest.
func newCISnapshot(rows []ciRow) CISnapshot {
	var snap CISnapshot
	for i, r := range rows {
		snap.Runs = append(snap.Runs, r.run)
		if i == 0 {
			snap.Latest = &board.CIStatus{
				Status:     r.run.Status,
				Conclusion: r.conclusion,
				Ref:        r.run.Branch,
				Name:       r.name,
				WebURL:     r.run.URL,
				UpdatedAt:  r.run.UpdatedAt,
				RunID:      r.run.RunID,
			}
		}
	}
	return snap
}

// forgePR is one pull or merge request decoded from any forge.
type forgePR struct {
	Branch    string
	URL       string
	UpdatedAt string
	MergedAt  string
	ID        int
	Conflict  bool
	Draft     bool
}

// mergedReviews maps merged reviews to board rows and drops any with no
// source branch. The result feeds prune safety, so every forge goes through
// this one filter.
func mergedReviews(prs []forgePR) []board.MergedReview {
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
	return out
}

// newOpenReviews counts open reviews under the forge's noun (merge requests
// on GitLab, pull requests elsewhere) and keeps one cached row per review.
func newOpenReviews(host boardconfig.Host, prs []forgePR) OpenReviewsSnapshot {
	var snap OpenReviewsSnapshot
	if host.UsesMergeRequests() {
		snap.MergeRequests = len(prs)
	} else {
		snap.PullRequests = len(prs)
	}
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
	return snap
}

// cliAuthStatus probes a forge CLI login. A missing binary reports notFound;
// a failing status command reports its error without the argv prefix.
func cliAuthStatus(ctx context.Context, run cliexec.Exec, notFound, bin string, args ...string) (installed, authed bool, detail string) {
	if _, err := run.LookPath(bin); err != nil {
		return false, false, notFound
	}
	if _, err := run.Run(ctx, bin, args...); err != nil {
		prefix := bin + " " + strings.Join(args, " ") + ": "
		return true, false, strings.TrimPrefix(err.Error(), prefix)
	}
	return true, true, ""
}

// newRepoRef builds a discovered repository. The path must have the shape
// boardconfig.SplitRepoPath accepts for host; a blank name falls back to the
// last path segment.
func newRepoRef(host boardconfig.Host, path, name string) (RepoRef, bool) {
	path = strings.Trim(strings.TrimSpace(path), "/")
	parts, err := boardconfig.SplitRepoPath(host, path)
	if err != nil {
		return RepoRef{}, false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = parts[len(parts)-1]
	}
	return RepoRef{Host: host, Path: path, Name: name}, true
}

// pageLoop calls fetch for pages 1 to maxListPages until fetch reports done.
// It checks ctx before every page so a cancelled discovery stops promptly.
func pageLoop(ctx context.Context, fetch func(page int) (done bool, err error)) error {
	for page := 1; page <= maxListPages; page++ {
		if err := ctx.Err(); err != nil {
			return gberr.Wrap(err, gberr.CodeCanceled, "remotegit.pageLoop", "repository listing cancelled")
		}
		done, err := fetch(page)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
	return nil
}

// splitRepo checks repo against the host's path rule, wrapping a bad shape
// as repoPathErr for op.
func splitRepo(host boardconfig.Host, op, repo string) ([]string, error) {
	parts, err := boardconfig.SplitRepoPath(host, repo)
	if err != nil {
		return nil, repoPathErr(err, op, repo)
	}
	return parts, nil
}

// requireArg returns a CodeInvalid error when value is blank.
func requireArg(op, what, value string) error {
	if strings.TrimSpace(value) == "" {
		return gberr.New(gberr.CodeInvalid, op, what+" is required")
	}
	return nil
}
