package main

import "testing"

func TestResolveVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		ldflag        string
		moduleVersion string
		want          string
	}{
		{name: "ldflag wins", ldflag: "v1.2.3", moduleVersion: "v9.9.9", want: "v1.2.3"},
		{name: "go install tag", ldflag: "dev", moduleVersion: "v0.1.0", want: "v0.1.0"},
		{name: "local devel stays dev", ldflag: "dev", moduleVersion: "(devel)", want: "dev"},
		{name: "empty module stays dev", ldflag: "dev", moduleVersion: "", want: "dev"},
		{name: "whitespace ldflag ignored", ldflag: "  ", moduleVersion: "v0.2.0", want: "v0.2.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveVersion(tc.ldflag, tc.moduleVersion)
			if got != tc.want {
				t.Fatalf("resolveVersion(%q, %q) = %q, want %q", tc.ldflag, tc.moduleVersion, got, tc.want)
			}
		})
	}
}
