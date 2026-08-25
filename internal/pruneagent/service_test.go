package pruneagent

import "testing"

func TestSynthesizeCard(t *testing.T) {
	t.Parallel()
	drop := synthesizeCard(Evidence{
		Branch: "feat/x", DefaultBranch: "main", WorktreePath: "/tmp/wt",
		RelatedHistories: true,
	})
	if drop.Verdict != "drop" || drop.Command == "" {
		t.Fatalf("drop: %+v", drop)
	}
	keep := synthesizeCard(Evidence{
		Branch: "feat/x", DefaultBranch: "main", RelatedHistories: true, UniqueCommitN: 3,
	})
	if keep.Verdict != "keep" {
		t.Fatalf("keep: %+v", keep)
	}
	dirty := synthesizeCard(Evidence{Branch: "feat/x", Dirty: true, RelatedHistories: true})
	if dirty.Verdict != "ask_user" {
		t.Fatalf("dirty: %+v", dirty)
	}
	unrelated := synthesizeCard(Evidence{Branch: "feat/x", RelatedHistories: false})
	if unrelated.Verdict != "ask_user" {
		t.Fatalf("unrelated: %+v", unrelated)
	}
}
