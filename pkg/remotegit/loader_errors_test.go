package remotegit

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/remotegit/remotegittest"
)

var errForgeDown = errors.New("forge down: stderr detail")

type summarizer interface {
	ProjectSummary(ctx context.Context, p boardconfig.Project, opts SummaryOpts) (board.ProjectSummary, error)
}

// forgeOpenFailureCase is a forge whose open-review or merged call fails
// after fail(slice) and recovers after heal(slice). The open-review slice
// returns one open review when healthy.
type forgeOpenFailureCase struct {
	name    string
	client  summarizer
	loader  summaryLoader
	project boardconfig.Project
	fail    func(slice string)
	heal    func(slice string)
	// isCause reports whether err carries the injected failure.
	isCause func(error) bool
}

func openFailureCases(t *testing.T) []forgeOpenFailureCase {
	t.Helper()
	azFake := &remotegittest.FakeExec{
		Responses: map[string][]byte{
			"account show":        []byte(`{"id":"sub"}`),
			"repos show":          []byte(`{"defaultBranch":"refs/heads/main"}`),
			"ref list":            []byte(`[{"name":"refs/heads/main"},{"name":"refs/heads/feat/z"}]`),
			"pipelines runs list": []byte(`[]`),
			"--status active":     []byte(`[{"pullRequestId":5,"sourceRefName":"refs/heads/feat/z","mergeStatus":"succeeded"}]`),
			"--status completed":  []byte(`[]`),
		},
		Errors: map[string]error{},
	}
	azKeys := map[string]string{"open": "--status active", "merged": "--status completed"}
	bbFake := remotegittest.NewFakeHTTP(map[string][]byte{
		"repositories/acme/app/refs/branches":             []byte(`{"values":[{"name":"main"},{"name":"feat/b"}]}`),
		"repositories/acme/app/pullrequests?state=OPEN":   []byte(`{"values":[{"id":8,"source":{"branch":{"name":"feat/b"}}}]}`),
		"repositories/acme/app/pullrequests?state=MERGED": []byte(`{"values":[]}`),
		"repositories/acme/app/pipelines/":                []byte(`{"values":[]}`),
		"repositories/acme/app":                           []byte(`{"mainbranch":{"name":"main"}}`),
	})
	bbKeys := map[string]string{
		"open":   "repositories/acme/app/pullrequests?state=OPEN",
		"merged": "repositories/acme/app/pullrequests?state=MERGED",
	}
	az := NewAzureDevOps(azFake)
	bb := newTestBitbucket(t, bbFake, BitbucketCredentials{Token: "test-token"})
	return []forgeOpenFailureCase{
		{
			name: "azuredevops", client: az, loader: az,
			project: boardconfig.Project{ID: "az-app", Host: boardconfig.HostAzureDevOps, Path: "acme/proj/app"},
			fail:    func(slice string) { azFake.Errors[azKeys[slice]] = errForgeDown },
			heal:    func(slice string) { delete(azFake.Errors, azKeys[slice]) },
			isCause: func(err error) bool { return errors.Is(err, errForgeDown) },
		},
		{
			name: "bitbucket", client: bb, loader: bb,
			project: boardconfig.Project{ID: "bb-app", Host: boardconfig.HostBitbucket, Path: "acme/app"},
			// 403 is permanent, so the failure costs one request and no backoff.
			fail: func(slice string) { bbFake.SetStatus(bbKeys[slice], http.StatusForbidden) },
			heal: func(slice string) { bbFake.ClearStatus(bbKeys[slice]) },
			isCause: func(err error) bool {
				var se *httpStatusError
				return errors.As(err, &se) && se.status == http.StatusForbidden
			},
		},
	}
}

// A failed open-review load must surface as an error and must not be cached
// as "zero open reviews": the open-review set is the last guard before a
// prune, and a cached empty success would pass it for the whole TTL.
func TestForgeOpenReviewFailureIsNotCachedAsEmptySuccess(t *testing.T) {
	for _, tc := range openFailureCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			tc.fail("open")
			assertOpenLoadFails(t, tc)
			opts := SummaryOpts{Now: time.Now, Cache: NewTTLCache(time.Now), OpenTTL: time.Hour}
			sum := summarize(t, tc, opts)
			if !strings.Contains(sum.Error, "open pull requests failed") || sum.OpenItems.PullRequests != 0 {
				t.Fatalf("failed load must be reported, with no open reviews: error=%q items=%+v", sum.Error, sum.OpenItems)
			}

			// The forge recovers inside the open TTL. A cached empty
			// success would hide the open review until the TTL expired.
			tc.heal("open")
			sum = summarize(t, tc, opts)
			if sum.OpenItems.PullRequests != 1 || sum.Error != "" {
				t.Fatalf("after recovery: open=%d error=%q (failure was cached)", sum.OpenItems.PullRequests, sum.Error)
			}
		})
	}
}

func assertOpenLoadFails(t *testing.T, tc forgeOpenFailureCase) {
	t.Helper()
	_, err := tc.loader.loadOpenReviews(deadlineCtx(t), tc.project.Path)
	if !tc.isCause(err) {
		t.Fatalf("loadOpenReviews err = %v, want the forge cause", err)
	}
	if got := gberr.CodeOf(err); got != gberr.CodeUnavailable {
		t.Fatalf("loadOpenReviews code = %q, want %q", got, gberr.CodeUnavailable)
	}
}

func summarize(t *testing.T, tc forgeOpenFailureCase, opts SummaryOpts) board.ProjectSummary {
	t.Helper()
	sum, err := tc.client.ProjectSummary(deadlineCtx(t), tc.project, opts)
	if err != nil {
		t.Fatal(err)
	}
	return sum
}

func TestForgeMergedFailureClearsMergedOK(t *testing.T) {
	for _, tc := range openFailureCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			tc.fail("open")
			tc.fail("merged")
			sum, err := tc.client.ProjectSummary(deadlineCtx(t), tc.project,
				SummaryOpts{Now: time.Now, Cache: NewTTLCache(time.Now), MergedTTL: time.Hour})
			if err != nil {
				t.Fatal(err)
			}
			if sum.MergedOK {
				t.Fatal("a failed merged load must not report MergedOK")
			}
			// Both failures appear, not only the first one.
			if !strings.Contains(sum.Error, "open pull requests failed") || !strings.Contains(sum.Error, "merged pull requests failed") &&
				!strings.Contains(sum.Error, "completed pull requests failed") {
				t.Fatalf("summary error must list every failed slice: %q", sum.Error)
			}
		})
	}
}

func TestTTLCacheNilLoader(t *testing.T) {
	c := NewTTLCache(time.Now)
	if _, err := c.GetOrLoadHeads("k", time.Minute, false, nil); !errors.Is(err, ErrNilLoader) {
		t.Fatalf("heads: %v", err)
	}
	if _, err := c.GetOrLoadOpen("k", time.Minute, false, nil); !errors.Is(err, ErrNilLoader) {
		t.Fatalf("open: %v", err)
	}
	if _, err := c.GetOrLoadCI("k", time.Minute, false, nil); !errors.Is(err, ErrNilLoader) {
		t.Fatalf("ci: %v", err)
	}
	if _, err := c.GetOrLoadMerged("k", time.Minute, false, nil); !errors.Is(err, ErrNilLoader) {
		t.Fatalf("merged: %v", err)
	}
	if _, err := c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/a", NegativeTTL: time.Minute, Fresh: false}, nil); !errors.Is(err, ErrNilLoader) {
		t.Fatalf("merged branch: %v", err)
	}
	if installed, _, _ := NewAuthCache().GetOrCheck(boardconfig.HostGitHub, nil); installed {
		t.Fatal("nil auth check must not report installed")
	}
}

func TestForgeClientWithoutExecReturnsError(t *testing.T) {
	ctx := context.Background()
	p := boardconfig.Project{ID: "x", Host: boardconfig.HostGitHub, Path: "acme/app"}
	clients := map[string]interface {
		summarizer
		MergedForBranch(ctx context.Context, repo, branch string) ([]board.MergedReview, error)
	}{
		"github":      &GitHub{},
		"gitlab":      &GitLab{},
		"azuredevops": &AzureDevOps{},
		"bitbucket":   &Bitbucket{},
	}
	for name, c := range clients {
		if _, err := c.ProjectSummary(ctx, p, SummaryOpts{Now: time.Now}); !errors.Is(err, ErrClientMissing) {
			t.Errorf("%s ProjectSummary: %v", name, err)
		}
		if _, err := c.MergedForBranch(ctx, "acme/app", "feat/a"); !errors.Is(err, ErrClientMissing) {
			t.Errorf("%s MergedForBranch: %v", name, err)
		}
	}
	var gh *GitHub
	if installed, _, _ := gh.AuthStatus(ctx); installed {
		t.Error("nil GitHub must not report installed")
	}
}
