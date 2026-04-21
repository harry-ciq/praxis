// Package llm wraps the Anthropic Messages API for Praxis-internal use.
//
// We talk HTTP directly (no SDK) — it's one endpoint and the payload is
// trivial. Keeping the dep surface small avoids pulling in a heavy SDK tree.
//
// Prompt caching: the system prompt is marked with `cache_control: ephemeral`
// so repeat calls within 5 minutes read from cache at ~10% of the usual
// input-token price.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	anthropicEndpoint    = "https://api.anthropic.com/v1/messages"
	anthropicVersion     = "2023-06-01"
	defaultMaxTokens     = 1024
	defaultHTTPTimeout   = 30 * time.Second
	anthropicBetaHeader  = "prompt-caching-2024-07-31"
)

// Client talks to the Anthropic Messages API.
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient constructs a client. model can be any Anthropic model string, e.g.
// "claude-haiku-4-5" or "claude-sonnet-4-5".
func NewClient(apiKey, model string, logger *zap.Logger) *Client {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		logger:     logger,
	}
}

// IsConfigured returns true when an API key is set. Handlers can use this to
// return a graceful 503 when the feature is disabled in this environment.
func (c *Client) IsConfigured() bool {
	return c != nil && c.apiKey != ""
}

// ContentBlock is one block in a user or system message. Setting CacheControl
// to "ephemeral" marks the block for prompt caching (5-minute TTL).
type ContentBlock struct {
	Type         string              `json:"type"`
	Text         string              `json:"text"`
	CacheControl *cacheControlMarker `json:"cache_control,omitempty"`
}

type cacheControlMarker struct {
	Type string `json:"type"`
}

// CacheableBlock builds a cacheable text block.
func CacheableBlock(text string) ContentBlock {
	return ContentBlock{
		Type:         "text",
		Text:         text,
		CacheControl: &cacheControlMarker{Type: "ephemeral"},
	}
}

// TextBlock builds a plain text block (not cached).
func TextBlock(text string) ContentBlock {
	return ContentBlock{Type: "text", Text: text}
}

type message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type messagesRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	System    []ContentBlock `json:"system,omitempty"`
	Messages  []message      `json:"messages"`
}

type messagesResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

type errorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends a single-turn request to the Messages API. Returns the
// concatenated text of the first content block and the token-usage report.
type Usage struct {
	InputTokens              int
	OutputTokens             int
	CacheCreationInputTokens int
	CacheReadInputTokens     int
}

func (c *Client) Complete(ctx context.Context, system []ContentBlock, user []ContentBlock, maxTokens int) (string, *Usage, error) {
	if !c.IsConfigured() {
		return "", nil, fmt.Errorf("anthropic client not configured (ANTHROPIC_API_KEY missing)")
	}
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	reqBody := messagesRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    system,
		Messages: []message{{
			Role:    "user",
			Content: user,
		}},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("anthropic-beta", anthropicBetaHeader)

	started := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var er errorResponse
		if err := json.Unmarshal(body, &er); err == nil && er.Error.Message != "" {
			return "", nil, fmt.Errorf("anthropic %d: %s (%s)", resp.StatusCode, er.Error.Message, er.Error.Type)
		}
		return "", nil, fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(body))
	}

	var mr messagesResponse
	if err := json.Unmarshal(body, &mr); err != nil {
		return "", nil, fmt.Errorf("parse response: %w", err)
	}

	if len(mr.Content) == 0 {
		return "", nil, fmt.Errorf("anthropic returned no content")
	}

	usage := &Usage{
		InputTokens:              mr.Usage.InputTokens,
		OutputTokens:             mr.Usage.OutputTokens,
		CacheCreationInputTokens: mr.Usage.CacheCreationInputTokens,
		CacheReadInputTokens:     mr.Usage.CacheReadInputTokens,
	}

	if c.logger != nil {
		c.logger.Info("anthropic call",
			zap.String("model", mr.Model),
			zap.Duration("duration", time.Since(started)),
			zap.Int("inputTokens", usage.InputTokens),
			zap.Int("outputTokens", usage.OutputTokens),
			zap.Int("cacheCreation", usage.CacheCreationInputTokens),
			zap.Int("cacheRead", usage.CacheReadInputTokens),
		)
	}

	// Concatenate all text blocks in the response.
	var out string
	for _, blk := range mr.Content {
		if blk.Type == "text" {
			out += blk.Text
		}
	}
	return out, usage, nil
}
