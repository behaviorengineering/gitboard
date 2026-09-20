package dashboard

import (
	"testing"

	"github.com/behaviorengineering/gitboard/internal/config"
)

func TestAddTrackedProjectValidatesHost(t *testing.T) {
	c := NewCommands(&Service{})
	doc := config.File{}
	_, err := c.AddTrackedProject(doc, "bitbucket", "acme/app")
	if err == nil || !IsBadRequest(err) {
		t.Fatalf("want bad request, got %v", err)
	}
}

func TestSetViewsRequiresNonEmpty(t *testing.T) {
	c := NewCommands(&Service{})
	doc := config.File{
		Projects: []config.Project{{ID: "a", Label: "A", Host: config.HostGitHub, Path: "o/a"}},
	}
	_, err := c.SetViews(doc, nil)
	if err == nil || !IsBadRequest(err) {
		t.Fatalf("want bad request, got %v", err)
	}
}

func TestSetProjectLocalPath(t *testing.T) {
	c := NewCommands(&Service{})
	doc := config.File{
		Projects: []config.Project{{ID: "a", Label: "A", Host: config.HostGitHub, Path: "o/a"}},
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
	if err == nil || !IsBadRequest(err) {
		t.Fatalf("want bad request, got %v", err)
	}
}

func TestSetLocalRootsDedupes(t *testing.T) {
	c := NewCommands(&Service{})
	doc, err := c.SetLocalRoots(config.File{}, []string{" ~/code ", "~/code", "", "~/other"})
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
