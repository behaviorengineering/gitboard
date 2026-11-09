package triage

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/behaviorengineering/gitboard/internal/ai/llm"

	ollyerrors "github.com/behaviorengineering/olly/pkg/errors"
	"go.opentelemetry.io/otel"
)

// systemPrompt sets the model's role for every triage call.
const systemPrompt = "You triage CI failures for a Go monorepo. Do not invent file paths or line numbers absent from the log."

// promptTmpl is the user prompt. parseAnswer reads the same section names.
const promptTmpl = `Analyze this CI failure. Use ONLY the log excerpt below.
{{- if .JobName}}
Job: {{.JobName}}
{{- end}}

Respond in plain text with sections:
SUMMARY: one sentence
ROOT_CAUSE: short paragraph
FIX_STEPS: numbered list
CONFIDENCE: high|medium|low

LOG:
{{.Log}}`

var prompt = template.Must(template.New("triage").Parse(promptTmpl))

// promptData is what promptTmpl renders.
type promptData struct {
	JobName string
	Log     string
}

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

func (a *Analyzer) Enabled() bool {
	return a != nil && a.LLM.Enabled()
}

// Analyze sends logs to the configured model.
func (a *Analyzer) Analyze(ctx context.Context, req Request) (resp Response, err error) {
	ctx, span := otel.Tracer("gitboard").Start(ctx, "triage.Analyze")
	defer func() { ollyerrors.EndWithStatus(span, err) }()

	if !a.Enabled() {
		return Response{
			Unavailable: "set llm.base_url in config.yaml (or GITBOARD_LLM_BASE_URL / POLYPUS_BASE_URL) for AI triage",
		}, nil
	}
	logText := strings.TrimSpace(req.Log)
	if logText == "" {
		return Response{}, fmt.Errorf("empty log excerpt")
	}
	user, err := buildPrompt(req, logText)
	if err != nil {
		return Response{}, err
	}
	raw, model, err := a.LLM.Chat(ctx, systemPrompt, user)
	if err != nil {
		return Response{}, fmt.Errorf("triage chat: %w", err)
	}
	return parseAnswer(raw, model), nil
}

func buildPrompt(req Request, logText string) (string, error) {
	var b strings.Builder
	if err := prompt.Execute(&b, promptData{JobName: req.JobName, Log: logText}); err != nil {
		return "", fmt.Errorf("triage prompt: %w", err)
	}
	return b.String(), nil
}

func parseAnswer(raw, model string) Response {
	sections := llm.ParseSections(raw, "SUMMARY", "ROOT_CAUSE", "FIX_STEPS", "CONFIDENCE")
	out := Response{
		RawAnswer:  raw,
		Model:      model,
		Summary:    strings.Join(sections["SUMMARY"], " "),
		RootCause:  strings.Join(sections["ROOT_CAUSE"], " "),
		Confidence: llm.First(sections["CONFIDENCE"]),
	}
	for _, line := range sections["FIX_STEPS"] {
		if step := strings.TrimLeft(line, "0123456789.) "); step != "" {
			out.FixSteps = append(out.FixSteps, step)
		}
	}
	return out
}
