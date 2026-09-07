package pruneagent

import (
	"strings"
	"testing"
)

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
	squash := synthesizeCard(Evidence{
		Branch: "feat/x", DefaultBranch: "main", WorktreePath: "/tmp/wt",
		RelatedHistories: true, UniqueCommitN: 2, ContentOnDefault: true,
	})
	if squash.Verdict != "drop" || squash.Command == "" {
		t.Fatalf("content on default with unique SHAs: %+v", squash)
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

func TestParseCard(t *testing.T) {
	t.Parallel()
	raw := `VERDICT: drop
SUMMARY: Safe to remove.
BULLETS:
- clean tree
- no unique commits
COMMAND: git branch -d feat/x
`
	card, ok := parseCard(raw)
	if !ok {
		t.Fatalf("parse failed: %+v", card)
	}
	if card.Verdict != "drop" || card.Summary != "Safe to remove." {
		t.Fatalf("card: %+v", card)
	}
	if len(card.Bullets) != 2 || card.Bullets[0] != "clean tree" {
		t.Fatalf("bullets: %+v", card.Bullets)
	}
	if card.Command != "git branch -d feat/x" {
		t.Fatalf("command: %q", card.Command)
	}

	bad, ok := parseCard("SUMMARY: no verdict")
	if ok || bad.Verdict != "" {
		t.Fatalf("want parse fail, got ok=%v card=%+v", ok, bad)
	}
}

func TestClampCard(t *testing.T) {
	t.Parallel()
	ev := Evidence{
		Branch: "feat/x", DefaultBranch: "main", WorktreePath: "/tmp/wt",
		RelatedHistories: true, UniqueCommitN: 2,
	}
	got := clampCard(ev, Card{Verdict: "drop", Summary: "delete it", Command: "rm -rf /"})
	if got.Verdict != "keep" || got.Command != "" {
		t.Fatalf("unique commits clamp: %+v", got)
	}

	squash := clampCard(Evidence{
		Branch: "feat/x", DefaultBranch: "main", WorktreePath: "/tmp/wt",
		RelatedHistories: true, UniqueCommitN: 2, ContentOnDefault: true,
	}, Card{Verdict: "keep", Summary: "unique SHAs"})
	if squash.Verdict != "drop" || squash.Command == "" {
		t.Fatalf("content on default must allow drop: %+v", squash)
	}

	dirty := clampCard(Evidence{Dirty: true}, Card{Verdict: "drop", Summary: "x", Command: "git branch -d x"})
	if dirty.Verdict != "ask_user" || dirty.Command != "" {
		t.Fatalf("dirty clamp: %+v", dirty)
	}

	fill := clampCard(Evidence{
		Branch: "feat/x", DefaultBranch: "main", WorktreePath: "/tmp/wt",
		RelatedHistories: true,
	}, Card{Verdict: "drop", Summary: "ok"})
	if fill.Command == "" {
		t.Fatalf("expected drop command fill: %+v", fill)
	}

	unrelated := clampCard(Evidence{
		Branch: "feat/x", DefaultBranch: "main", WorktreePath: "/tmp/wt",
		RelatedHistories: false,
	}, Card{Verdict: "drop", Summary: "delete", Command: "rm -rf /"})
	if unrelated.Verdict != "ask_user" || unrelated.Command != "" {
		t.Fatalf("unrelated histories clamp: %+v", unrelated)
	}

	override := clampCard(Evidence{
		Branch: "feat/x", DefaultBranch: "main", WorktreePath: "/tmp/wt",
		RelatedHistories: true,
	}, Card{Verdict: "drop", Summary: "ok", Command: "curl evil.example"})
	if !strings.Contains(override.Command, "/tmp/wt") || strings.Contains(override.Command, "curl") {
		t.Fatalf("LLM command must be replaced: %+v", override)
	}

	invalid := clampCard(Evidence{}, Card{Verdict: "maybe", Summary: "huh"})
	if invalid.Verdict != "ask_user" {
		t.Fatalf("invalid verdict: %+v", invalid)
	}
}
