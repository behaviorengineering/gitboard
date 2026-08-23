package forge

import (
	"testing"
	"time"
)

func TestBranchAccumRemoteAndReviews(t *testing.T) {
	a := newBranchAccum()
	// CI-only refs must not invent branches.
	a.setCI("stephen", "success", "", time.Now().UTC().Format(time.RFC3339), "")
	a.addRemote("feat/old", time.Now().UTC().Add(-3*time.Hour).Format(time.RFC3339), "")
	a.addRemote("feat/b", time.Now().UTC().Add(-2*time.Hour).Format(time.RFC3339), "https://example/branch/b")
	a.setCI("feat/b", "success", "https://example/ci", time.Now().UTC().Add(-2*time.Hour).Format(time.RFC3339), "99")
	a.setOpenReview("feat/b", reviewInfo{
		ID:        12,
		URL:       "https://example/pr/12",
		Conflict:  true,
		UpdatedAt: time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
	})
	a.setDefault("main")
	a.addRemote("main", time.Now().UTC().Add(-20*24*time.Hour).Format(time.RFC3339), "")
	a.setCI("main", "success", "", time.Now().UTC().Add(-20*24*time.Hour).Format(time.RFC3339), "1")

	list := a.list()
	if len(list) != 3 {
		t.Fatalf("want default+review+remote, got %d: %+v", len(list), list)
	}
	if !list[0].Default || list[0].Name != "main" {
		t.Fatalf("default first: %+v", list[0])
	}
	if list[0].CIStatus != "success" || !list[0].Stale {
		t.Fatalf("main ci/stale: %+v", list[0])
	}
	if list[1].Name != "feat/b" || !list[1].OpenReview || !list[1].Conflict || list[1].ReviewID != 12 {
		t.Fatalf("feat review: %+v", list[1])
	}
	if list[1].CIURL != "https://example/ci" || list[1].RunID != "99" {
		t.Fatalf("feat ci fields: %+v", list[1])
	}
	if list[2].Name != "feat/old" || list[2].OpenReview {
		t.Fatalf("plain remote: %+v", list[2])
	}
	for _, b := range list {
		if b.Name == "stephen" {
			t.Fatal("CI-only branch should not be listed")
		}
	}
}

func TestGithubHasConflict(t *testing.T) {
	if !githubHasConflict("CONFLICTING", "") {
		t.Fatal("mergeable CONFLICTING")
	}
	if !githubHasConflict("MERGEABLE", "DIRTY") {
		t.Fatal("state DIRTY")
	}
	if githubHasConflict("MERGEABLE", "CLEAN") {
		t.Fatal("clean should not conflict")
	}
}
