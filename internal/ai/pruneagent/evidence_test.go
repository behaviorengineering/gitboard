package pruneagent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit/remotegittest"
)

var errGitLog = errors.New("git log: fatal: /home/me/repo: unable to read index")

func evidenceService(fk *remotegittest.FakeExec) *Service {
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	return &Service{
		in:        localgit.NewInspector(fk),
		mutations: localgit.NewMutationCoordinator(nil),
		clock:     func() time.Time { return fixed },
	}
}

// When git log fails, unique commits are unknown. That must never read as
// "zero unique commits", which the rules card turns into a drop verdict.
func TestGatherMarksEvidenceIncompleteWhenGitFails(t *testing.T) {
	fk := &remotegittest.FakeExec{
		Responses: map[string][]byte{
			"status --porcelain":            []byte(""),
			"merge-base":                    []byte("abc\n"),
			"rev-list --left-right --count": []byte("0\t1\n"),
			"diff --name-only":              []byte(""),
		},
		Errors: map[string]error{"log --oneline": errGitLog},
	}
	s := evidenceService(fk)
	ev, gaps := s.gather(context.Background(), "/tmp/wt", "feat/x", "main")

	if !ev.Incomplete {
		t.Fatalf("evidence must be incomplete: %+v", ev)
	}
	if ev.UniqueCommitN != 0 {
		t.Fatalf("unique commits must stay unknown, got %d", ev.UniqueCommitN)
	}
	var sawLog bool
	for _, g := range gaps {
		if errors.Is(g, errGitLog) {
			sawLog = true
		}
	}
	if !sawLog {
		t.Fatalf("gaps must keep the git log cause: %v", gaps)
	}
	for _, note := range ev.Notes {
		if strings.Contains(note, "/home/me") || strings.Contains(note, "unable to read index") {
			t.Fatalf("notes are sent to clients and must not carry git stderr: %q", note)
		}
	}

	if card := synthesizeCard(ev); card.Verdict != VerdictAskUser || card.Command != "" {
		t.Fatalf("rules card must ask the user: %+v", card)
	}
	llmDrop := Card{Verdict: VerdictDrop, Summary: "No unique commits; safe to delete.", Command: "git branch -d feat/x"}
	if card := clampCard(ev, llmDrop); card.Verdict != VerdictAskUser || card.Command != "" || card.Summary != summaryIncomplete {
		t.Fatalf("clamp must override an LLM drop: %+v", card)
	}
}

func TestGatherCompleteEvidenceStillDrops(t *testing.T) {
	fk := &remotegittest.FakeExec{
		Responses: map[string][]byte{
			"status --porcelain":            []byte(""),
			"merge-base":                    []byte("abc\n"),
			"rev-list --left-right --count": []byte("0\t0\n"),
			"log --oneline":                 []byte(""),
			"diff --name-only":              []byte(""),
		},
	}
	ev, _ := evidenceService(fk).gather(context.Background(), "/tmp/wt", "feat/x", "main")
	if ev.Incomplete {
		t.Fatalf("all unique-work reads succeeded: %+v", ev)
	}
	if card := synthesizeCard(ev); card.Verdict != VerdictDrop {
		t.Fatalf("clean evidence with no unique commits: %+v", card)
	}
}

func TestSessionErrorsAreTyped(t *testing.T) {
	svc, err := Config{
		AgentsDir: t.TempDir(),
		Exec:      &remotegittest.FakeExec{},
		Mutations: localgit.NewMutationCoordinator(nil),
	}.CreateService()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, err = svc.Session(ctx, "missing-session")
	if !IsSessionNotFound(err) || gberr.CodeOf(err) != gberr.CodeNotFound {
		t.Fatalf("missing session: code=%q err=%v", gberr.CodeOf(err), err)
	}
	_, err = svc.Session(ctx, "../etc")
	if gberr.CodeOf(err) != gberr.CodeInvalid {
		t.Fatalf("bad id: code=%q err=%v", gberr.CodeOf(err), err)
	}
	_, err = svc.FailurePath("missing-session")
	if !IsSessionNotFound(err) {
		t.Fatalf("missing failure dump: %v", err)
	}

	var nilSvc *Service
	if _, err := nilSvc.Investigate(ctx, Request{}); gberr.CodeOf(err) != gberr.CodeInternal {
		t.Fatalf("nil service: %v", err)
	}
	if _, err := (&Service{}).in.IsDirty(ctx, "/tmp"); !errors.Is(err, localgit.ErrInspectorMissing) {
		t.Fatalf("service without Exec: %v", err)
	}
}

func TestInvestigateValidationIsInvalid(t *testing.T) {
	svc, err := Config{
		AgentsDir: t.TempDir(),
		Exec:      &remotegittest.FakeExec{},
		Mutations: localgit.NewMutationCoordinator(nil),
	}.CreateService()
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Investigate(context.Background(), Request{Branch: "feat/x"})
	if gberr.CodeOf(err) != gberr.CodeInvalid {
		t.Fatalf("missing worktree: code=%q err=%v", gberr.CodeOf(err), err)
	}
	_, err = svc.Investigate(context.Background(), Request{Branch: "feat/x", WorktreePath: "/definitely/not/here"})
	e, ok := gberr.As(err)
	if !ok || e.Code != gberr.CodeInvalid || strings.Contains(e.Msg, "/definitely") {
		t.Fatalf("missing path must be invalid without the path in Msg: %+v", e)
	}
}
