package remotegit_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit/remotegittest"
)

func TestHostCapabilitiesFailedJobsMatrix(t *testing.T) {
	cases := []struct {
		host boardconfig.Host
		want bool
	}{
		{boardconfig.HostGitHub, true},
		{boardconfig.HostGitLab, true},
		{boardconfig.HostAzureDevOps, false},
		{boardconfig.HostBitbucket, false},
		{boardconfig.Host("unknown"), false},
	}
	for _, tc := range cases {
		got := remotegit.HostCapabilities(tc.host)
		if got.FailedJobs != tc.want {
			t.Fatalf("%s: FailedJobs=%v want %v", tc.host, got.FailedJobs, tc.want)
		}
	}
}

func TestProjectSummaryAdvertisesFailedJobsCapability(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := remotegit.SummaryOpts{Fresh: true, Now: time.Now}
	bbSrv := httptest.NewServer(remotegittest.ConformityBitbucketHTTP())
	defer bbSrv.Close()
	bb := remotegit.BitbucketConfig{
		HTTP: bbSrv.Client(), BaseURL: bbSrv.URL,
		Credentials: func() remotegit.BitbucketCredentials {
			return remotegit.BitbucketCredentials{Token: "conformity-token"}
		},
	}.CreateBitbucket()

	cases := []struct {
		name   string
		host   boardconfig.Host
		path   string
		want   bool
		client remotegit.Client
	}{
		{"github", boardconfig.HostGitHub, "acme/app", true, remotegit.NewGitHub(remotegittest.ConformityGitHubExec())},
		{"gitlab", boardconfig.HostGitLab, "acme/app", true, remotegit.NewGitLab(remotegittest.ConformityGitLabExec())},
		{"azuredevops", boardconfig.HostAzureDevOps, "acme/proj/app", false, remotegit.NewAzureDevOps(remotegittest.ConformityAzureExec())},
		{"bitbucket", boardconfig.HostBitbucket, "acme/app", false, bb},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := boardconfig.Project{ID: "p1", Label: "P1", Host: tc.host, Path: tc.path}
			row, err := tc.client.ProjectSummary(ctx, p, opts)
			if err != nil {
				t.Fatalf("summary: %v", err)
			}
			if row.Capabilities.FailedJobs != tc.want {
				t.Fatalf("FailedJobs=%v want %v", row.Capabilities.FailedJobs, tc.want)
			}
		})
	}
}
