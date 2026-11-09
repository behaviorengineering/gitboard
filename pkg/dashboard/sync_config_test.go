package dashboard

import (
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
)

func TestAddTrackedProjectValidatesHost(t *testing.T) {
	c := NewCommands(newTestService(&fakeExec{}, nil))
	doc := boardconfig.File{}
	_, err := c.AddTrackedProject(doc, "gitea", "acme/app")
	if err == nil || !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("want bad request, got %v", err)
	}
}

func TestSetViewsRequiresNonEmpty(t *testing.T) {
	c := NewCommands(newTestService(&fakeExec{}, nil))
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{ID: "a", Label: "A", Host: boardconfig.HostGitHub, Path: "o/a"}},
	}
	_, err := c.SetViews(doc, nil)
	if err == nil || !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("want bad request, got %v", err)
	}
}

func TestSetProjectLocalPath(t *testing.T) {
	c := NewCommands(newTestService(&fakeExec{}, nil))
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{ID: "a", Label: "A", Host: boardconfig.HostGitHub, Path: "o/a"}},
	}
	doc, err := c.SetProjectLocalPath(doc, "a", "~/code/a")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Projects[0].LocalPath != "~/code/a" {
		t.Fatalf("path=%q", doc.Projects[0].LocalPath)
	}
	doc, err = c.SetProjectLocalPath(doc, "a", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Projects[0].LocalPath != "" {
		t.Fatalf("want cleared, got %q", doc.Projects[0].LocalPath)
	}
	_, err = c.SetProjectLocalPath(doc, "missing", "/tmp/x")
	if err == nil || !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("want bad request, got %v", err)
	}
}

func TestSetLocalRootsDedupes(t *testing.T) {
	c := NewCommands(newTestService(&fakeExec{}, nil))
	doc, err := c.SetLocalRoots(boardconfig.File{}, []string{" ~/code ", "~/code", "", "~/other"})
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Local.Roots) != 2 || doc.Local.Roots[0] != "~/code" || doc.Local.Roots[1] != "~/other" {
		t.Fatalf("%+v", doc.Local.Roots)
	}
	doc, err = c.SetLocalRoots(doc, nil)
	if err != nil || len(doc.Local.Roots) != 0 {
		t.Fatalf("empty roots: %v %+v", err, doc.Local.Roots)
	}
}

// A sync source that fails during rediscovery leaves its repositories out of
// the candidates. The selection must fail as unavailable (retry later), not
// as invalid operator input.
func TestApplySyncSelectionReportsDiscoveryWarnings(t *testing.T) {
	c := NewCommands(newTestService(&fakeExec{}, nil)) // every forge call fails.
	doc := boardconfig.File{}
	doc.Sync.GitHub.Orgs = []string{"acme"}
	_, err := c.ApplySyncSelection(t.Context(), doc, "", []projectsync.RepoRef{{Host: boardconfig.HostGitHub, Path: "acme/app"}})
	if !gberr.HasCode(err, gberr.CodeUnavailable) {
		t.Fatalf("want unavailable, got code=%q err=%v", gberr.CodeOf(err), err)
	}
	if got := gberr.FieldsOf(err)["warnings"]; got == "" {
		t.Fatalf("warnings must stay in the error fields for the server log: %v", err)
	}
}

func TestDiscoverCandidatesRejectsUnknownHostFilter(t *testing.T) {
	c := NewCommands(newTestService(&fakeExec{}, nil))
	doc := boardconfig.File{}
	doc.Sync.GitHub.Orgs = []string{"acme"}
	_, err := c.DiscoverCandidates(t.Context(), doc, "gitea")
	if !gberr.HasCode(err, gberr.CodeInvalid) {
		t.Fatalf("want invalid, got code=%q err=%v", gberr.CodeOf(err), err)
	}
}
