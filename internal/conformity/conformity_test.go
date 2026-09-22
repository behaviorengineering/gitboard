// Package conformity runs network-retrieval contract tests across forge adapters.
package conformity

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
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

func (c *countingExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	c.n.Add(1)
	c.mu.Lock()
	c.log = append(c.log, name+" "+strings.Join(args, " "))
	c.mu.Unlock()
	return c.inner.Run(ctx, name, args...)
}

func (c *countingExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	c.n.Add(1)
	c.mu.Lock()
	c.log = append(c.log, name+" "+strings.Join(args, " "))
	c.mu.Unlock()
	return c.inner.RunJSON(ctx, name, args...)
}

func (c *countingExec) calls() int64 { return c.n.Load() }

func (c *countingExec) reset() {
	c.n.Store(0)
	c.mu.Lock()
	c.log = nil
	c.mu.Unlock()
}

// AdapterCase is one forge registered in the conformity matrix.
type AdapterCase struct {
	Name    string
	Host    config.Host
	Path    string
	NewExec func() cliexec.Exec
	Wire    func(run cliexec.Exec) *dashboard.Service
	Setup   func(t *testing.T)
}

func matrix() []AdapterCase {
	return []AdapterCase{
		{
			Name: "github",
			Host: config.HostGitHub,
			Path: "acme/app",
			NewExec: func() cliexec.Exec {
				return remotegit.ConformityGitHubExec()
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(remotegit.NewGitHub(run), nil, nil, nil, nil)
			},
		},
		{
			Name: "gitlab",
			Host: config.HostGitLab,
			Path: "acme/app",
			NewExec: func() cliexec.Exec {
				return remotegit.ConformityGitLabExec()
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(nil, remotegit.NewGitLab(run), nil, nil, nil)
			},
		},
		{
			Name: "azuredevops",
			Host: config.HostAzureDevOps,
			Path: "acme/proj/app",
			NewExec: func() cliexec.Exec {
				return remotegit.ConformityAzureExec()
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(nil, nil, remotegit.NewAzureDevOps(run), nil, nil)
			},
		},
		{
			Name: "bitbucket",
			Host: config.HostBitbucket,
			Path: "acme/app",
			NewExec: func() cliexec.Exec {
				return remotegit.ConformityBitbucketExec()
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(nil, nil, nil, remotegit.NewBitbucket(run), nil)
			},
			Setup: func(t *testing.T) {
				t.Setenv("BITBUCKET_TOKEN", "conformity-token")
			},
		},
	}
}

func TestWarmPollBudget(t *testing.T) {
	for _, tc := range matrix() {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Setup != nil {
				tc.Setup(t)
			}
			inner := tc.NewExec()
			exec := &countingExec{inner: inner}
			svc := tc.Wire(exec)
			heads := 120
			merged := 600
			doc := config.File{
				Upstream: config.Upstream{HeadsSeconds: &heads, MergedSeconds: &merged},
				Projects: []config.Project{{
					ID: "p1", Label: "P1", Host: tc.Host, Path: tc.Path,
				}},
			}
			ctx := context.Background()
			if _, err := svc.Collect(ctx, doc, false, ""); err != nil {
				t.Fatalf("cold: %v", err)
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
			if tc.Setup != nil {
				tc.Setup(t)
			}
			if tc.Name == "bitbucket" {
				t.Skip("bitbucket auth is env-based (no CLI auth status)")
			}
			inner := tc.NewExec()
			exec := &countingExec{inner: inner}
			svc := tc.Wire(exec)
			doc := config.File{
				Projects: []config.Project{
					{ID: "a", Label: "A", Host: tc.Host, Path: tc.Path},
					{ID: "b", Label: "B", Host: tc.Host, Path: tc.Path},
				},
			}
			ctx := context.Background()
			if _, err := svc.Collect(ctx, doc, true, ""); err != nil {
				t.Fatalf("collect: %v", err)
			}
			authCalls := 0
			exec.mu.Lock()
			for _, line := range exec.log {
				if strings.Contains(line, "auth status") || strings.Contains(line, "account show") {
					authCalls++
				}
			}
			exec.mu.Unlock()
			if authCalls != 1 {
				t.Fatalf("want 1 auth call per host per request, got %d log=%v", authCalls, exec.log)
			}
		})
	}
}

func TestHeadsStampedeCoalesce(t *testing.T) {
	for _, tc := range matrix() {
		t.Run(tc.Name, func(t *testing.T) {
			loads := atomic.Int64{}
			cache := remotegit.NewTTLCache()
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
			if tc.Setup != nil {
				tc.Setup(t)
			}
			inner := tc.NewExec()
			exec := &countingExec{inner: inner}
			svc := tc.Wire(exec)
			heads := 600
			doc := config.File{
				Upstream: config.Upstream{HeadsSeconds: &heads},
				Projects: []config.Project{{
					ID: "p1", Label: "P1", Host: tc.Host, Path: tc.Path,
				}},
			}
			ctx := context.Background()
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
	cache := remotegit.NewTTLCache()
	_, ok := cache.GetOrLoadMerged("k", time.Minute, true, func() ([]board.MergedReview, error) {
		return nil, fmt.Errorf("forge down")
	})
	if ok {
		t.Fatal("failed merged load must not report ok")
	}
}

func TestLiveOptInDocumented(t *testing.T) {
	// Live mode is env-gated; this test documents the contract without requiring network.
	if os.Getenv("GITBOARD_CONFORMITY_LIVE") == "" {
		t.Log("set GITBOARD_CONFORMITY_LIVE=1 with logged-in forge CLIs to run live cases")
	}
}
