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

func TestViewSummariesImplicit(t *testing.T) {
	doc := config.File{
		Projects: []config.Project{{ID: "a", Label: "A", Host: config.HostGitHub, Path: "o/a"}},
	}
	views := ViewSummaries(doc)
	if len(views) != 1 || !views[0].Implicit || views[0].Count != 1 {
		t.Fatalf("%+v", views)
	}
}
