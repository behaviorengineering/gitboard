// Package conformity runs network-retrieval contract tests across forge adapters.
package conformity

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit/remotegittest"
)

// countingExec wraps an Exec and counts process invocations.
type countingExec struct {
	inner cliexec.Exec
	n     atomic.Int64
	mu    sync.Mutex
	log   []string
}

func (c *countingExec) LookPath(name string) (string, error) {
	return c.inner.LookPath(name)
}

// record counts one forge hop: a CLI process or an HTTP request.
func (c *countingExec) record(line string) {
	c.n.Add(1)
	c.mu.Lock()
	c.log = append(c.log, line)
	c.mu.Unlock()
}

func (c *countingExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	c.record(name + " " + strings.Join(args, " "))
	return c.inner.Run(ctx, name, args...)
}

func (c *countingExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	c.record(name + " " + strings.Join(args, " "))
	return c.inner.RunJSON(ctx, name, args...)
}

// countingHandler records each HTTP request on exec before serving it, so
// REST forges count against the same hop budget as CLI forges.
func countingHandler(exec *countingExec, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exec.record(r.Method + " " + r.URL.RequestURI())
		inner.ServeHTTP(w, r)
	})
}

func (c *countingExec) calls() int64 { return c.n.Load() }

func (c *countingExec) reset() {
	c.n.Store(0)
	c.mu.Lock()
	c.log = nil
	c.mu.Unlock()
}

// AdapterCase is one forge registered in the conformity matrix. Wire builds
// the service on exec; a REST forge also routes its HTTP requests through
// exec's counter.
type AdapterCase struct {
	Name    string
	Host    boardconfig.Host
	Path    string
	NewExec func() cliexec.Exec
	Wire    func(t *testing.T, exec *countingExec) *dashboard.Service
}

// wireCLI builds the default service on the counting exec.
func wireCLI(_ *testing.T, exec *countingExec) *dashboard.Service {
	return dashboard.DefaultServiceConfig(exec).CreateService()
}

// wireBitbucket serves the Bitbucket fixture over httptest and points the
// service's Bitbucket client at it.
func wireBitbucket(t *testing.T, exec *countingExec) *dashboard.Service {
	t.Helper()
	srv := httptest.NewServer(countingHandler(exec, remotegittest.ConformityBitbucketHTTP()))
	t.Cleanup(srv.Close)
	cfg := dashboard.DefaultServiceConfig(exec)
	cfg.Bitbucket = remotegit.BitbucketConfig{
		HTTP:    srv.Client(),
		BaseURL: srv.URL,
		Credentials: func() remotegit.BitbucketCredentials {
			return remotegit.BitbucketCredentials{Token: "conformity-token"}
		},
	}.CreateBitbucket()
	return cfg.CreateService()
}

// collectCtx is the request budget the server would set for one collect.
func collectCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func matrix() []AdapterCase {
	return []AdapterCase{
		{
			Name: "github",
			Host: boardconfig.HostGitHub,
			Path: "acme/app",
			NewExec: func() cliexec.Exec {
				return remotegittest.ConformityGitHubExec()
			},
			Wire: wireCLI,
		},
		{
			Name: "gitlab",
			Host: boardconfig.HostGitLab,
			Path: "acme/app",
			NewExec: func() cliexec.Exec {
				return remotegittest.ConformityGitLabExec()
			},
			Wire: wireCLI,
		},
		{
			Name: "azuredevops",
			Host: boardconfig.HostAzureDevOps,
			Path: "acme/proj/app",
			NewExec: func() cliexec.Exec {
				return remotegittest.ConformityAzureExec()
			},
			Wire: wireCLI,
		},
		{
			Name: "bitbucket",
			Host: boardconfig.HostBitbucket,
			Path: "acme/app",
			// Bitbucket is REST over net/http: no CLI calls, so an empty
			// exec fails any unexpected argv.
			NewExec: func() cliexec.Exec {
				return &remotegittest.FakeExec{}
			},
			Wire: wireBitbucket,
		},
	}
}

func TestWarmPollBudget(t *testing.T) {
	for _, tc := range matrix() {
		t.Run(tc.Name, func(t *testing.T) {
			inner := tc.NewExec()
			exec := &countingExec{inner: inner}
			svc := tc.Wire(t, exec)
			heads := 120
			merged := 600
			doc := boardconfig.File{
				Upstream: boardconfig.Upstream{HeadsSeconds: &heads, MergedSeconds: &merged},
				Projects: []boardconfig.Project{{
					ID: "p1", Label: "P1", Host: tc.Host, Path: tc.Path,
				}},
			}
			ctx := collectCtx(t)
			dash, err := svc.Collect(ctx, doc, false, "")
			if err != nil {
				t.Fatalf("cold: %v", err)
			}
			if row := dash.Projects[0]; row.Error != "" || !row.RemoteNamesOK {
				t.Fatalf("cold row: error=%q remote_names_ok=%v log=%v", row.Error, row.RemoteNamesOK, exec.log)
			}
			cold := exec.calls()
			if cold < 1 {
				t.Fatalf("cold should call forge: %d", cold)
			}
			exec.reset()
			if _, err := svc.Collect(ctx, doc, false, ""); err != nil {
				t.Fatalf("warm: %v", err)
			}
			warm := exec.calls()
			if warm >= cold {
				t.Fatalf("warm calls should drop below cold: warm=%d cold=%d log=%v", warm, cold, exec.log)
			}
		})
	}
}

func TestAuthOncePerHostPerRequest(t *testing.T) {
	for _, tc := range matrix() {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Name == "bitbucket" {
				t.Skip("bitbucket auth checks credentials and makes no request")
			}
			inner := tc.NewExec()
			exec := &countingExec{inner: inner}
			svc := tc.Wire(t, exec)
			doc := boardconfig.File{
				Projects: []boardconfig.Project{
					{ID: "a", Label: "A", Host: tc.Host, Path: tc.Path},
					{ID: "b", Label: "B", Host: tc.Host, Path: tc.Path},
				},
			}
			ctx := collectCtx(t)
			if _, err := svc.Collect(ctx, doc, true, ""); err != nil {
				t.Fatalf("collect: %v", err)
			}
			assertOneAuthProbePerForge(t, exec)
		})
	}
}

// assertOneAuthProbePerForge checks that Collect probed every forge's tooling,
// and each probe ran once even with two projects on the same host.
func assertOneAuthProbePerForge(t *testing.T, exec *countingExec) {
	t.Helper()
	authCalls := map[string]int{}
	exec.mu.Lock()
	defer exec.mu.Unlock()
	for _, line := range exec.log {
		if strings.Contains(line, "auth status") || strings.Contains(line, "account show") {
			authCalls[line]++
		}
	}
	if len(authCalls) == 0 {
		t.Fatalf("want auth probes, got none log=%v", exec.log)
	}
	for line, n := range authCalls {
		if n != 1 {
			t.Fatalf("want 1 %q per request, got %d log=%v", line, n, exec.log)
		}
	}
}

func TestHeadsStampedeCoalesce(t *testing.T) {
	for _, tc := range matrix() {
		t.Run(tc.Name, func(t *testing.T) {
			loads := atomic.Int64{}
			cache := remotegit.NewTTLCache(time.Now)
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := cache.GetOrLoadHeads(string(tc.Host)+"/"+tc.Path, time.Minute, true, func() (remotegit.HeadsSnapshot, error) {
						loads.Add(1)
						time.Sleep(20 * time.Millisecond)
						return remotegit.HeadsSnapshot{DefaultBranch: "main"}, nil
					})
					if err != nil {
						t.Errorf("load: %v", err)
					}
				}()
			}
			wg.Wait()
			if got := loads.Load(); got != 1 {
				t.Fatalf("expected singleflight coalesce to 1 load, got %d", got)
			}
		})
	}
}

func TestFreshBypassesTTL(t *testing.T) {
	for _, tc := range matrix() {
		t.Run(tc.Name, func(t *testing.T) {
			inner := tc.NewExec()
			exec := &countingExec{inner: inner}
			svc := tc.Wire(t, exec)
			heads := 600
			doc := boardconfig.File{
				Upstream: boardconfig.Upstream{HeadsSeconds: &heads},
				Projects: []boardconfig.Project{{
					ID: "p1", Label: "P1", Host: tc.Host, Path: tc.Path,
				}},
			}
			ctx := collectCtx(t)
			if _, err := svc.Collect(ctx, doc, false, ""); err != nil {
				t.Fatalf("warm seed: %v", err)
			}
			exec.reset()
			if _, err := svc.Collect(ctx, doc, true, ""); err != nil {
				t.Fatalf("fresh: %v", err)
			}
			if exec.calls() < 1 {
				t.Fatalf("fresh must hit forge")
			}
		})
	}
}

func TestFailClosedMerged(t *testing.T) {
	cache := remotegit.NewTTLCache(time.Now)
	_, err := cache.GetOrLoadMerged("k", time.Minute, true, func() ([]board.MergedReview, error) {
		return nil, fmt.Errorf("forge down")
	})
	if err == nil {
		t.Fatal("failed merged load must return its error")
	}
}

func TestLiveDisabledByDefault(t *testing.T) {
	if os.Getenv("GITBOARD_CONFORMITY_LIVE") == "1" {
		t.Skip("live mode enabled; skip default-disabled assertion")
	}
	// Ordinary go test / make conformity must not require network credentials.
}
