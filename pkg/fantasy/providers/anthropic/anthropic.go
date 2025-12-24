// Package anthropic provides Anthropic Claude API integration for fantasy.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/uglyswap/push/pkg/fantasy"
)

// Name is the provider name for Anthropic.
const Name = "anthropic"

const defaultBaseURL = "https://api.anthropic.com/v1"

// CacheControl represents Anthropic's cache control options.
type CacheControl struct {
	Type string `json:"type"`
}

// ProviderCacheControlOptions holds Anthropic-specific cache control options.
type ProviderCacheControlOptions struct {
	CacheControl CacheControl `json:"cache_control"`
}

// ReasoningOptionMetadata holds reasoning metadata from Anthropic.
type ReasoningOptionMetadata struct {
	Signature string `json:"signature"`
}

// Options holds Anthropic-specific options.
type Options struct {
	Thinking *ThinkingOptions `json:"thinking,omitempty"`
}

// ThinkingOptions holds thinking/reasoning configuration.
type ThinkingOptions struct {
	Enabled     bool   `json:"enabled,omitempty"`
	BudgetToken int64  `json:"budget_token,omitempty"`
	Type        string `json:"type,omitempty"`
}

// ParseOptions parses provider options into Anthropic options.
func ParseOptions(opts map[string]any) (*Options, error) {
	data, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}
	var o Options
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// Option configures the provider.
type Option func(*provider)

// provider implements fantasy.Provider for Anthropic.
type provider struct {
	apiKey     string
	baseURL    string
	headers    map[string]string
	httpClient *http.Client
}

// WithAPIKey sets the API key.
func WithAPIKey(key string) Option {
	return func(p *provider) {
		p.apiKey = key
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(url string) Option {
	return func(p *provider) {
		p.baseURL = url
	}
}

// WithHeaders sets custom headers.
func WithHeaders(headers map[string]string) Option {
	return func(p *provider) {
		p.headers = headers
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(p *provider) {
		p.httpClient = client
	}
}

// New creates a new Anthropic provider.
func New(opts ...Option) (fantasy.Provider, error) {
	p := &provider{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

// LanguageModel returns a language model for the given model ID.
func (p *provider) LanguageModel(ctx context.Context, modelID string) (fantasy.LanguageModel, error) {
	return &Client{
		apiKey:     p.apiKey,
		baseURL:    p.baseURL,
		headers:    p.headers,
		httpClient: p.httpClient,
		model:      modelID,
	}, nil
}

// Client is an Anthropic API client.
type Client struct {
	apiKey     string
	baseURL    string
	headers    map[string]string
	httpClient *http.Client
	model      string
}

// NewClient creates a new Anthropic client.
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		model: model,
	}
}

// WithBaseURL sets the base URL.
func (c *Client) WithBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

// WithHeaders sets custom headers.
func (c *Client) WithHeaders(headers map[string]string) *Client {
	c.headers = headers
	return c
}

// WithHTTPClient sets a custom HTTP client.
func (c *Client) WithHTTPClient(client *http.Client) *Client {
	c.httpClient = client
	return c
}

// Model returns the model name.
func (c *Client) Model() string {
	return c.model
}

// Provider returns the provider name.
func (c *Client) Provider() string {
	return Name
}

// messageRequest represents the API request.
type messageRequest struct {
	Model       string           `json:"model"`
	MaxTokens   int64            `json:"max_tokens"`
	Messages    []messagePayload `json:"messages"`
	System      string           `json:"system,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
	TopP        *float64         `json:"top_p,omitempty"`
	TopK        *int64           `json:"top_k,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Tools       []toolPayload    `json:"tools,omitempty"`
}

type messagePayload struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type toolPayload struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type messageResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []contentBlock `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        usageInfo `json:"usage"`
}

type contentBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type usageInfo struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens int64 `json:"cache_read_input_tokens,omitempty"`
}

// Generate performs a non-streaming generation.
func (c *Client) Generate(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions) (*fantasy.Response, error) {
	req, systemPrompt := c.buildRequest(messages, opts, false)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &fantasy.ProviderError{
			Title:      "API Error",
			Message:    string(bodyBytes),
			StatusCode: resp.StatusCode,
			Provider:   Name,
		}
	}

	var msgResp messageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.convertResponse(&msgResp, systemPrompt), nil
}

// Stream performs a streaming generation.
func (c *Client) Stream(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions, callbacks fantasy.StreamCallbacks) (*fantasy.Response, error) {
	req, _ := c.buildRequest(messages, opts, true)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &fantasy.ProviderError{
			Title:      "API Error",
			Message:    string(bodyBytes),
			StatusCode: resp.StatusCode,
			Provider:   Name,
		}
	}

	// Process SSE stream
	var fullText string
	var usage fantasy.Usage
	var finishReason fantasy.FinishReason = fantasy.FinishReasonStop

	decoder := json.NewDecoder(resp.Body)
	for {
		// Read SSE event
		var line string
		if _, err := fmt.Fscanln(resp.Body, &line); err != nil {
			if err == io.EOF {
				break
			}
			// Try to continue on non-fatal errors
			continue
		}

		// Parse event data
		if len(line) > 6 && line[:6] == "data: " {
			data := line[6:]
			if data == "[DONE]" {
				break
			}

			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Usage usageInfo `json:"usage"`
			}
			if err := decoder.Decode(&event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta.Type == "text_delta" {
					fullText += event.Delta.Text
					if callbacks.OnTextDelta != nil {
						if err := callbacks.OnTextDelta("", event.Delta.Text); err != nil {
							return nil, err
						}
					}
				}
			case "message_delta":
				usage.InputTokens = event.Usage.InputTokens
				usage.OutputTokens = event.Usage.OutputTokens
				usage.CacheCreationTokens = event.Usage.CacheCreationInputTokens
				usage.CacheReadTokens = event.Usage.CacheReadInputTokens
			}
		}
	}

	return &fantasy.Response{
		Content: fantasy.ResponseContent{
			Parts: []fantasy.MessagePart{fantasy.TextPart{Text: fullText}},
		},
		FinishReason: finishReason,
		Usage:        usage,
	}, nil
}

func (c *Client) buildRequest(messages []fantasy.Message, opts fantasy.GenerateOptions, stream bool) (*messageRequest, string) {
	var systemPrompt string
	var msgPayloads []messagePayload

	for _, msg := range messages {
		if msg.Role == fantasy.MessageRoleSystem {
			for _, part := range msg.Content {
				if tp, ok := part.(fantasy.TextPart); ok {
					if systemPrompt != "" {
						systemPrompt += "\n\n"
					}
					systemPrompt += tp.Text
				}
			}
			continue
		}

		var content []interface{}
		for _, part := range msg.Content {
			switch p := part.(type) {
			case fantasy.TextPart:
				content = append(content, map[string]string{
					"type": "text",
					"text": p.Text,
				})
			case fantasy.FilePart:
				content = append(content, map[string]interface{}{
					"type": "image",
					"source": map[string]interface{}{
						"type":       "base64",
						"media_type": p.MediaType,
						"data":       string(p.Data),
					},
				})
			}
		}

		if len(content) > 0 {
			msgPayloads = append(msgPayloads, messagePayload{
				Role:    string(msg.Role),
				Content: content,
			})
		}
	}

	maxTokens := int64(4096)
	if opts.MaxOutputTokens != nil {
		maxTokens = *opts.MaxOutputTokens
	}

	req := &messageRequest{
		Model:       c.model,
		MaxTokens:   maxTokens,
		Messages:    msgPayloads,
		System:      systemPrompt,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		TopK:        opts.TopK,
		Stream:      stream,
	}

	// Add tools if provided
	for _, tool := range opts.Tools {
		req.Tools = append(req.Tools, toolPayload{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.Parameters(),
		})
	}

	return req, systemPrompt
}

func (c *Client) convertResponse(resp *messageResponse, systemPrompt string) *fantasy.Response {
	var parts []fantasy.MessagePart
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			parts = append(parts, fantasy.TextPart{Text: block.Text})
		case "tool_use":
			parts = append(parts, fantasy.ToolCallPart{
				ToolCallID: block.ID,
				ToolName:   block.Name,
				Input:      string(block.Input),
			})
		}
	}

	finishReason := fantasy.FinishReasonStop
	switch resp.StopReason {
	case "end_turn":
		finishReason = fantasy.FinishReasonStop
	case "max_tokens":
		finishReason = fantasy.FinishReasonLength
	case "tool_use":
		finishReason = fantasy.FinishReasonToolCalls
	}

	return &fantasy.Response{
		Content: fantasy.ResponseContent{Parts: parts},
		FinishReason: finishReason,
		Usage: fantasy.Usage{
			InputTokens:         resp.Usage.InputTokens,
			OutputTokens:        resp.Usage.OutputTokens,
			CacheCreationTokens: resp.Usage.CacheCreationInputTokens,
			CacheReadTokens:     resp.Usage.CacheReadInputTokens,
			TotalTokens:         resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		ProviderMetadata: fantasy.ProviderMetadata{
			Name: map[string]interface{}{
				"model": resp.Model,
				"id":    resp.ID,
			},
		},
	}
}

// Verify Client implements LanguageModel
var _ fantasy.LanguageModel = (*Client)(nil)

// Verify provider implements Provider
var _ fantasy.Provider = (*provider)(nil)
