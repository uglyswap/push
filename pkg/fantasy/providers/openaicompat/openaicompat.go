// Package openaicompat provides a generic OpenAI-compatible API integration for fantasy.
// This supports providers like ZAI (z.ai), GLM-4, and other OpenAI-compatible APIs.
package openaicompat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/uglyswap/crush/pkg/fantasy"
)

// Name is the provider name for OpenAI-compatible APIs.
const Name = "openai-compat"

// Options holds OpenAI-compatible provider-specific options.
type Options struct {
	ExtraBody map[string]interface{} `json:"extra_body,omitempty"`
}

// ParseOptions parses provider options into OpenAI-compatible options.
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

// SDKOptions holds SDK-level options that can be passed through.
type SDKOptions struct {
	ExtraBody map[string]interface{} `json:"extra_body,omitempty"`
}

// Option configures the provider.
type Option func(*provider)

// provider implements fantasy.Provider for OpenAI-compatible APIs.
type provider struct {
	apiKey     string
	baseURL    string
	headers    map[string]string
	sdkOptions *SDKOptions
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
		p.baseURL = strings.TrimSuffix(url, "/")
	}
}

// WithHeaders sets custom headers.
func WithHeaders(headers map[string]string) Option {
	return func(p *provider) {
		p.headers = headers
	}
}

// WithSDKOptions sets SDK-level options.
func WithSDKOptions(opts *SDKOptions) Option {
	return func(p *provider) {
		p.sdkOptions = opts
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(p *provider) {
		p.httpClient = client
	}
}

// New creates a new OpenAI-compatible provider.
func New(opts ...Option) (fantasy.Provider, error) {
	p := &provider{
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
	client := &Client{
		apiKey:     p.apiKey,
		baseURL:    p.baseURL,
		model:      modelID,
		headers:    p.headers,
		httpClient: p.httpClient,
	}
	if p.sdkOptions != nil {
		client.extraBody = p.sdkOptions.ExtraBody
	}
	return client, nil
}

// Client is a generic OpenAI-compatible API client.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	headers    map[string]string
	extraBody  map[string]interface{}
	httpClient *http.Client
}

// NewClient creates a new OpenAI-compatible client.
func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// WithHeaders sets custom headers for requests.
func (c *Client) WithHeaders(headers map[string]string) *Client {
	c.headers = headers
	return c
}

// WithExtraBody sets extra body parameters for requests.
func (c *Client) WithExtraBody(extra map[string]interface{}) *Client {
	c.extraBody = extra
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

// chatRequest represents an OpenAI-compatible Chat Completion request.
type chatRequest struct {
	Model            string                 `json:"model"`
	Messages         []chatMessage          `json:"messages"`
	MaxTokens        *int64                 `json:"max_tokens,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Tools            []chatTool             `json:"tools,omitempty"`
	StreamOptions    *streamOptions         `json:"stream_options,omitempty"`
	Extra            map[string]interface{} `json:"-"` // Will be merged into JSON
}

// MarshalJSON custom marshaler to include extra fields.
func (r chatRequest) MarshalJSON() ([]byte, error) {
	type Alias chatRequest
	data, err := json.Marshal(Alias(r))
	if err != nil {
		return nil, err
	}

	if len(r.Extra) == 0 {
		return data, nil
	}

	// Merge extra fields
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	for k, v := range r.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatMessage struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"`
	Name       string      `json:"name,omitempty"`
	ToolCalls  []toolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type chatTool struct {
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

type toolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type toolCall struct {
	Index    int              `json:"index"`
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function toolCallFunction `json:"function"`
}

type toolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []choiceObject `json:"choices"`
	Usage   usageObject    `json:"usage"`
	Error   *errorObject   `json:"error,omitempty"`
}

type choiceObject struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	Delta        chatMessage `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

type usageObject struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type errorObject struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Generate performs a non-streaming generation.
func (c *Client) Generate(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions) (*fantasy.Response, error) {
	req, err := c.buildRequest(messages, opts, false)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	return c.parseResponse(resp)
}

// Stream performs a streaming generation.
func (c *Client) Stream(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions, callbacks fantasy.StreamCallbacks) (*fantasy.Response, error) {
	req, err := c.buildRequest(messages, opts, true)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(req))
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Creation Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	c.setHeaders(httpReq)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		var errResp chatResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != nil {
			return nil, &fantasy.ProviderError{
				Title:      errResp.Error.Type,
				Message:    errResp.Error.Message,
				StatusCode: httpResp.StatusCode,
				Provider:   Name,
			}
		}
		return nil, &fantasy.ProviderError{
			Title:      "API Error",
			Message:    string(body),
			StatusCode: httpResp.StatusCode,
			Provider:   Name,
		}
	}

	return c.parseSSE(httpResp.Body, callbacks)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
}

func (c *Client) buildRequest(messages []fantasy.Message, opts fantasy.GenerateOptions, stream bool) ([]byte, error) {
	chatMsgs := make([]chatMessage, 0, len(messages))

	for _, msg := range messages {
		chatMsg, err := c.convertMessage(msg)
		if err != nil {
			return nil, err
		}
		chatMsgs = append(chatMsgs, chatMsg)
	}

	req := chatRequest{
		Model:            c.model,
		Messages:         chatMsgs,
		MaxTokens:        opts.MaxOutputTokens,
		Temperature:      opts.Temperature,
		TopP:             opts.TopP,
		FrequencyPenalty: opts.FrequencyPenalty,
		PresencePenalty:  opts.PresencePenalty,
		Stream:           stream,
		Extra:            c.extraBody,
	}

	if stream {
		req.StreamOptions = &streamOptions{IncludeUsage: true}
	}

	// Convert tools
	if len(opts.Tools) > 0 {
		req.Tools = make([]chatTool, 0, len(opts.Tools))
		for _, tool := range opts.Tools {
			req.Tools = append(req.Tools, chatTool{
				Type: "function",
				Function: toolFunction{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  tool.Parameters(),
				},
			})
		}
	}

	return json.Marshal(req)
}

func (c *Client) convertMessage(msg fantasy.Message) (chatMessage, error) {
	role := string(msg.Role)

	// Check if message contains tool results
	for _, part := range msg.Content {
		if trp, ok := part.(fantasy.ToolResultPart); ok {
			var content string
			if textOutput, ok := trp.Output.(fantasy.ToolResultOutputContentText); ok {
				content = textOutput.Text
			} else if errOutput, ok := trp.Output.(fantasy.ToolResultOutputContentError); ok {
				content = errOutput.Error.Error()
			}
			return chatMessage{
				Role:       "tool",
				Content:    content,
				ToolCallID: trp.ToolCallID,
			}, nil
		}
	}

	// Check for tool calls in assistant messages
	var toolCalls []toolCall
	for _, part := range msg.Content {
		if tcp, ok := part.(fantasy.ToolCallPart); ok {
			toolCalls = append(toolCalls, toolCall{
				ID:   tcp.ToolCallID,
				Type: "function",
				Function: toolCallFunction{
					Name:      tcp.ToolName,
					Arguments: tcp.Input,
				},
			})
		}
	}

	if len(toolCalls) > 0 {
		return chatMessage{
			Role:      "assistant",
			Content:   "",
			ToolCalls: toolCalls,
		}, nil
	}

	// Check for multipart content (text + images)
	var hasImages bool
	for _, part := range msg.Content {
		if _, ok := part.(fantasy.FilePart); ok {
			hasImages = true
			break
		}
	}

	if hasImages {
		parts := make([]contentPart, 0, len(msg.Content))
		for _, part := range msg.Content {
			switch p := part.(type) {
			case fantasy.TextPart:
				parts = append(parts, contentPart{
					Type: "text",
					Text: p.Text,
				})
			case fantasy.FilePart:
				dataURL := fmt.Sprintf("data:%s;base64,%s", p.MediaType, string(p.Data))
				parts = append(parts, contentPart{
					Type: "image_url",
					ImageURL: &imageURL{
						URL:    dataURL,
						Detail: "auto",
					},
				})
			}
		}
		return chatMessage{
			Role:    role,
			Content: parts,
		}, nil
	}

	// Simple text message
	var text string
	for _, part := range msg.Content {
		if tp, ok := part.(fantasy.TextPart); ok {
			text = tp.Text
			break
		}
	}

	return chatMessage{
		Role:    role,
		Content: text,
	}, nil
}

func (c *Client) doRequest(ctx context.Context, body []byte) (*chatResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Creation Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Response Read Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Response Parse Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	if chatResp.Error != nil {
		return nil, &fantasy.ProviderError{
			Title:      chatResp.Error.Type,
			Message:    chatResp.Error.Message,
			StatusCode: resp.StatusCode,
			Provider:   Name,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &fantasy.ProviderError{
			Title:      "API Error",
			Message:    string(respBody),
			StatusCode: resp.StatusCode,
			Provider:   Name,
		}
	}

	return &chatResp, nil
}

func (c *Client) parseResponse(resp *chatResponse) (*fantasy.Response, error) {
	if len(resp.Choices) == 0 {
		return &fantasy.Response{
			Content:      fantasy.ResponseContent{Parts: []fantasy.MessagePart{}},
			FinishReason: fantasy.FinishReasonStop,
			Usage: fantasy.Usage{
				InputTokens:  resp.Usage.PromptTokens,
				OutputTokens: resp.Usage.CompletionTokens,
				TotalTokens:  resp.Usage.TotalTokens,
			},
		}, nil
	}

	choice := resp.Choices[0]
	parts := make([]fantasy.MessagePart, 0)

	if content, ok := choice.Message.Content.(string); ok && content != "" {
		parts = append(parts, fantasy.TextPart{Text: content})
	}

	for _, tc := range choice.Message.ToolCalls {
		parts = append(parts, fantasy.ToolCallPart{
			ToolCallID: tc.ID,
			ToolName:   tc.Function.Name,
			Input:      tc.Function.Arguments,
		})
	}

	finishReason := fantasy.FinishReasonStop
	switch choice.FinishReason {
	case "stop":
		finishReason = fantasy.FinishReasonStop
	case "length":
		finishReason = fantasy.FinishReasonLength
	case "tool_calls":
		finishReason = fantasy.FinishReasonToolCalls
	}

	return &fantasy.Response{
		Content:      fantasy.ResponseContent{Parts: parts},
		FinishReason: finishReason,
		Usage: fantasy.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}, nil
}

func (c *Client) parseSSE(body io.Reader, callbacks fantasy.StreamCallbacks) (*fantasy.Response, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var fullText strings.Builder
	var toolCalls []fantasy.ToolCallPart
	var usage fantasy.Usage
	var finishReason fantasy.FinishReason = fantasy.FinishReasonStop

	toolCallDeltas := make(map[int]*toolCall)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk chatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if chunk.Usage.TotalTokens > 0 {
			usage = fantasy.Usage{
				InputTokens:  chunk.Usage.PromptTokens,
				OutputTokens: chunk.Usage.CompletionTokens,
				TotalTokens:  chunk.Usage.TotalTokens,
			}
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]

		if content, ok := choice.Delta.Content.(string); ok && content != "" {
			fullText.WriteString(content)
			if callbacks.OnTextDelta != nil {
				if err := callbacks.OnTextDelta(chunk.ID, content); err != nil {
					return nil, err
				}
			}
		}

		for _, tc := range choice.Delta.ToolCalls {
			if _, exists := toolCallDeltas[tc.Index]; !exists {
				toolCallDeltas[tc.Index] = &toolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: toolCallFunction{
						Name:      tc.Function.Name,
						Arguments: "",
					},
				}
			}
			if tc.Function.Arguments != "" {
				toolCallDeltas[tc.Index].Function.Arguments += tc.Function.Arguments
			}
		}

		if choice.FinishReason != "" {
			switch choice.FinishReason {
			case "stop":
				finishReason = fantasy.FinishReasonStop
			case "length":
				finishReason = fantasy.FinishReasonLength
			case "tool_calls":
				finishReason = fantasy.FinishReasonToolCalls
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Stream Read Error",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	for _, tc := range toolCallDeltas {
		toolCallPart := fantasy.ToolCallPart{
			ToolCallID: tc.ID,
			ToolName:   tc.Function.Name,
			Input:      tc.Function.Arguments,
		}
		toolCalls = append(toolCalls, toolCallPart)

		if callbacks.OnToolCall != nil {
			if err := callbacks.OnToolCall(fantasy.ToolCallContent{
				ToolCallID: tc.ID,
				ToolName:   tc.Function.Name,
				Input:      tc.Function.Arguments,
			}); err != nil {
				return nil, err
			}
		}
	}

	parts := make([]fantasy.MessagePart, 0)
	if fullText.Len() > 0 {
		parts = append(parts, fantasy.TextPart{Text: fullText.String()})
	}
	for _, tc := range toolCalls {
		parts = append(parts, tc)
	}

	return &fantasy.Response{
		Content:      fantasy.ResponseContent{Parts: parts},
		FinishReason: finishReason,
		Usage:        usage,
	}, nil
}

// Verify Client implements LanguageModel
var _ fantasy.LanguageModel = (*Client)(nil)

// Verify provider implements Provider
var _ fantasy.Provider = (*provider)(nil)
