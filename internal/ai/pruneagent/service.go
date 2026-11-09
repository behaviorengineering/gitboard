// Package pruneagent investigates local-only / likely-removable branches.
package pruneagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/ai/llm"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/clock"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/strop/agentsession"
)

const (
	kindPruneInvestigate = "prune_investigate"
	sourceLLM            = "llm"
	sourceRules          = "rules"
	roleSystem           = "system"
	defaultBranchName    = "main"

	// Card summaries shared by the rules card and the clamp, so both read
	// the same in the UI.
	summaryIncomplete = "Evidence is incomplete because a git read failed; inspect before deleting."
	summaryDirty      = "Working tree is dirty; do not delete until changes are committed or discarded."
	summaryUnrelated  = "Branch history is unrelated to the default branch; inspect before deleting."
	summaryOnDefault  = "Branch tip matches default (or is already merged); safe to delete the local branch after switching away."
	summaryUnique     = "Unique commits remain on this branch; keep until you merge, cherry-pick, or explicitly drop that work."
	summaryNoUnique   = "No unique commits vs default and tree is clean; safe to delete the local branch after switching away."
	summaryAmbiguous  = "Ambiguous verdict; inspect before deleting."

	// maxCardCommits and maxCardFiles cap the lists sent to the LLM and UI.
	maxCardCommits = 20
	maxCardFiles   = 40
)

// Verdict is the prune card decision.
type Verdict string

// Prune card verdicts.
const (
	VerdictDrop    Verdict = "drop"
	VerdictKeep    Verdict = "keep"
	VerdictAskUser Verdict = "ask_user"
)

const cardSystemPrompt = `You judge whether a local git branch is safe to delete.
Use ONLY the evidence JSON in the user message. Do not invent commits, files, or dirty state.
Verdict must be exactly one of: drop, keep, ask_user.
- drop: clean tree, related histories, and either content_on_default is true (tip trees match default / branch is ancestor) or no unique commits vs default; safe to delete after switching away
- keep: unique commits remain and content_on_default is false (salvageable work not on default)
- ask_user: dirty tree, unrelated histories, or ambiguous

Respond in plain text with sections:
VERDICT: drop|keep|ask_user
SUMMARY: one or two sentences
BULLETS:
- short evidence bullets
COMMAND: shell command only when VERDICT is drop; otherwise leave empty`

// Request is the Investigate API body.
type Request struct {
	ProjectID     string `json:"project_id"`
	Branch        string `json:"branch"`
	WorktreePath  string `json:"worktree_path"`
	DefaultBranch string `json:"default_branch"`
}

// Evidence is structured git/forge facts (no LLM).
type Evidence struct {
	Branch           string   `json:"branch"`
	WorktreePath     string   `json:"worktree_path"`
	DefaultBranch    string   `json:"default_branch"`
	Dirty            bool     `json:"dirty"`
	RelatedHistories bool     `json:"related_histories"`
	ContentOnDefault bool     `json:"content_on_default"`
	UniqueCommits    []string `json:"unique_commits,omitempty"`
	UniqueCommitN    int      `json:"unique_commit_count"`
	UniqueFiles      []string `json:"unique_files,omitempty"`
	UniqueFileN      int      `json:"unique_file_count"`
	AheadOfDefault   int      `json:"ahead_of_default"`
	BehindDefault    int      `json:"behind_default"`
	// Incomplete is true when a git read about unique work (rev-list, log or
	// diff) failed. The card then never recommends drop: unknown unique
	// commits must not read as zero.
	Incomplete bool     `json:"incomplete,omitempty"`
	Notes      []string `json:"notes,omitempty"`
	GatheredAt string   `json:"gathered_at"`
}

// Card is the UI verdict.
type Card struct {
	Verdict  Verdict  `json:"verdict"`
	Summary  string   `json:"summary"`
	Bullets  []string `json:"bullets"`
	Command  string   `json:"command,omitempty"`
	Evidence string   `json:"evidence_ref,omitempty"`
}

// Result is returned to the API/UI.
type Result struct {
	SessionID string   `json:"session_id"`
	Card      Card     `json:"card"`
	Evidence  Evidence `json:"evidence"`
	Model     string   `json:"model,omitempty"`
	Source    string   `json:"source,omitempty"` // llm | rules.
}

// Service runs investigations into agentsession directories.
type Service struct {
	Store     SessionStore
	in        *localgit.Inspector
	llm       *llm.Client
	mutations *localgit.MutationCoordinator
	clock     clock.Clock
}

// Config holds what a prune Service needs at construction time.
type Config struct {
	// AgentsDir is the agentsession root (for example ~/.config/gitboard/agents).
	AgentsDir string
	// Exec runs git. Required.
	Exec cliexec.Exec
	// LLM writes the verdict card. Optional: nil uses the rules-only card.
	LLM *llm.Client
	// Mutations serialises the investigate origin fetch with dashboard pulls
	// and prunes. Required; pass the dashboard's coordinator.
	Mutations *localgit.MutationCoordinator
	// Clock stamps evidence and failure records. Optional: nil uses the wall clock.
	Clock clock.Clock
}

// CreateService opens the session store under AgentsDir and builds the Service.
// Panics if Exec or Mutations is nil (required dependencies at wire-up).
func (cfg Config) CreateService() (*Service, error) {
	store, err := agentsession.New(cfg.AgentsDir)
	if err != nil {
		return nil, err
	}
	return newWithStore(store, cfg), nil
}

// newWithStore builds a Service around store. CreateService uses it for the
// on-disk agentsession store; it is the single place the required
// dependencies are checked.
func newWithStore(store SessionStore, cfg Config) *Service {
	if store == nil {
		panic("pruneagent.CreateService: Store is required")
	}
	if cfg.Exec == nil {
		panic("pruneagent.CreateService: Exec is required")
	}
	if cfg.Mutations == nil {
		panic("pruneagent.CreateService: Mutations is required")
	}
	return &Service{
		Store:     store,
		in:        localgit.NewInspector(cfg.Exec),
		llm:       cfg.LLM,
		mutations: cfg.Mutations,
		clock:     clock.OrWall(cfg.Clock),
	}
}

// Investigate gathers evidence, writes a session directory, and synthesizes a card.
// Errors are gberr.Error: CodeInvalid for request fields, CodeUnavailable
// for the session store and LLM, CodeInternal for missing wiring.
func (s *Service) Investigate(ctx context.Context, req Request) (*Result, error) {
	const op = "pruneagent.Investigate"
	if err := s.ready(op); err != nil {
		return nil, err
	}
	branch := strings.TrimSpace(req.Branch)
	wt := strings.TrimSpace(req.WorktreePath)
	def := strings.TrimSpace(req.DefaultBranch)
	if branch == "" || wt == "" {
		return nil, gberr.New(gberr.CodeInvalid, op, "branch and worktree_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return nil, gberr.Wrap(err, gberr.CodeInvalid, op, "branch is not a valid branch name")
	}
	if def == "" {
		def = defaultBranchName
	}
	if err := localgit.ValidateBranchName(def); err != nil {
		return nil, gberr.Wrap(err, gberr.CodeInvalid, op, "default_branch is not a valid branch name")
	}
	abs, err := localgit.ResolvePath(wt)
	if err != nil {
		return nil, gberr.Wrap(err, gberr.CodeInvalid, op, "worktree_path is not a usable path")
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, gberr.Wrap(err, gberr.CodeInvalid, op, "worktree_path does not exist").With("path", abs)
	}

	meta, err := s.Store.Create(ctx, kindPruneInvestigate, map[string]any{
		"project_id":     strings.TrimSpace(req.ProjectID),
		"branch":         branch,
		"worktree_path":  abs,
		"default_branch": def,
	})
	if err != nil {
		return nil, storeErr(err, op, "create session", "")
	}

	ev, gaps := s.gather(ctx, abs, branch, def)
	if len(gaps) > 0 {
		// The causes (git stderr, paths) stay in the server-side session;
		// Evidence.Notes carries only stable text.
		if err := s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{
			Role:    roleSystem,
			Content: "evidence gaps: " + errors.Join(gaps...).Error(),
		}); err != nil {
			return nil, storeErr(err, op, "append evidence gaps", meta.ID)
		}
	}
	if err := s.Store.SaveJSON(ctx, meta.ID, agentsession.FileEvidence, ev); err != nil {
		return nil, storeErr(err, op, "save evidence", meta.ID)
	}

	userPrompt, err := buildUserPrompt(ev)
	if err != nil {
		return nil, gberr.Wrap(err, gberr.CodeInternal, op, "build prompt failed")
	}
	if err := s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{Role: "user", Content: userPrompt}); err != nil {
		return nil, storeErr(err, op, "append user turn", meta.ID)
	}

	card, model, source, err := s.synthesize(ctx, meta.ID, ev, userPrompt)
	if err != nil {
		return nil, err
	}
	card = clampCard(ev, card)

	if err := s.Store.SaveJSON(ctx, meta.ID, agentsession.FileCard, card); err != nil {
		return nil, storeErr(err, op, "save card", meta.ID)
	}
	if err := s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{
		Role:    "assistant",
		Content: card.Summary,
		Refs:    map[string]any{"card": true, "verdict": card.Verdict, "source": source, "model": model},
	}); err != nil {
		return nil, storeErr(err, op, "append assistant turn", meta.ID)
	}

	return &Result{
		SessionID: meta.ID,
		Card:      card,
		Evidence:  ev,
		Model:     model,
		Source:    source,
	}, nil
}

func (s *Service) synthesize(ctx context.Context, sessionID string, ev Evidence, userPrompt string) (Card, string, string, error) {
	if s == nil || s.llm == nil || !s.llm.Enabled() {
		return synthesizeCard(ev), "", sourceRules, nil
	}
	const op = "pruneagent.synthesize"
	raw, model, err := s.llm.Chat(ctx, cardSystemPrompt, userPrompt)
	if err != nil {
		if persistErr := s.persistLLMFailure(ctx, sessionID, map[string]any{
			"error": err.Error(),
			"at":    s.clock().UTC().Format(time.RFC3339),
		}, agentsession.Turn{
			Role:    roleSystem,
			Content: "llm error, falling back to rules: " + err.Error(),
		}); persistErr != nil {
			return Card{}, "", "", storeErr(errors.Join(
				fmt.Errorf("llm chat: %w", err),
				fmt.Errorf("persist failure: %w", persistErr),
			), op, "record llm failure", sessionID)
		}
		return synthesizeCard(ev), "", sourceRules, nil
	}
	card, ok := parseCard(raw)
	if !ok {
		if persistErr := s.persistLLMFailure(ctx, sessionID, map[string]any{
			"error": "unparseable card",
			"raw":   raw,
			"at":    s.clock().UTC().Format(time.RFC3339),
		}, agentsession.Turn{
			Role:    roleSystem,
			Content: "llm response unparseable, falling back to rules",
			Refs:    map[string]any{"raw": raw},
		}); persistErr != nil {
			return Card{}, model, "", storeErr(persistErr, op, "record unparseable card", sessionID)
		}
		return synthesizeCard(ev), model, sourceRules, nil
	}
	return card, model, sourceLLM, nil
}

func (s *Service) persistLLMFailure(ctx context.Context, sessionID string, failure map[string]any, turn agentsession.Turn) error {
	if err := s.Store.SaveJSON(ctx, sessionID, agentsession.FileFailure, failure); err != nil {
		return fmt.Errorf("save failure: %w", err)
	}
	if err := s.Store.AppendTurn(ctx, sessionID, turn); err != nil {
		return fmt.Errorf("append failure turn: %w", err)
	}
	return nil
}

func buildUserPrompt(ev Evidence) (string, error) {
	raw, err := json.MarshalIndent(ev, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal evidence: %w", err)
	}
	return "Judge this local branch for deletion. Evidence JSON:\n" + string(raw), nil
}

func parseCard(raw string) (Card, bool) {
	sections := llm.ParseSections(raw, "VERDICT", "SUMMARY", "BULLETS", "COMMAND")
	out := Card{
		Verdict: Verdict(strings.ToLower(llm.First(sections["VERDICT"]))),
		Summary: strings.Join(sections["SUMMARY"], " "),
		Command: strings.Join(sections["COMMAND"], " "),
	}
	for _, b := range sections["BULLETS"] {
		out.Bullets = append(out.Bullets, strings.TrimLeft(b, "-*• "))
	}
	switch out.Verdict {
	case VerdictDrop, VerdictKeep, VerdictAskUser:
		return out, out.Summary != ""
	default:
		return out, false
	}
}

func clampCard(ev Evidence, card Card) Card {
	switch card.Verdict {
	case VerdictDrop, VerdictKeep, VerdictAskUser:
	default:
		card.Verdict = VerdictAskUser
		if card.Summary == "" {
			card.Summary = summaryAmbiguous
		}
		card.Command = ""
	}
	if ev.Dirty {
		card.Verdict = VerdictAskUser
		card.Command = ""
		if card.Summary == "" {
			card.Summary = summaryDirty
		}
	}
	if !ev.RelatedHistories {
		card.Verdict = VerdictAskUser
		card.Command = ""
		if card.Summary == "" {
			card.Summary = summaryUnrelated
		}
	}
	if ev.Incomplete {
		card.Verdict = VerdictAskUser
		card.Command = ""
		card.Summary = summaryIncomplete
	}
	if ev.ContentOnDefault && card.Verdict != VerdictAskUser {
		card.Verdict = VerdictDrop
		if card.Summary == "" {
			card.Summary = summaryOnDefault
		}
	}
	if ev.UniqueCommitN > 0 && !ev.ContentOnDefault && card.Verdict == VerdictDrop {
		card.Verdict = VerdictKeep
		card.Command = ""
		if card.Summary == "" {
			card.Summary = summaryUnique
		}
	}
	if card.Verdict == VerdictDrop {
		card.Command = dropCommand(ev)
	} else {
		card.Command = ""
	}
	return card
}

func dropCommand(ev Evidence) string {
	return fmt.Sprintf("git -C %q switch %s && git branch -d %s", ev.WorktreePath, ev.DefaultBranch, ev.Branch)
}

// gather reads git facts for the card. Every git read that fails leaves a
// stable note in Evidence and its cause in gaps; reads about unique work also
// set Evidence.Incomplete so the card fails closed.
func (s *Service) gather(ctx context.Context, dir, branch, def string) (Evidence, []error) {
	ev := Evidence{
		Branch:        branch,
		WorktreePath:  dir,
		DefaultBranch: def,
		GatheredAt:    s.clock().UTC().Format(time.RFC3339),
	}
	var gaps []error

	dirty, err := s.in.IsDirty(ctx, dir)
	ev.Dirty = dirty
	if err != nil {
		// Fail closed: unknown dirty state must not recommend drop.
		ev.Dirty = true
		ev.Notes = append(ev.Notes, "git status failed; dirty state unknown")
		gaps = append(gaps, err)
	}

	// merge-base also fails when there is no common ancestor, so either way
	// the card asks the user.
	ev.RelatedHistories, err = s.in.SharesHistory(ctx, dir, def, branch)
	if err != nil {
		ev.Notes = append(ev.Notes, "histories unrelated to "+def)
		gaps = append(gaps, err)
	}
	if ev.RelatedHistories {
		gaps = append(gaps, s.gatherUniqueWork(ctx, dir, branch, def, &ev)...)
	}

	// Best-effort refresh so ContentOnDefault can prefer origin/<default>.
	if err := s.fetchOrigin(ctx, dir); err != nil {
		ev.Notes = append(ev.Notes, "fetch origin failed; content check used local refs")
		gaps = append(gaps, fmt.Errorf("fetch origin: %w", err))
	}
	ok, reason, err := s.in.ContentOnDefault(ctx, dir, branch, def)
	if err != nil {
		// Fail closed: ContentOnDefault stays false.
		ev.Notes = append(ev.Notes, "content_on_default check failed")
		gaps = append(gaps, fmt.Errorf("content_on_default: %w", err))
	} else {
		ev.ContentOnDefault = ok
		if reason != "" {
			ev.Notes = append(ev.Notes, "content_on_default: "+reason)
		}
	}
	return ev, gaps
}

// gatherUniqueWork fills ahead/behind, unique commits and unique files. Any
// failed read marks ev.Incomplete and returns its cause.
func (s *Service) gatherUniqueWork(ctx context.Context, dir, branch, def string, ev *Evidence) []error {
	var gaps []error
	incomplete := func(note string, err error) {
		ev.Incomplete = true
		ev.Notes = append(ev.Notes, note)
		gaps = append(gaps, err)
	}

	if ahead, behind, err := s.in.AheadBehind(ctx, dir, def, branch); err != nil {
		incomplete("git rev-list failed; ahead and behind unknown", err)
	} else {
		ev.AheadOfDefault, ev.BehindDefault = ahead, behind
	}
	if commits, err := s.in.UniqueCommits(ctx, dir, def, branch); err != nil {
		incomplete("git log failed; unique commits unknown", err)
	} else {
		ev.UniqueCommitN = len(commits)
		ev.UniqueCommits = firstN(commits, maxCardCommits)
	}
	if files, err := s.in.ChangedFiles(ctx, dir, def, branch); err != nil {
		incomplete("git diff failed; unique files unknown", err)
	} else {
		ev.UniqueFileN = len(files)
		ev.UniqueFiles = firstN(files, maxCardFiles)
	}
	return gaps
}

func firstN(lines []string, n int) []string {
	if len(lines) > n {
		return lines[:n]
	}
	return lines
}

func synthesizeCard(ev Evidence) Card {
	bullets := []string{
		fmt.Sprintf("branch %s vs %s", ev.Branch, ev.DefaultBranch),
	}
	if ev.Dirty {
		return Card{
			Verdict: VerdictAskUser,
			Summary: summaryDirty,
			Bullets: append(bullets, "dirty working tree"),
			Command: "",
		}
	}
	if !ev.RelatedHistories {
		return Card{
			Verdict: VerdictAskUser,
			Summary: summaryUnrelated,
			Bullets: append(bullets, "unrelated histories"),
			Command: "",
		}
	}
	if ev.Incomplete {
		return Card{
			Verdict: VerdictAskUser,
			Summary: summaryIncomplete,
			Bullets: append(bullets, "evidence incomplete: a git read failed"),
			Command: "",
		}
	}
	if ev.ContentOnDefault {
		if ev.UniqueCommitN > 0 {
			bullets = append(bullets, fmt.Sprintf("%d unique commit SHA(s) vs %s (content already on default)", ev.UniqueCommitN, ev.DefaultBranch))
		} else {
			bullets = append(bullets, "content already on "+ev.DefaultBranch)
		}
		return Card{
			Verdict: VerdictDrop,
			Summary: summaryOnDefault,
			Bullets: bullets,
			Command: dropCommand(ev),
		}
	}
	if ev.UniqueCommitN > 0 {
		bullets = append(bullets, fmt.Sprintf("%d commit(s) not in %s", ev.UniqueCommitN, ev.DefaultBranch))
		if ev.UniqueFileN > 0 {
			bullets = append(bullets, fmt.Sprintf("%d file(s) differ from %s", ev.UniqueFileN, ev.DefaultBranch))
		}
		return Card{
			Verdict: VerdictKeep,
			Summary: summaryUnique,
			Bullets: bullets,
			Command: "",
		}
	}
	bullets = append(bullets, "no unique commits vs "+ev.DefaultBranch)
	return Card{
		Verdict: VerdictDrop,
		Summary: summaryNoUnique,
		Bullets: bullets,
		Command: dropCommand(ev),
	}
}

// fetchOrigin refreshes origin for dir under the repository's mutation lease,
// so it queues behind dashboard pulls and prunes on the same common git dir.
func (s *Service) fetchOrigin(ctx context.Context, dir string) (err error) {
	common, err := localgit.ResolveCommonDir(ctx, s.in.CommonGitDir, dir)
	if err != nil {
		return err
	}
	lease, err := s.mutations.Acquire(ctx, common)
	if err != nil {
		return err
	}
	defer func() {
		if rerr := lease.Release(); rerr != nil {
			err = errors.Join(err, fmt.Errorf("release mutation lease: %w", rerr))
		}
	}()
	m, err := lease.Mutator(s.in.Run)
	if err != nil {
		return err
	}
	updated, err := m.FetchOriginSmart(ctx, dir, localgit.FetchPolicy{Fresh: true}, nil)
	if err != nil {
		return err
	}
	if updated {
		lease.BumpEpoch()
	}
	return nil
}
