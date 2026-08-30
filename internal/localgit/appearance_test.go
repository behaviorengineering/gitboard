package localgit

import (
	"testing"
)

func TestResolvePathPrefersLocalPath(t *testing.T) {
	disc := Discovery{ByKey: map[string][]Checkout{
		"github:acme/repo": {{Path: "/scanned/repo", Role: RoleStandalone}},
	}}
	p, ok := resolvePath("~/explicit", "github", "acme/repo", disc)
	if !ok || p != "~/explicit" {
		t.Fatalf("got %q ok=%v", p, ok)
	}
	p, ok = resolvePath("", "github", "acme/repo", disc)
	if !ok || p != "/scanned/repo" {
		t.Fatalf("scan match: %q ok=%v", p, ok)
	}
}
