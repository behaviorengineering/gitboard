package boardconfig

import (
	"errors"
	"testing"
)

func TestParseHost(t *testing.T) {
	for _, in := range []string{"github", " GitLab ", "AZUREDEVOPS", "bitbucket"} {
		h, err := ParseHost(in)
		if err != nil || !h.Valid() {
			t.Fatalf("ParseHost(%q) = %q, %v", in, h, err)
		}
	}
	for _, in := range []string{"", "gitea", "git hub"} {
		if _, err := ParseHost(in); !errors.Is(err, ErrUnknownHost) {
			t.Fatalf("ParseHost(%q) err = %v, want ErrUnknownHost", in, err)
		}
	}
	if len(Hosts()) != 4 {
		t.Fatalf("Hosts() = %v", Hosts())
	}
}

func TestSplitRepoPath(t *testing.T) {
	cases := []struct {
		host Host
		path string
		want int
		ok   bool
	}{
		{HostGitHub, "acme/app", 2, true},
		{HostGitLab, "/group/sub/app/", 3, true},
		{HostBitbucket, "ws/repo", 2, true},
		{HostAzureDevOps, "org/proj/repo", 3, true},
		{HostAzureDevOps, "org/repo", 0, false},
		{HostGitHub, "app", 0, false},
		{HostGitHub, "acme//app", 0, false},
	}
	for _, tc := range cases {
		parts, err := SplitRepoPath(tc.host, tc.path)
		if (err == nil) != tc.ok || len(parts) != tc.want {
			t.Fatalf("SplitRepoPath(%s, %q) = %v, %v", tc.host, tc.path, parts, err)
		}
	}
}

func TestUsesMergeRequests(t *testing.T) {
	if !HostGitLab.UsesMergeRequests() || HostGitHub.UsesMergeRequests() || HostBitbucket.UsesMergeRequests() {
		t.Fatal("only GitLab counts merge requests")
	}
}
