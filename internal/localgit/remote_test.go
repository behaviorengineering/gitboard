package localgit_test

import (
	"testing"

	"github.com/behaviorengineering/gitboard/internal/localgit"
)

func TestParseRemoteURL(t *testing.T) {
	cases := []struct {
		in   string
		host string
		path string
		ok   bool
	}{
		{"https://github.com/acme/repo.git", "github", "acme/repo", true},
		{"git@github.com:acme/repo.git", "github", "acme/repo", true},
		{"https://gitlab.com/group/sub/repo.git", "gitlab", "group/sub/repo", true},
		{"git@gitlab.com:group/repo.git", "gitlab", "group/repo", true},
		{"ssh://git@github.com/acme/repo.git", "github", "acme/repo", true},
		{"https://example.com/acme/repo.git", "", "", false},
		{"", "", "", false},
	}
	for _, tc := range cases {
		got, ok := localgit.ParseRemoteURL(tc.in)
		if ok != tc.ok {
			t.Fatalf("%q ok=%v want %v", tc.in, ok, tc.ok)
		}
		if !tc.ok {
			continue
		}
		if got.Host != tc.host || got.Path != tc.path {
			t.Fatalf("%q => %+v want %s %s", tc.in, got, tc.host, tc.path)
		}
	}
}

func TestExpandPathHome(t *testing.T) {
	got, err := localgit.ExpandPath("~/tmp-gitboard-test")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" || got[0] != '/' {
		t.Fatalf("expected absolute path, got %q", got)
	}
}
