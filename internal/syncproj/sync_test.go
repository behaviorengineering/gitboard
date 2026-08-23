package syncproj_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/forge"
	"github.com/behaviorengineering/gitboard/internal/syncproj"
)

type fakeLister struct {
	gh map[string][]forge.RepoRef
	gl map[string][]forge.RepoRef
}

func (f fakeLister) ListGitHub(_ context.Context, org string) ([]forge.RepoRef, error) {
	return f.gh[org], nil
}

func (f fakeLister) ListGitLab(_ context.Context, group string) ([]forge.RepoRef, error) {
	return f.gl[group], nil
}

func TestDiscoverAndApplySelection(t *testing.T) {
	doc := config.File{
		Sync: config.SyncSources{
			GitHub: config.GitHubSync{Orgs: []string{"acme"}},
			GitLab: config.GitLabSync{Groups: []string{"acme"}},
		},
		Projects: []config.Project{{
			ID: "alpha", Label: "alpha", Host: config.HostGitHub, Path: "acme/alpha",
		}},
	}
	lister := fakeLister{
		gh: map[string][]forge.RepoRef{
			"acme": {
				{Host: config.HostGitHub, Path: "acme/alpha", Name: "alpha"},
				{Host: config.HostGitHub, Path: "acme/beta", Name: "beta"},
			},
		},
		gl: map[string][]forge.RepoRef{
			"acme": {
				{Host: config.HostGitLab, Path: "acme/gamma", Name: "gamma"},
			},
		},
	}
	cands, err := syncproj.Discover(context.Background(), lister, doc, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 3 {
		t.Fatalf("candidates=%d want 3", len(cands))
	}
	tracked := 0
	for _, c := range cands {
		if c.Tracked {
			tracked++
			if c.Path != "acme/alpha" {
				t.Fatalf("unexpected tracked %s", c.Path)
			}
		}
	}
	if tracked != 1 {
		t.Fatalf("tracked=%d", tracked)
	}

	selected, err := syncproj.ParseSelection("2,3", cands)
	if err != nil {
		t.Fatal(err)
	}
	projects, err := syncproj.ApplySelection(cands, selected, doc.Projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects=%d", len(projects))
	}
}

func TestParseSelectionEmptyKeepsTracked(t *testing.T) {
	cands := []syncproj.Candidate{
		{Index: 1, Tracked: true, Host: config.HostGitHub, Path: "a/b"},
		{Index: 2, Tracked: false, Host: config.HostGitHub, Path: "a/c"},
	}
	got, err := syncproj.ParseSelection("", cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("got %#v", got)
	}
}

func TestParseSelectionAll(t *testing.T) {
	cands := []syncproj.Candidate{{Index: 1}, {Index: 2}}
	got, err := syncproj.ParseSelection("all", cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
}

func TestAddAndRemoveProject(t *testing.T) {
	var projects []config.Project
	var err error
	projects, err = syncproj.AddProject(projects, config.HostGitHub, "org/repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].ID != "repo" {
		t.Fatalf("%+v", projects)
	}
	projects, err = syncproj.AddProject(projects, config.HostGitLab, "org/repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("want 2 got %d", len(projects))
	}
	// Same host+path updates in place.
	projects, err = syncproj.AddProject(projects, config.HostGitHub, "org/repo")
	if err != nil || len(projects) != 2 {
		t.Fatalf("update: %v len=%d", err, len(projects))
	}
	projects, found := syncproj.RemoveProject(projects, "repo")
	if !found || len(projects) != 1 {
		t.Fatalf("remove: found=%v len=%d", found, len(projects))
	}
}

func TestFormatCandidates(t *testing.T) {
	var buf bytes.Buffer
	syncproj.FormatCandidates(&buf, []syncproj.Candidate{
		{Index: 1, Tracked: true, Host: config.HostGitHub, Path: "a/b"},
	})
	if !strings.Contains(buf.String(), "[1] [x] github a/b") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestDiscoverHostFilter(t *testing.T) {
	doc := config.File{
		Sync: config.SyncSources{
			GitHub: config.GitHubSync{Orgs: []string{"acme"}},
			GitLab: config.GitLabSync{Groups: []string{"acme"}},
		},
	}
	lister := fakeLister{
		gh: map[string][]forge.RepoRef{
			"acme": {{Host: config.HostGitHub, Path: "acme/a", Name: "a"}},
		},
		gl: map[string][]forge.RepoRef{
			"acme": {{Host: config.HostGitLab, Path: "acme/b", Name: "b"}},
		},
	}
	cands, err := syncproj.Discover(context.Background(), lister, doc, "github")
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 1 || cands[0].Host != config.HostGitHub {
		t.Fatalf("%+v", cands)
	}
}
