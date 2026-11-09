// Package llm provides an OpenAI-compatible chat completions client.
package llm

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"

	ollyerrors "github.com/behaviorengineering/olly/pkg/errors"
	"go.opentelemetry.io/otel"
)

// maxResponseBytes caps a chat completion body read into memory.
const maxResponseBytes = 2 << 20

// chatTemperature keeps answers close to the evidence for triage and prune
// cards.
const chatTemperature = 0.2

// Message roles in a chat request.
const (
	roleSystem = "system"
	roleUser   = "user"
)

// errNoHTTPClient is returned by Chat on a Client built without New and
// without an HTTP client.
var errNoHTTPClient = errors.New("llm: http client is not set")

// Client calls an OpenAI-compatible /chat/completions endpoint.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	// HTTP sends the requests. It has no timeout of its own: every call is
	// bounded by the caller's ctx deadline, which Chat requires.
	HTTP *http.Client

	resilienceMu sync.Mutex
	resilience   *clientResilience
}

// New builds a client from effective LLM settings (file + env overlay).
// Panics if httpClient is nil (required dependency at wire-up).
func New(cfg boardconfig.LLM, httpClient *http.Client) *Client {
	if httpClient == nil {
		panic("llm.New: http client is required")
	}
	return &Client{
		BaseURL:    strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		APIKey:     strings.TrimSpace(cfg.APIKey),
		Model:      cmp.Or(strings.TrimSpace(cfg.Model), config.DefaultLLMModel),
		HTTP:       httpClient,
		resilience: newClientResilience(),
	}
}

// Enabled reports whether a base URL is configured.
func (c *Client) Enabled() bool {
	return c != nil && strings.TrimSpace(c.BaseURL) != ""
}

// chatMessage is one message of a chat request.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the /chat/completions request body.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// chatResponse is the part of the /chat/completions response Chat reads.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
}

// Chat sends a system + user message pair and returns assistant content and model id.
// Requires ctx deadline. Transient transport and 429/5xx failures retry with failsafe-go.
func (c *Client) Chat(ctx context.Context, system, user string) (content, model string, err error) {
	ctx, span := otel.Tracer("gitboard").Start(ctx, "llm.Chat")
	defer func() { ollyerrors.EndWithStatus(span, err) }()

	if !c.Enabled() {
		return "", "", fmt.Errorf("llm: base_url is not set")
	}
	rawBody, err := json.Marshal(chatRequest{
		Model:       c.Model,
		Messages:    []chatMessage{{Role: roleSystem, Content: system}, {Role: roleUser, Content: user}},
		Temperature: chatTemperature,
	})
	if err != nil {
		return "", "", fmt.Errorf("llm marshal: %w", err)
	}
	result, err := c.runChatWithResilience(ctx, func() (chatResult, error) {
		return c.chatOnce(ctx, rawBody)
	})
	if err != nil {
		return "", "", err
	}
	return result.content, result.model, nil
}

// chatOnce sends one request; the resilience policy decides on retries.
func (c *Client) chatOnce(ctx context.Context, rawBody []byte) (chatResult, error) {
	if c.HTTP == nil {
		return chatResult{}, errNoHTTPClient
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(rawBody))
	if err != nil {
		return chatResult{}, fmt.Errorf("llm request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return chatResult{}, fmt.Errorf("llm chat: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return chatResult{}, fmt.Errorf("llm read: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return chatResult{}, httpStatusError(resp.Status, resp.StatusCode, string(respBody))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return chatResult{}, fmt.Errorf("llm parse: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return chatResult{}, fmt.Errorf("llm: empty choices")
	}
	return chatResult{
		content: parsed.Choices[0].Message.Content,
		model:   cmp.Or(strings.TrimSpace(parsed.Model), c.Model),
	}, nil
}
