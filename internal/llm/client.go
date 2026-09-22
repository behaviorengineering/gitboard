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
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// Client calls an OpenAI-compatible /chat/completions endpoint.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client

	resilienceMu sync.Mutex
	resilience   *clientResilience
}

// New builds a client from effective LLM settings (file + env overlay).
func New(cfg config.LLM) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		APIKey:     strings.TrimSpace(cfg.APIKey),
		Model:      firstNonEmpty(cfg.Model, config.DefaultLLMModel),
		HTTP:       &http.Client{Timeout: 120 * time.Second},
		resilience: newClientResilience(),
	}
}

// Enabled reports whether a base URL is configured.
func (c *Client) Enabled() bool {
	return c != nil && strings.TrimSpace(c.BaseURL) != ""
}

// Chat sends a system + user message pair and returns assistant content and model id.
// Requires ctx deadline. Transient transport and 429/5xx failures retry with failsafe-go.
func (c *Client) Chat(ctx context.Context, system, user string) (content, model string, err error) {
	tr := otel.Tracer("gitboard")
	ctx, span := tr.Start(ctx, "llm.Chat")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

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

	result, err := c.runChatWithResilience(ctx, func() (chatResult, error) {
		httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(rawBody))
		if reqErr != nil {
			return chatResult{}, fmt.Errorf("llm request: %w", reqErr)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if c.APIKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
		}
		resp, doErr := httpClient.Do(httpReq)
		if doErr != nil {
			return chatResult{}, fmt.Errorf("llm chat: %w", doErr)
		}
		defer func() { _ = resp.Body.Close() }()
		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		if readErr != nil {
			return chatResult{}, fmt.Errorf("llm read: %w", readErr)
		}
		if resp.StatusCode >= 400 {
			return chatResult{}, httpStatusError(resp.Status, resp.StatusCode, string(respBody))
		}
		var parsed struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Model string `json:"model"`
		}
		if unmarshalErr := json.Unmarshal(respBody, &parsed); unmarshalErr != nil {
			return chatResult{}, fmt.Errorf("llm parse: %w", unmarshalErr)
		}
		if len(parsed.Choices) == 0 {
			return chatResult{}, fmt.Errorf("llm: empty choices")
		}
		return chatResult{
			content: parsed.Choices[0].Message.Content,
			model:   firstNonEmpty(parsed.Model, c.Model),
		}, nil
	})
	if err != nil {
		return "", "", err
	}
	return result.content, result.model, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
