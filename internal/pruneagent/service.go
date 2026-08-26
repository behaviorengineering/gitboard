// Package pruneagent investigates local-only / likely-removable branches.
package pruneagent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/strop/agentsession"
)

const kindPruneInvestigate = "prune_investigate"

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
}

// Service runs investigations into agentsession directories.
type Service struct {
	Store *agentsession.Store
	Run   cliexec.Exec
}

// New builds a Service under agentsRoot (e.g. ~/.config/gitboard/agents).
func New(agentsRoot string, run cliexec.Exec) (*Service, error) {
	store, err := agentsession.New(agentsRoot)
	if err != nil {
		return nil, err
	}
	if run == nil {
		run = cliexec.New()
	}
	return &Service{Store: store, Run: run}, nil
}

// Investigate gathers evidence, writes a session directory, and synthesizes a card.
func (s *Service) Investigate(ctx context.Context, req Request) (*Result, error) {
	branch := strings.TrimSpace(req.Branch)
	wt := strings.TrimSpace(req.WorktreePath)
	def := strings.TrimSpace(req.DefaultBranch)
	if branch == "" || wt == "" {
		return nil, fmt.Errorf("branch and worktree_path are required")
	}
	if def == "" {
		def = "main"
	}
	abs, err := filepath.Abs(wt)
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
	card := synthesizeCard(ev)
	if err := s.Store.SaveJSON(ctx, meta.ID, agentsession.FileCard, card); err != nil {
		return nil, err
	}
	_ = s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{
		Role:    "assistant",
		Content: card.Summary,
		Refs:    map[string]any{"card": true, "verdict": card.Verdict},
	})

	return &Result{SessionID: meta.ID, Card: card, Evidence: ev}, nil
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
	cmd := fmt.Sprintf("git -C %q switch %s && git branch -d %s", ev.WorktreePath, ev.DefaultBranch, ev.Branch)
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
		Command: cmd,
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
