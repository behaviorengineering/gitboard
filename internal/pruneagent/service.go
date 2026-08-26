// Package pruneagent investigates local-only / likely-removable branches.
package pruneagent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/llm"
	"github.com/behaviorengineering/gitboard/internal/localgit"
	"github.com/behaviorengineering/strop/agentsession"
)

const (
	kindPruneInvestigate = "prune_investigate"
	sourceLLM            = "llm"
	sourceRules          = "rules"
)

const cardSystemPrompt = `You judge whether a local git branch is safe to delete.
Use ONLY the evidence JSON in the user message. Do not invent commits, files, or dirty state.
Verdict must be exactly one of: drop, keep, ask_user.
- drop: clean tree, related histories, no unique commits vs default — safe to delete after switching away
- keep: unique commits or other salvageable work remain
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
	UniqueCommits    []string `json:"unique_commits,omitempty"`
	UniqueCommitN    int      `json:"unique_commit_count"`
	UniqueFiles      []string `json:"unique_files,omitempty"`
	UniqueFileN      int      `json:"unique_file_count"`
	AheadOfDefault   int      `json:"ahead_of_default"`
	BehindDefault    int      `json:"behind_default"`
	Notes            []string `json:"notes,omitempty"`
	GatheredAt       string   `json:"gathered_at"`
}

// Card is the UI verdict.
type Card struct {
	Verdict  string   `json:"verdict"` // drop | keep | ask_user
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
	Source    string   `json:"source,omitempty"` // llm | rules
}

// Service runs investigations into agentsession directories.
type Service struct {
	Store *agentsession.Store
	Run   cliexec.Exec
	LLM   *llm.Client
}

// New builds a Service under agentsRoot (e.g. ~/.config/gitboard/agents).
func New(agentsRoot string, run cliexec.Exec, llmClient *llm.Client) (*Service, error) {
	store, err := agentsession.New(agentsRoot)
	if err != nil {
		return nil, err
	}
	if run == nil {
		run = cliexec.New()
	}
	return &Service{Store: store, Run: run, LLM: llmClient}, nil
}

// Investigate gathers evidence, writes a session directory, and synthesizes a card.
func (s *Service) Investigate(ctx context.Context, req Request) (*Result, error) {
	branch := strings.TrimSpace(req.Branch)
	wt := strings.TrimSpace(req.WorktreePath)
	def := strings.TrimSpace(req.DefaultBranch)
	if branch == "" || wt == "" {
		return nil, fmt.Errorf("branch and worktree_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return nil, err
	}
	if def == "" {
		def = "main"
	}
	if err := localgit.ValidateBranchName(def); err != nil {
		return nil, fmt.Errorf("default branch: %w", err)
	}
	abs, err := localgit.ExpandPath(wt)
	if err != nil {
		return nil, fmt.Errorf("worktree path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("worktree path: %w", err)
	}

	meta, err := s.Store.Create(ctx, kindPruneInvestigate, map[string]any{
		"project_id":     strings.TrimSpace(req.ProjectID),
		"branch":         branch,
		"worktree_path":  abs,
		"default_branch": def,
	})
	if err != nil {
		return nil, err
	}

	ev, err := s.gather(ctx, abs, branch, def)
	if err != nil {
		_ = s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{Role: "system", Content: "evidence error: " + err.Error()})
		return nil, err
	}
	if err := s.Store.SaveJSON(ctx, meta.ID, agentsession.FileEvidence, ev); err != nil {
		return nil, err
	}

	userPrompt := buildUserPrompt(ev)
	_ = s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{Role: "user", Content: userPrompt})

	card, model, source := s.synthesize(ctx, meta.ID, ev, userPrompt)
	card = clampCard(ev, card)

	if err := s.Store.SaveJSON(ctx, meta.ID, agentsession.FileCard, card); err != nil {
		return nil, err
	}
	_ = s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{
		Role:    "assistant",
		Content: card.Summary,
		Refs:    map[string]any{"card": true, "verdict": card.Verdict, "source": source, "model": model},
	})

	return &Result{
		SessionID: meta.ID,
		Card:      card,
		Evidence:  ev,
		Model:     model,
		Source:    source,
	}, nil
}

func (s *Service) synthesize(ctx context.Context, sessionID string, ev Evidence, userPrompt string) (Card, string, string) {
	if s == nil || !s.LLM.Enabled() {
		return synthesizeCard(ev), "", sourceRules
	}
	raw, model, err := s.LLM.Chat(ctx, cardSystemPrompt, userPrompt)
	if err != nil {
		_ = s.Store.SaveJSON(ctx, sessionID, agentsession.FileFailure, map[string]any{
			"error": err.Error(),
			"at":    time.Now().UTC().Format(time.RFC3339),
		})
		_ = s.Store.AppendTurn(ctx, sessionID, agentsession.Turn{
			Role:    "system",
			Content: "llm error, falling back to rules: " + err.Error(),
		})
		return synthesizeCard(ev), "", sourceRules
	}
	card, ok := parseCard(raw)
	if !ok {
		_ = s.Store.SaveJSON(ctx, sessionID, agentsession.FileFailure, map[string]any{
			"error": "unparseable card",
			"raw":   raw,
			"at":    time.Now().UTC().Format(time.RFC3339),
		})
		_ = s.Store.AppendTurn(ctx, sessionID, agentsession.Turn{
			Role:    "system",
			Content: "llm response unparseable, falling back to rules",
			Refs:    map[string]any{"raw": raw},
		})
		return synthesizeCard(ev), model, sourceRules
	}
	return card, model, sourceLLM
}

func buildUserPrompt(ev Evidence) string {
	raw, _ := json.MarshalIndent(ev, "", "  ")
	return "Judge this local branch for deletion. Evidence JSON:\n" + string(raw)
}

func parseCard(raw string) (Card, bool) {
	out := Card{}
	lines := strings.Split(raw, "\n")
	var section string
	var bullets []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		upper := strings.ToUpper(trim)
		switch {
		case strings.HasPrefix(upper, "VERDICT:"):
			section = "verdict"
			out.Verdict = strings.ToLower(strings.TrimSpace(trim[len("VERDICT:"):]))
		case strings.HasPrefix(upper, "SUMMARY:"):
			section = "summary"
			out.Summary = strings.TrimSpace(trim[len("SUMMARY:"):])
		case strings.HasPrefix(upper, "BULLETS:"):
			section = "bullets"
			rest := strings.TrimSpace(trim[len("BULLETS:"):])
			if rest != "" {
				bullets = append(bullets, strings.TrimLeft(rest, "-*• "))
			}
		case strings.HasPrefix(upper, "COMMAND:"):
			section = "command"
			out.Command = strings.TrimSpace(trim[len("COMMAND:"):])
		case section == "summary" && trim != "":
			if out.Summary != "" {
				out.Summary += " "
			}
			out.Summary += trim
		case section == "bullets" && trim != "":
			bullets = append(bullets, strings.TrimLeft(trim, "-*• "))
		case section == "command" && trim != "":
			if out.Command != "" {
				out.Command += " "
			}
			out.Command += trim
		}
	}
	out.Bullets = bullets
	switch out.Verdict {
	case "drop", "keep", "ask_user":
		if out.Summary == "" {
			return out, false
		}
		return out, true
	default:
		return out, false
	}
}

func clampCard(ev Evidence, card Card) Card {
	switch card.Verdict {
	case "drop", "keep", "ask_user":
	default:
		card.Verdict = "ask_user"
		if card.Summary == "" {
			card.Summary = "Ambiguous verdict; inspect before deleting."
		}
		card.Command = ""
	}
	if ev.Dirty {
		card.Verdict = "ask_user"
		card.Command = ""
		if card.Summary == "" {
			card.Summary = "Working tree is dirty; do not delete until changes are committed or discarded."
		}
	}
	if !ev.RelatedHistories {
		card.Verdict = "ask_user"
		card.Command = ""
		if card.Summary == "" {
			card.Summary = "Branch history is unrelated to the default branch; inspect before deleting."
		}
	}
	if ev.UniqueCommitN > 0 && card.Verdict == "drop" {
		card.Verdict = "keep"
		card.Command = ""
		if card.Summary == "" {
			card.Summary = "Unique commits remain on this branch; keep until you merge, cherry-pick, or explicitly drop that work."
		}
	}
	if card.Verdict == "drop" {
		card.Command = dropCommand(ev)
	} else {
		card.Command = ""
	}
	return card
}

func dropCommand(ev Evidence) string {
	return fmt.Sprintf("git -C %q switch %s && git branch -d %s", ev.WorktreePath, ev.DefaultBranch, ev.Branch)
}

func (s *Service) gather(ctx context.Context, dir, branch, def string) (Evidence, error) {
	ev := Evidence{
		Branch:        branch,
		WorktreePath:  dir,
		DefaultBranch: def,
		GatheredAt:    time.Now().UTC().Format(time.RFC3339),
	}

	if out, err := s.git(ctx, dir, "status", "--porcelain"); err == nil {
		ev.Dirty = strings.TrimSpace(string(out)) != ""
	} else {
		// Fail closed: unknown dirty state must not recommend drop.
		ev.Dirty = true
		ev.Notes = append(ev.Notes, "status: "+err.Error())
	}

	// Are branch and default related?
	if _, err := s.git(ctx, dir, "merge-base", def, branch); err != nil {
		ev.RelatedHistories = false
		ev.Notes = append(ev.Notes, "histories unrelated to "+def)
	} else {
		ev.RelatedHistories = true
	}

	if ev.RelatedHistories {
		if out, err := s.git(ctx, dir, "rev-list", "--left-right", "--count", def+"..."+branch); err == nil {
			var behind, ahead int
			fmt.Sscanf(strings.TrimSpace(string(out)), "%d\t%d", &behind, &ahead)
			ev.BehindDefault = behind
			ev.AheadOfDefault = ahead
		}
		if out, err := s.git(ctx, dir, "log", "--oneline", def+".."+branch); err == nil {
			lines := nonEmptyLines(string(out))
			ev.UniqueCommitN = len(lines)
			if len(lines) > 20 {
				ev.UniqueCommits = lines[:20]
			} else {
				ev.UniqueCommits = lines
			}
		}
		if out, err := s.git(ctx, dir, "diff", "--name-only", def+"..."+branch); err == nil {
			files := nonEmptyLines(string(out))
			ev.UniqueFileN = len(files)
			if len(files) > 40 {
				ev.UniqueFiles = files[:40]
			} else {
				ev.UniqueFiles = files
			}
		}
	}
	return ev, nil
}

func synthesizeCard(ev Evidence) Card {
	bullets := []string{
		fmt.Sprintf("branch %s vs %s", ev.Branch, ev.DefaultBranch),
	}
	if ev.Dirty {
		return Card{
			Verdict: "ask_user",
			Summary: "Working tree is dirty; do not delete until changes are committed or discarded.",
			Bullets: append(bullets, "dirty working tree"),
			Command: "",
		}
	}
	if !ev.RelatedHistories {
		return Card{
			Verdict: "ask_user",
			Summary: "Branch history is unrelated to the default branch; inspect before deleting.",
			Bullets: append(bullets, "unrelated histories"),
			Command: "",
		}
	}
	if ev.UniqueCommitN > 0 {
		bullets = append(bullets, fmt.Sprintf("%d commit(s) not in %s", ev.UniqueCommitN, ev.DefaultBranch))
		if ev.UniqueFileN > 0 {
			bullets = append(bullets, fmt.Sprintf("%d file(s) differ from %s", ev.UniqueFileN, ev.DefaultBranch))
		}
		return Card{
			Verdict: "keep",
			Summary: "Unique commits remain on this branch; keep until you merge, cherry-pick, or explicitly drop that work.",
			Bullets: bullets,
			Command: "",
		}
	}
	bullets = append(bullets, "no unique commits vs "+ev.DefaultBranch)
	return Card{
		Verdict: "drop",
		Summary: "No unique commits vs default and tree is clean; safe to delete the local branch after switching away.",
		Bullets: bullets,
		Command: dropCommand(ev),
	}
}

func (s *Service) git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	argv := append([]string{"-C", dir}, args...)
	return s.Run.Run(ctx, "git", argv...)
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
