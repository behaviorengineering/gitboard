package boardconfig

import "testing"

func TestEffectiveUpstreamSeconds(t *testing.T) {
	doc := File{}
	if got := doc.EffectiveHeadsSeconds(); got != defaultHeadsSeconds {
		t.Fatalf("heads default=%d", got)
	}
	if got := doc.EffectiveMergedSeconds(); got != defaultMergedSeconds {
		t.Fatalf("merged default=%d", got)
	}
	if got := doc.EffectiveFetchSeconds(); got != defaultFetchSeconds {
		t.Fatalf("fetch default=%d", got)
	}
	if got := doc.EffectiveScanSeconds(); got != defaultScanSeconds {
		t.Fatalf("scan default=%d", got)
	}
	zero := 0
	doc.Upstream.HeadsSeconds = &zero
	doc.Upstream.MergedSeconds = &zero
	doc.Local.FetchSeconds = &zero
	doc.Local.ScanSeconds = &zero
	if got := doc.EffectiveHeadsSeconds(); got != 0 {
		t.Fatalf("heads explicit zero=%d", got)
	}
	if got := doc.EffectiveMergedSeconds(); got != 0 {
		t.Fatalf("merged explicit zero=%d", got)
	}
	if got := doc.EffectiveFetchSeconds(); got != 0 {
		t.Fatalf("fetch explicit zero=%d", got)
	}
	if got := doc.EffectiveScanSeconds(); got != 0 {
		t.Fatalf("scan explicit zero=%d", got)
	}
	neg := -5
	doc.Upstream.HeadsSeconds = &neg
	doc.Local.FetchSeconds = &neg
	doc.Local.ScanSeconds = &neg
	if got := doc.EffectiveHeadsSeconds(); got != defaultHeadsSeconds {
		t.Fatalf("heads negative=%d", got)
	}
	if got := doc.EffectiveFetchSeconds(); got != defaultFetchSeconds {
		t.Fatalf("fetch negative=%d", got)
	}
	if got := doc.EffectiveScanSeconds(); got != defaultScanSeconds {
		t.Fatalf("scan negative=%d", got)
	}
}

func TestEffectiveViewsImplicitDefault(t *testing.T) {
	doc := File{
		Projects: []Project{
			{ID: "a", Label: "A", Host: HostGitHub, Path: "o/a"},
			{ID: "b", Label: "B", Host: HostGitHub, Path: "o/b"},
		},
	}
	views := doc.EffectiveViews()
	if len(views) != 1 || views[0].ID != DefaultViewID {
		t.Fatalf("views: %+v", views)
	}
	if len(views[0].Projects) != 2 {
		t.Fatalf("membership: %+v", views[0].Projects)
	}
}

func TestNormalizeRejectsIgnoredUISettings(t *testing.T) {
	bad := File{UI: UI{HideBranches: []string{"feat/*", "[unclosed"}}}
	if err := bad.Normalize(); err == nil {
		t.Fatal("a malformed hide_branches pattern must be an error, not a no-op")
	}
	neg := -5
	if err := (&File{UI: UI{PollSeconds: &neg}}).Normalize(); err == nil {
		t.Fatal("a negative poll_seconds must be an error")
	}
	ok := File{UI: UI{HideBranches: []string{"dependabot/*"}}}
	if err := ok.Normalize(); err != nil {
		t.Fatalf("valid pattern: %v", err)
	}
}
