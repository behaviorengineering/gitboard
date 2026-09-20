package syncproj_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/remotegit"
	"github.com/behaviorengineering/gitboard/internal/syncproj"
)

type fakeLister struct {
	gh    map[string][]remotegit.RepoRef
	gl    map[string][]remotegit.RepoRef
	ghErr map[string]error
	glErr map[string]error
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
		gh: map[string][]remotegit.RepoRef{
			"acme": {
				{Host: config.HostGitHub, Path: "acme/alpha", Name: "alpha"},
				{Host: config.HostGitHub, Path: "acme/beta", Name: "beta"},
			},
		},
		gl: map[string][]remotegit.RepoRef{
			"acme": {
				{Host: config.HostGitLab, Path: "acme/gamma", Name: "gamma"},
			},
		},
	}
	res, err := syncproj.Discover(context.Background(), lister, doc, "")
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
		gh: map[string][]remotegit.RepoRef{
			"acme": {{Host: config.HostGitHub, Path: "acme/a", Name: "a"}},
		},
		gl: map[string][]remotegit.RepoRef{
			"acme": {{Host: config.HostGitLab, Path: "acme/b", Name: "b"}},
		},
	}
	res, err := syncproj.Discover(context.Background(), lister, doc, "github")
	if err != nil {
		t.Fatal(err)
	}
	cands := res.Candidates
	if len(cands) != 1 || cands[0].Host != config.HostGitHub {
		t.Fatalf("%+v", cands)
	}
}

func TestDiscoverPartialSourceFailure(t *testing.T) {
	doc := config.File{
		Sync: config.SyncSources{
			GitHub: config.GitHubSync{Orgs: []string{"acme"}},
			GitLab: config.GitLabSync{Groups: []string{"broken"}},
		},
	}
	lister := fakeLister{
		gh: map[string][]remotegit.RepoRef{
			"acme": {{Host: config.HostGitHub, Path: "acme/a", Name: "a"}},
		},
		glErr: map[string]error{
			"broken": errors.New("signal: killed"),
		},
	}
	res, err := syncproj.Discover(context.Background(), lister, doc, "")
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
	cands := []syncproj.Candidate{
		{Host: config.HostGitHub, Path: "acme/a", Name: "a", Index: 1},
		{Host: config.HostGitHub, Path: "acme/b", Name: "b", Index: 2},
	}
	existing := []config.Project{{
		ID: "a", Label: "a", Host: config.HostGitHub, Path: "acme/a", LocalPath: "~/code/a",
	}}
	projects, err := syncproj.ApplySelectionByRefs(cands, []syncproj.RepoRef{
		{Host: config.HostGitHub, Path: "acme/b"},
		{Host: config.HostGitHub, Path: "acme/a"},
	}, existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects=%d", len(projects))
	}
	byPath := map[string]config.Project{}
	for _, p := range projects {
		byPath[p.Path] = p
	}
	if byPath["acme/a"].LocalPath != "~/code/a" {
		t.Fatalf("local_path lost: %+v", byPath["acme/a"])
	}
	if _, err := syncproj.ApplySelectionByRefs(cands, []syncproj.RepoRef{
		{Host: config.HostGitHub, Path: "acme/missing"},
	}, existing); err == nil {
		t.Fatal("expected unknown ref error")
	}
}
