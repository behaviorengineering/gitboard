package remotegit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

const testBitbucketToken = "s3cret-bitbucket-token"

// newTestBitbucket serves h with httptest and returns a client pointed at it.
func newTestBitbucket(t *testing.T, h http.Handler, creds BitbucketCredentials) *Bitbucket {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return BitbucketConfig{
		HTTP:        srv.Client(),
		BaseURL:     srv.URL,
		Credentials: func() BitbucketCredentials { return creds },
	}.CreateBitbucket()
}

// deadlineCtx is a test context with the deadline every forge hop requires.
func deadlineCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// countingHandler answers each request with the next status in statuses
// (the last one repeats) and counts the requests.
type countingHandler struct {
	n        atomic.Int32
	statuses []int
	header   http.Header
	body     string
}

func (h *countingHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	i := int(h.n.Add(1)) - 1
	status := h.statuses[min(i, len(h.statuses)-1)]
	for k, v := range h.header {
		w.Header()[k] = v
	}
	w.WriteHeader(status)
	if status == http.StatusOK {
		_, _ = fmt.Fprint(w, h.body)
	}
}

func TestBitbucketGetDecodesAndSendsBearer(t *testing.T) {
	var auth atomic.Value
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.Store(r.Header.Get("Authorization"))
		_, _ = fmt.Fprint(w, `{"values":[{"id":3,"updated_on":"2026-01-02T00:00:00Z","source":{"branch":{"name":"feat/a"}},"links":{"html":{"href":"https://bb/pr/3"}}}]}`)
	})
	bb := newTestBitbucket(t, h, BitbucketCredentials{Token: testBitbucketToken})
	merged, err := bb.loadMerged(deadlineCtx(t), "acme/app")
	if err != nil {
		t.Fatal(err)
	}
	if len(merged) != 1 || merged[0].Branch != "feat/a" || merged[0].ID != 3 || merged[0].MergedAt != "2026-01-02T00:00:00Z" {
		t.Fatalf("merged = %+v", merged)
	}
	if got := auth.Load(); got != "Bearer "+testBitbucketToken {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestBitbucketBasicAuthFallback(t *testing.T) {
	var user, pass atomic.Value
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, _ := r.BasicAuth()
		user.Store(u)
		pass.Store(p)
		_, _ = fmt.Fprint(w, `{"values":[]}`)
	})
	bb := newTestBitbucket(t, h, BitbucketCredentials{Username: "ann", AppPassword: "app-pass"})
	if _, err := bb.loadOpenReviews(deadlineCtx(t), "acme/app"); err != nil {
		t.Fatal(err)
	}
	if user.Load() != "ann" || pass.Load() != "app-pass" {
		t.Fatalf("basic auth = %v:%v", user.Load(), pass.Load())
	}
}

func TestBitbucketUnauthorisedIsNotRetried(t *testing.T) {
	h := &countingHandler{statuses: []int{http.StatusUnauthorized}}
	bb := newTestBitbucket(t, h, BitbucketCredentials{Token: testBitbucketToken})
	_, err := bb.apiGET(deadlineCtx(t), "/repositories/acme/app")
	if got := h.n.Load(); got != 1 {
		t.Fatalf("401 sent %d requests, want 1", got)
	}
	var se *httpStatusError
	if !errors.As(err, &se) || se.status != http.StatusUnauthorized {
		t.Fatalf("err = %v, want http 401", err)
	}
	if gberr.CodeOf(err) != gberr.CodeUnavailable {
		t.Fatalf("code = %q", gberr.CodeOf(err))
	}
}

func TestBitbucketNotFoundIsTypedAndNotRetried(t *testing.T) {
	h := &countingHandler{statuses: []int{http.StatusNotFound}}
	bb := newTestBitbucket(t, h, BitbucketCredentials{Token: testBitbucketToken})
	_, err := bb.apiGET(deadlineCtx(t), "/repositories/acme/missing")
	if got := h.n.Load(); got != 1 {
		t.Fatalf("404 sent %d requests, want 1", got)
	}
	if gberr.CodeOf(err) != gberr.CodeNotFound {
		t.Fatalf("code = %q, err = %v", gberr.CodeOf(err), err)
	}
}

func TestBitbucketRateLimitRetriedAfterRetryAfter(t *testing.T) {
	h := &countingHandler{
		statuses: []int{http.StatusTooManyRequests, http.StatusOK},
		header:   http.Header{"Retry-After": []string{"1"}},
		body:     `{"mainbranch":{"name":"main"}}`,
	}
	bb := newTestBitbucket(t, h, BitbucketCredentials{Token: testBitbucketToken})
	start := time.Now()
	body, err := bb.apiGET(deadlineCtx(t), "/repositories/acme/app")
	if err != nil {
		t.Fatal(err)
	}
	if got := h.n.Load(); got != 2 {
		t.Fatalf("requests = %d, want 2", got)
	}
	if elapsed := time.Since(start); elapsed < time.Second {
		t.Fatalf("retried after %v, before the 1s Retry-After", elapsed)
	}
	if !strings.Contains(string(body), "main") {
		t.Fatalf("body = %s", body)
	}
}

func TestBitbucketRateLimitPastDeadlineFailsAtOnce(t *testing.T) {
	h := &countingHandler{
		statuses: []int{http.StatusTooManyRequests},
		header:   http.Header{"Retry-After": []string{"3600"}},
	}
	bb := newTestBitbucket(t, h, BitbucketCredentials{Token: testBitbucketToken})
	_, err := bb.apiGET(deadlineCtx(t), "/repositories/acme/app")
	if got := h.n.Load(); got != 1 {
		t.Fatalf("requests = %d, want 1 (Retry-After is past the deadline)", got)
	}
	var se *httpStatusError
	if !errors.As(err, &se) || se.status != http.StatusTooManyRequests {
		t.Fatalf("err = %v, want http 429", err)
	}
}

func TestBitbucketServerErrorRetriedThenTyped(t *testing.T) {
	h := &countingHandler{statuses: []int{http.StatusInternalServerError}}
	bb := newTestBitbucket(t, h, BitbucketCredentials{Token: testBitbucketToken})
	_, err := bb.loadCI(deadlineCtx(t), "acme/app")
	if got := h.n.Load(); got != 1+httpRetryMax {
		t.Fatalf("requests = %d, want %d", got, 1+httpRetryMax)
	}
	var se *httpStatusError
	if !errors.As(err, &se) || se.status != http.StatusInternalServerError {
		t.Fatalf("err = %v, want http 500 in the chain", err)
	}
	if gberr.CodeOf(err) != gberr.CodeUnavailable {
		t.Fatalf("code = %q", gberr.CodeOf(err))
	}
	if fields := gberr.FieldsOf(err); fields["http_status"] != "500" || fields["repo"] != "acme/app" {
		t.Fatalf("fields = %v", fields)
	}
}

func TestBitbucketRefusesContextWithoutDeadline(t *testing.T) {
	h := &countingHandler{statuses: []int{http.StatusOK}, body: `{}`}
	bb := newTestBitbucket(t, h, BitbucketCredentials{Token: testBitbucketToken})
	_, err := bb.apiGET(context.Background(), "/repositories/acme/app")
	if !errors.Is(err, cliexec.ErrMissingDeadline) {
		t.Fatalf("err = %v, want missing deadline", err)
	}
	if gberr.CodeOf(err) != gberr.CodeInvalid {
		t.Fatalf("code = %q", gberr.CodeOf(err))
	}
	if got := h.n.Load(); got != 0 {
		t.Fatalf("sent %d requests without a deadline", got)
	}
}

// No returned error may carry a credential, even when the server echoes the
// Authorization header back in its error body.
func TestBitbucketErrorsNeverCarryCredentials(t *testing.T) {
	echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		//nolint:gosec // G705: the fake echoes the header to prove it never reaches an error.
		_, _ = fmt.Fprintf(w, `{"error":{"message":"bad header %s"}}`, r.Header.Get("Authorization"))
	})
	cases := []struct {
		name  string
		creds BitbucketCredentials
		h     http.Handler
	}{
		{"token echoed", BitbucketCredentials{Token: testBitbucketToken}, echo},
		{"basic echoed", BitbucketCredentials{Username: "ann", AppPassword: testBitbucketToken}, echo},
		{"server error", BitbucketCredentials{Token: testBitbucketToken}, &countingHandler{statuses: []int{http.StatusBadGateway}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bb := newTestBitbucket(t, tc.h, tc.creds)
			p := boardconfig.Project{ID: "bb", Host: boardconfig.HostBitbucket, Path: "acme/app"}
			sum, err := bb.ProjectSummary(deadlineCtx(t), p, SummaryOpts{Fresh: true, Now: time.Now})
			if err != nil {
				t.Fatal(err)
			}
			_, mErr := bb.MergedForBranch(deadlineCtx(t), "acme/app", "feat/a")
			_, lErr := bb.ListWorkspaceRepos(deadlineCtx(t), "acme")
			for _, text := range []string{sum.Error, fmt.Sprint(mErr), fmt.Sprintf("%+v", lErr), fmt.Sprintf("%#v", tc.creds)} {
				if strings.Contains(text, testBitbucketToken) {
					t.Fatalf("credential leaked: %q", text)
				}
			}
			if sum.Error == "" || mErr == nil || lErr == nil {
				t.Fatalf("want failures: summary=%q merged=%v list=%v", sum.Error, mErr, lErr)
			}
		})
	}
}

func TestBitbucketAuthStatusChecksCredentials(t *testing.T) {
	bb := newTestBitbucket(t, http.NotFoundHandler(), BitbucketCredentials{})
	installed, authed, detail := bb.AuthStatus(context.Background())
	if !installed || authed || detail != bitbucketCredentialsHint {
		t.Fatalf("installed=%v authed=%v detail=%q", installed, authed, detail)
	}
}

func TestBitbucketListWorkspaceFollowsNextAndRefusesOffHost(t *testing.T) {
	var srvURL string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "":
			_, _ = fmt.Fprintf(w, `{"next":%q,"values":[{"full_name":"acme/one","name":"One"}]}`, srvURL+"/repositories/acme?pagelen=100&page=2")
		case "2":
			_, _ = fmt.Fprint(w, `{"next":"https://elsewhere.example/repositories/acme?page=3","values":[{"slug":"two"}]}`)
		default:
			t.Errorf("unexpected request %s", r.URL)
		}
	})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	srvURL = srv.URL
	bb := BitbucketConfig{
		HTTP: srv.Client(), BaseURL: srv.URL,
		Credentials: func() BitbucketCredentials { return BitbucketCredentials{Token: testBitbucketToken} },
	}.CreateBitbucket()
	_, err := bb.ListWorkspaceRepos(deadlineCtx(t), "acme")
	if err == nil || !strings.Contains(err.Error(), "unusable next page link") {
		t.Fatalf("off-host next link must fail, got %v", err)
	}
}

func TestBitbucketListWorkspaceFollowsNextToTheEnd(t *testing.T) {
	var srvURL string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			_, _ = fmt.Fprint(w, `{"values":[{"slug":"two"}]}`)
			return
		}
		_, _ = fmt.Fprintf(w, `{"next":%q,"values":[{"full_name":"acme/one","name":"One"}]}`, srvURL+"/repositories/acme?pagelen=100&page=2")
	})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	srvURL = srv.URL
	bb := BitbucketConfig{
		HTTP: srv.Client(), BaseURL: srv.URL,
		Credentials: func() BitbucketCredentials { return BitbucketCredentials{Token: testBitbucketToken} },
	}.CreateBitbucket()
	refs, err := bb.ListWorkspaceRepos(deadlineCtx(t), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 2 || refs[0] != (RepoRef{Host: boardconfig.HostBitbucket, Path: "acme/one", Name: "One"}) ||
		refs[1] != (RepoRef{Host: boardconfig.HostBitbucket, Path: "acme/two", Name: "two"}) {
		t.Fatalf("refs = %+v", refs)
	}
}
