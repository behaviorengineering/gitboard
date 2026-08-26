package triage

import (
	"context"
	"fmt"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/llm"
)

// Request is the triage input from the dashboard.
type Request struct {
	ProjectID string `json:"project_id"`
	RunID     string `json:"run_id,omitempty"`
	JobID     string `json:"job_id,omitempty"`
	JobName   string `json:"job_name,omitempty"`
	Log       string `json:"log,omitempty"`
}

// Response is the model analysis.
type Response struct {
	Summary     string   `json:"summary"`
	RootCause   string   `json:"root_cause"`
	FixSteps    []string `json:"fix_steps"`
	Confidence  string   `json:"confidence"`
	RawAnswer   string   `json:"raw_answer,omitempty"`
	Model       string   `json:"model,omitempty"`
	Unavailable string   `json:"unavailable,omitempty"`
}

// Analyzer calls an OpenAI-compatible chat endpoint via llm.Client.
type Analyzer struct {
	LLM *llm.Client
}

// New builds an analyzer from effective LLM settings (file + env overlay).
func New(cfg config.LLM) *Analyzer {
	return &Analyzer{LLM: llm.New(cfg)}
}

func (a *Analyzer) Enabled() bool {
	return a != nil && a.LLM.Enabled()
}

// Analyze sends logs to the configured model.
func (a *Analyzer) Analyze(ctx context.Context, req Request) (Response, error) {
	if !a.Enabled() {
		return Response{
			Unavailable: "set llm.base_url in config.yaml (or GITBOARD_LLM_BASE_URL / POLYPUS_BASE_URL) for AI triage",
		}, nil
	}
	logText := strings.TrimSpace(req.Log)
	if logText == "" {
		return Response{}, fmt.Errorf("empty log excerpt")
	}
	prompt := buildPrompt(req, logText)
	raw, model, err := a.LLM.Chat(ctx,
		"You triage CI failures for a Go monorepo. Do not invent file paths or line numbers absent from the log.",
		prompt,
	)
	if err != nil {
		return Response{}, err
	}
	return parseAnswer(raw, model), nil
}

func buildPrompt(req Request, logText string) string {
	var b strings.Builder
	b.WriteString("Analyze this CI failure. Use ONLY the log excerpt below.\n")
	if req.JobName != "" {
		b.WriteString("Job: ")
		b.WriteString(req.JobName)
		b.WriteByte('\n')
	}
	b.WriteString("\nRespond in plain text with sections:\n")
	b.WriteString("SUMMARY: one sentence\n")
	b.WriteString("ROOT_CAUSE: short paragraph\n")
	b.WriteString("FIX_STEPS: numbered list\n")
	b.WriteString("CONFIDENCE: high|medium|low\n\n")
	b.WriteString("LOG:\n")
	b.WriteString(logText)
	return b.String()
}

func parseAnswer(raw, model string) Response {
	out := Response{RawAnswer: raw, Model: model}
	lines := strings.Split(raw, "\n")
	var section string
	var steps []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(strings.ToUpper(trim), "SUMMARY:"):
			section = "summary"
			out.Summary = strings.TrimSpace(trim[len("SUMMARY:"):])
		case strings.HasPrefix(strings.ToUpper(trim), "ROOT_CAUSE:"):
			section = "root"
			out.RootCause = strings.TrimSpace(trim[len("ROOT_CAUSE:"):])
		case strings.HasPrefix(strings.ToUpper(trim), "FIX_STEPS:"):
			section = "fix"
		case strings.HasPrefix(strings.ToUpper(trim), "CONFIDENCE:"):
			section = "confidence"
			out.Confidence = strings.TrimSpace(trim[len("CONFIDENCE:"):])
		case section == "summary" && out.Summary != "" && trim != "":
			out.Summary += " " + trim
		case section == "root" && trim != "":
			if out.RootCause != "" {
				out.RootCause += " "
			}
			out.RootCause += trim
		case section == "fix" && trim != "":
			step := strings.TrimLeft(trim, "0123456789.) ")
			if step != "" {
				steps = append(steps, step)
			}
		}
	}
	out.FixSteps = steps
	return out
}
