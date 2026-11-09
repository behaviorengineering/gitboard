package projectsync_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

type fakeLister struct {
	gh    map[string][]remotegit.RepoRef
	gl    map[string][]remotegit.RepoRef
	az    map[string][]remotegit.RepoRef
	bb    map[string][]remotegit.RepoRef
	ghErr map[string]error
	glErr map[string]error
	azErr map[string]error
	bbErr map[string]error
}

func (f fakeLister) ListGitHub(_ context.Context, org string) ([]remotegit.RepoRef, error) {
	if f.ghErr != nil {
		if err := f.ghErr[org]; err != nil {
			return nil, err
		}
	}
	return f.gh[org], nil
}

func (f fakeLister) ListGitLab(_ context.Context, group string) ([]remotegit.RepoRef, error) {
	if f.glErr != nil {
		if err := f.glErr[group]; err != nil {
			return nil, err
		}
	}
	return f.gl[group], nil
}

func (f fakeLister) ListAzureDevOps(_ context.Context, org string) ([]remotegit.RepoRef, error) {
	if f.azErr != nil {
		if err := f.azErr[org]; err != nil {
			return nil, err
		}
	}
	return f.az[org], nil
}

func (f fakeLister) ListBitbucket(_ context.Context, workspace string) ([]remotegit.RepoRef, error) {
	if f.bbErr != nil {
		if err := f.bbErr[workspace]; err != nil {
			return nil, err
		}
	}
	return f.bb[workspace], nil
}

func TestDiscoverAndApplySelection(t *testing.T) {
	doc := boardconfig.File{
		Sync: boardconfig.SyncSources{
			GitHub: boardconfig.GitHubSync{Orgs: []string{"acme"}},
			GitLab: boardconfig.GitLabSync{Groups: []string{"acme"}},
		},
		Projects: []boardconfig.Project{{
			ID: "alpha", Label: "alpha", Host: boardconfig.HostGitHub, Path: "acme/alpha",
		}},
	}
	lister := fakeLister{
		gh: map[string][]remotegit.RepoRef{
			"acme": {
				{Host: boardconfig.HostGitHub, Path: "acme/alpha", Name: "alpha"},
				{Host: boardconfig.HostGitHub, Path: "acme/beta", Name: "beta"},
			},
		},
		gl: map[string][]remotegit.RepoRef{
			"acme": {
				{Host: boardconfig.HostGitLab, Path: "acme/gamma", Name: "gamma"},
			},
		},
	}
	res, err := projectsync.Discover(context.Background(), lister, doc, "")
	if err != nil {
		t.Fatal(err)
	}
	cands := res.Candidates
	if len(res.Warnings) != 0 {
		t.Fatalf("warnings=%v", res.Warnings)
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

	refs := []projectsync.RepoRef{
		{Host: cands[1].Host, Path: cands[1].Path},
		{Host: cands[2].Host, Path: cands[2].Path},
	}
	projects, err := projectsync.ApplySelectionByRefs(cands, refs, doc.Projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects=%d", len(projects))
	}
}

func TestAddAndRemoveProject(t *testing.T) {
	var projects []boardconfig.Project
	var err error
	projects, err = projectsync.AddProject(projects, boardconfig.HostGitHub, "org/repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].ID != "repo" {
		t.Fatalf("%+v", projects)
	}
	projects, err = projectsync.AddProject(projects, boardconfig.HostGitLab, "org/repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("want 2 got %d", len(projects))
	}
	// Same host+path updates in place.
	projects, err = projectsync.AddProject(projects, boardconfig.HostGitHub, "org/repo")
	if err != nil || len(projects) != 2 {
		t.Fatalf("update: %v len=%d", err, len(projects))
	}
	projects, found := projectsync.RemoveProject(projects, "repo")
	if !found || len(projects) != 1 {
		t.Fatalf("remove: found=%v len=%d", found, len(projects))
	}
}

func TestDiscoverHostFilter(t *testing.T) {
	doc := boardconfig.File{
		Sync: boardconfig.SyncSources{
			GitHub: boardconfig.GitHubSync{Orgs: []string{"acme"}},
			GitLab: boardconfig.GitLabSync{Groups: []string{"acme"}},
		},
	}
	lister := fakeLister{
		gh: map[string][]remotegit.RepoRef{
			"acme": {{Host: boardconfig.HostGitHub, Path: "acme/a", Name: "a"}},
		},
		gl: map[string][]remotegit.RepoRef{
			"acme": {{Host: boardconfig.HostGitLab, Path: "acme/b", Name: "b"}},
		},
	}
	res, err := projectsync.Discover(context.Background(), lister, doc, "github")
	if err != nil {
		t.Fatal(err)
	}
	cands := res.Candidates
	if len(cands) != 1 || cands[0].Host != boardconfig.HostGitHub {
		t.Fatalf("%+v", cands)
	}
}

func TestDiscoverPartialSourceFailure(t *testing.T) {
	doc := boardconfig.File{
		Sync: boardconfig.SyncSources{
			GitHub: boardconfig.GitHubSync{Orgs: []string{"acme"}},
			GitLab: boardconfig.GitLabSync{Groups: []string{"broken"}},
		},
	}
	lister := fakeLister{
		gh: map[string][]remotegit.RepoRef{
			"acme": {{Host: boardconfig.HostGitHub, Path: "acme/a", Name: "a"}},
		},
		glErr: map[string]error{
			"broken": errors.New("signal: killed"),
		},
	}
	res, err := projectsync.Discover(context.Background(), lister, doc, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Candidates) != 1 || res.Candidates[0].Path != "acme/a" {
		t.Fatalf("candidates=%+v", res.Candidates)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "gitlab group broken") {
		t.Fatalf("warnings=%v", res.Warnings)
	}
}

func TestApplySelectionByRefs(t *testing.T) {
	cands := []projectsync.Candidate{
		{Host: boardconfig.HostGitHub, Path: "acme/a", Name: "a", Index: 1},
		{Host: boardconfig.HostGitHub, Path: "acme/b", Name: "b", Index: 2},
	}
	existing := []boardconfig.Project{{
		ID: "a", Label: "a", Host: boardconfig.HostGitHub, Path: "acme/a", LocalPath: "~/code/a",
	}}
	projects, err := projectsync.ApplySelectionByRefs(cands, []projectsync.RepoRef{
		{Host: boardconfig.HostGitHub, Path: "acme/b"},
		{Host: boardconfig.HostGitHub, Path: "acme/a"},
	}, existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects=%d", len(projects))
	}
	byPath := map[string]boardconfig.Project{}
	for _, p := range projects {
		byPath[p.Path] = p
	}
	if byPath["acme/a"].LocalPath != "~/code/a" {
		t.Fatalf("local_path lost: %+v", byPath["acme/a"])
	}
	if _, err := projectsync.ApplySelectionByRefs(cands, []projectsync.RepoRef{
		{Host: boardconfig.HostGitHub, Path: "acme/missing"},
	}, existing); err == nil {
		t.Fatal("expected unknown ref error")
	}
}

func TestDiscoverRejectsUnknownHostFilter(t *testing.T) {
	doc := boardconfig.File{Sync: boardconfig.SyncSources{GitHub: boardconfig.GitHubSync{Orgs: []string{"acme"}}}}
	if _, err := projectsync.Discover(context.Background(), fakeLister{}, doc, "gitea"); !errors.Is(err, boardconfig.ErrUnknownHost) {
		t.Fatalf("err = %v, want ErrUnknownHost", err)
	}
}

func TestDiscoverAzureProjectFilter(t *testing.T) {
	doc := boardconfig.File{Sync: boardconfig.SyncSources{AzureDevOps: boardconfig.AzureDevOpsSync{
		Orgs: []string{"org"}, Projects: []string{"Keep"},
	}}}
	lister := fakeLister{az: map[string][]remotegit.RepoRef{"org": {
		{Host: boardconfig.HostAzureDevOps, Path: "org/keep/a", Name: "a"},
		{Host: boardconfig.HostAzureDevOps, Path: "org/drop/b", Name: "b"},
		{Host: boardconfig.HostAzureDevOps, Path: "org/short", Name: "c"},
	}}}
	res, err := projectsync.Discover(context.Background(), lister, doc, "azuredevops")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Candidates) != 1 || res.Candidates[0].Path != "org/keep/a" || res.Candidates[0].Index != 1 {
		t.Fatalf("candidates = %+v", res.Candidates)
	}
}

func TestAddProjectChecksPathPerHost(t *testing.T) {
	if _, err := projectsync.AddProject(nil, boardconfig.HostAzureDevOps, "org/repo"); err == nil {
		t.Fatal("azure devops needs org/project/repo")
	}
	if _, err := projectsync.AddProject(nil, boardconfig.HostAzureDevOps, "org/proj/repo"); err != nil {
		t.Fatal(err)
	}
}
