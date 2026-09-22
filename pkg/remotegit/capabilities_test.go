package remotegit_test

import (
	"context"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

func TestHostCapabilitiesFailedJobsMatrix(t *testing.T) {
	cases := []struct {
		host config.Host
		want bool
	}{
		{config.HostGitHub, true},
		{config.HostGitLab, true},
		{config.HostAzureDevOps, false},
		{config.HostBitbucket, false},
		{config.Host("unknown"), false},
	}
	for _, tc := range cases {
		got := remotegit.HostCapabilities(tc.host)
		if got.FailedJobs != tc.want {
			t.Fatalf("%s: FailedJobs=%v want %v", tc.host, got.FailedJobs, tc.want)
		}
	}
}

func TestProjectSummaryAdvertisesFailedJobsCapability(t *testing.T) {
	t.Setenv("BITBUCKET_TOKEN", "conformity-token")
	ctx := context.Background()
	opts := remotegit.SummaryOpts{Fresh: true}

	cases := []struct {
		name   string
		host   config.Host
		path   string
		want   bool
		client remotegit.Client
	}{
		{"github", config.HostGitHub, "acme/app", true, remotegit.NewGitHub(remotegit.ConformityGitHubExec())},
		{"gitlab", config.HostGitLab, "acme/app", true, remotegit.NewGitLab(remotegit.ConformityGitLabExec())},
		{"azuredevops", config.HostAzureDevOps, "acme/proj/app", false, remotegit.NewAzureDevOps(remotegit.ConformityAzureExec())},
		{"bitbucket", config.HostBitbucket, "acme/app", false, remotegit.NewBitbucket(remotegit.ConformityBitbucketExec())},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := config.Project{ID: "p1", Label: "P1", Host: tc.host, Path: tc.path}
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
