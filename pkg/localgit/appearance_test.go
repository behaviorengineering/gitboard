package localgit

import (
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/board"
)

func TestPickPrimaryPrefersExplicitThenStandalone(t *testing.T) {
	disc := Discovery{ByKey: map[string][]Checkout{
		"github:acme/repo": {
			{Path: "/scanned/parent/sub/repo", Role: board.AppearanceSubmodule, Superproject: "/scanned/parent"},
			{Path: "/scanned/repo", Role: board.AppearanceStandalone},
		},
	}}
	checkouts := disc.Appearances("github", "acme/repo")
	c, ok := PickPrimary("", checkouts)
	if !ok || c.Path != "/scanned/repo" {
		t.Fatalf("scan match: %q ok=%v", c.Path, ok)
	}
	c, ok = PickPrimary("/explicit/repo", checkouts)
	if !ok || c.Path != "/explicit/repo" {
		t.Fatalf("explicit: %q ok=%v", c.Path, ok)
	}
}
