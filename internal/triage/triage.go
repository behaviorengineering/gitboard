package triage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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

// Analyzer calls an OpenAI-compatible chat endpoint.
type Analyzer struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// NewFromEnv builds an analyzer from GITBOARD_LLM_* or POLYPUS_BASE_URL.
func NewFromEnv() *Analyzer {
	base := firstNonEmpty(
		os.Getenv("GITBOARD_LLM_BASE_URL"),
		os.Getenv("POLYPUS_BASE_URL"),
	)
	if base != "" && !strings.HasSuffix(base, "/v1") {
		base = strings.TrimRight(base, "/") + "/v1"
	}
	return &Analyzer{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey:  firstNonEmpty(os.Getenv("GITBOARD_LLM_API_KEY"), os.Getenv("OPENAI_API_KEY")),
		Model:   firstNonEmpty(os.Getenv("GITBOARD_LLM_MODEL"), "cf_local/@cf/zai-org/glm-4.7-flash"),
		Client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *Analyzer) Enabled() bool {
	return a != nil && strings.TrimSpace(a.BaseURL) != ""
}

// Analyze sends logs to the configured model.
func (a *Analyzer) Analyze(ctx context.Context, req Request) (Response, error) {
	if !a.Enabled() {
		return Response{
			Unavailable: "set GITBOARD_LLM_BASE_URL (or POLYPUS_BASE_URL) for AI triage",
		}, nil
	}
	logText := strings.TrimSpace(req.Log)
	if logText == "" {
		return Response{}, fmt.Errorf("empty log excerpt")
	}
	prompt := buildPrompt(req, logText)
	raw, model, err := a.chat(ctx, prompt)
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

func (a *Analyzer) chat(ctx context.Context, prompt string) (string, string, error) {
	body := map[string]any{
		"model": a.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You triage CI failures for a Go monorepo. Do not invent file paths or line numbers absent from the log."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
	}
	rawBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BaseURL+"/chat/completions", bytes.NewReader(rawBody))
	if err != nil {
		return "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)
	}
	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("llm %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", "", err
	}
	if len(parsed.Choices) == 0 {
		return "", "", fmt.Errorf("llm: empty choices")
	}
	model := firstNonEmpty(parsed.Model, a.Model)
	return parsed.Choices[0].Message.Content, model, nil
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
