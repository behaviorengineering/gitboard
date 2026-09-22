package conformity

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// Live forge cases run only when GITBOARD_CONFORMITY_LIVE=1.
// Each host is skipped when its project path env is unset or credentials are missing.
// Never enable this in ordinary PR CI.

type liveCase struct {
	Name      string
	Host      config.Host
	PathEnv   string
	CredCheck func(t *testing.T) bool
	Wire      func(run cliexec.Exec) *dashboard.Service
}

func liveMatrix() []liveCase {
	return []liveCase{
		{
			Name:    "github",
			Host:    config.HostGitHub,
			PathEnv: "GITBOARD_LIVE_GITHUB_PATH",
			CredCheck: func(t *testing.T) bool {
				t.Helper()
				run := cliexec.New()
				if _, err := run.LookPath("gh"); err != nil {
					t.Log("skip github: gh not installed")
					return false
				}
				gh := remotegit.NewGitHub(run)
				_, authed, detail := gh.AuthStatus(context.Background())
				if !authed {
					t.Logf("skip github: not authenticated (%s)", sanitizeDetail(detail))
					return false
				}
				return true
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(remotegit.NewGitHub(run), nil, nil, nil, nil)
			},
		},
		{
			Name:    "gitlab",
			Host:    config.HostGitLab,
			PathEnv: "GITBOARD_LIVE_GITLAB_PATH",
			CredCheck: func(t *testing.T) bool {
				t.Helper()
				run := cliexec.New()
				if _, err := run.LookPath("glab"); err != nil {
					t.Log("skip gitlab: glab not installed")
					return false
				}
				gl := remotegit.NewGitLab(run)
				_, authed, detail := gl.AuthStatus(context.Background())
				if !authed {
					t.Logf("skip gitlab: not authenticated (%s)", sanitizeDetail(detail))
					return false
				}
				return true
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(nil, remotegit.NewGitLab(run), nil, nil, nil)
			},
		},
		{
			Name:    "azuredevops",
			Host:    config.HostAzureDevOps,
			PathEnv: "GITBOARD_LIVE_AZURE_PATH",
			CredCheck: func(t *testing.T) bool {
				t.Helper()
				run := cliexec.New()
				if _, err := run.LookPath("az"); err != nil {
					t.Log("skip azuredevops: az not installed")
					return false
				}
				az := remotegit.NewAzureDevOps(run)
				_, authed, detail := az.AuthStatus(context.Background())
				if !authed {
					t.Logf("skip azuredevops: not authenticated (%s)", sanitizeDetail(detail))
					return false
				}
				return true
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(nil, nil, remotegit.NewAzureDevOps(run), nil, nil)
			},
		},
		{
			Name:    "bitbucket",
			Host:    config.HostBitbucket,
			PathEnv: "GITBOARD_LIVE_BITBUCKET_PATH",
			CredCheck: func(t *testing.T) bool {
				t.Helper()
				if strings.TrimSpace(os.Getenv("BITBUCKET_TOKEN")) == "" {
					t.Log("skip bitbucket: BITBUCKET_TOKEN unset")
					return false
				}
				return true
			},
			Wire: func(run cliexec.Exec) *dashboard.Service {
				return dashboard.New(nil, nil, nil, remotegit.NewBitbucket(run), nil)
			},
		},
	}
}

func TestLiveAdapterContract(t *testing.T) {
	if os.Getenv("GITBOARD_CONFORMITY_LIVE") != "1" {
		t.Skip("set GITBOARD_CONFORMITY_LIVE=1 to run live forge conformity")
	}

	for _, tc := range liveMatrix() {
		t.Run(tc.Name, func(t *testing.T) {
			path := strings.TrimSpace(os.Getenv(tc.PathEnv))
			if path == "" {
				t.Skipf("%s unset", tc.PathEnv)
			}
			if !tc.CredCheck(t) {
				t.Skip("credentials unavailable")
			}

			inner := cliexec.New()
			exec := &countingExec{inner: inner}
			svc := tc.Wire(exec)
			doc := config.File{
				Projects: []config.Project{{
					ID: "live-" + tc.Name, Label: tc.Name, Host: tc.Host, Path: path,
				}},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			start := time.Now()
			dash, err := svc.Collect(ctx, doc, true, "")
			elapsed := time.Since(start)
			if err != nil {
				t.Fatalf("collect: %v", err)
			}
			if len(dash.Projects) != 1 {
				t.Fatalf("want 1 project row, got %d", len(dash.Projects))
			}
			row := dash.Projects[0]
			if row.Error != "" {
				t.Fatalf("project error: %s", row.Error)
			}
			if !row.RemoteNamesOK {
				t.Fatal("RemoteNamesOK=false")
			}
			caps := remotegit.HostCapabilities(tc.Host)
			if row.Capabilities.FailedJobs != caps.FailedJobs {
				t.Fatalf("capabilities.failed_jobs=%v want %v", row.Capabilities.FailedJobs, caps.FailedJobs)
			}

			t.Logf("provider=%s project=%s calls=%d latency_ms=%d branches=%d",
				tc.Name, path, exec.calls(), elapsed.Milliseconds(), len(row.Branches))
		})
	}
}

func sanitizeDetail(detail string) string {
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return "no detail"
	}
	// Truncate; never echo tokens that might appear in CLI stderr.
	if len(detail) > 120 {
		return detail[:120] + "…"
	}
	return detail
}
