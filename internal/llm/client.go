// Package llm provides an OpenAI-compatible chat completions client.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
)

// Client calls an OpenAI-compatible /chat/completions endpoint.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// New builds a client from effective LLM settings (file + env overlay).
func New(cfg config.LLM) *Client {
	return &Client{
		BaseURL: strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		APIKey:  strings.TrimSpace(cfg.APIKey),
		Model:   firstNonEmpty(cfg.Model, config.DefaultLLMModel),
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

// Enabled reports whether a base URL is configured.
func (c *Client) Enabled() bool {
	return c != nil && strings.TrimSpace(c.BaseURL) != ""
}

// Chat sends a system + user message pair and returns assistant content and model id.
func (c *Client) Chat(ctx context.Context, system, user string) (content, model string, err error) {
	if !c.Enabled() {
		return "", "", fmt.Errorf("llm: base_url is not set")
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 120 * time.Second}
	}
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("llm marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(rawBody))
	if err != nil {
		return "", "", fmt.Errorf("llm request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("llm chat: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", "", fmt.Errorf("llm read: %w", err)
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
		return "", "", fmt.Errorf("llm parse: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", "", fmt.Errorf("llm: empty choices")
	}
	return parsed.Choices[0].Message.Content, firstNonEmpty(parsed.Model, c.Model), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
