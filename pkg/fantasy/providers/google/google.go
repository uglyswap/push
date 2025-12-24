// Package google provides Google AI (Gemini) integration for fantasy.
package google

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

	"github.com/uglyswap/push/pkg/fantasy"
)

// Name is the provider name for Google.
const Name = "google"

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Options holds Google-specific options.
type Options struct {
	ThinkingConfig *ThinkingConfig `json:"thinking_config,omitempty"`
}

// ThinkingConfig holds thinking configuration.
type ThinkingConfig struct {
	ThinkingBudget  int64 `json:"thinking_budget,omitempty"`
	IncludeThoughts bool  `json:"include_thoughts,omitempty"`
}

// ReasoningMetadata holds reasoning metadata for Google provider.
type ReasoningMetadata struct {
	Signature string `json:"signature,omitempty"`
	ToolID    string `json:"tool_id,omitempty"`
}

// ParseOptions parses provider options into Google options.
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

// provider implements fantasy.Provider for Google.
type provider struct {
	apiKey     string
	baseURL    string
	headers    map[string]string
	httpClient *http.Client
	// Vertex AI configuration
	vertexProject  string
	vertexLocation string
	useVertex      bool
}

// WithGeminiAPIKey sets the API key for Gemini API.
func WithGeminiAPIKey(key string) Option {
	return func(p *provider) {
		p.apiKey = key
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(url string) Option {
	return func(p *provider) {
		if url != "" {
			p.baseURL = strings.TrimSuffix(url, "/")
		}
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

// WithVertex configures the provider for Vertex AI.
func WithVertex(project, location string) Option {
	return func(p *provider) {
		p.useVertex = true
		p.vertexProject = project
		p.vertexLocation = location
		p.baseURL = fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1", location)
	}
}

// New creates a new Google AI provider.
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
		provider: p,
		model:    modelID,
	}, nil
}

// Client is a Google AI client.
type Client struct {
	provider *provider
	model    string
}

// NewClient creates a new Google AI client (for backwards compatibility).
func NewClient(apiKey string, model string) *Client {
	return &Client{
		provider: &provider{
			apiKey:  apiKey,
			baseURL: defaultBaseURL,
			httpClient: &http.Client{
				Timeout: 5 * time.Minute,
			},
		},
		model: model,
	}
}

// Model returns the model name.
func (c *Client) Model() string {
	return c.model
}

// Provider returns the provider name.
func (c *Client) Provider() string {
	return Name
}

// Gemini API request/response types
type geminiRequest struct {
	Contents          []geminiContent         `json:"contents"`
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
	Tools             []geminiTool            `json:"tools,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string                   `json:"text,omitempty"`
	InlineData       *geminiInlineData        `json:"inlineData,omitempty"`
	FunctionCall     *geminiFunctionCall      `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse  `json:"functionResponse,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens *int64   `json:"maxOutputTokens,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int64   `json:"topK,omitempty"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

type geminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata *geminiUsage      `json:"usageMetadata"`
	Error         *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content       geminiContent `json:"content"`
	FinishReason  string        `json:"finishReason"`
	SafetyRatings []interface{} `json:"safetyRatings"`
}

type geminiUsage struct {
	PromptTokenCount     int64 `json:"promptTokenCount"`
	CandidatesTokenCount int64 `json:"candidatesTokenCount"`
	TotalTokenCount      int64 `json:"totalTokenCount"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Generate performs a non-streaming generation.
func (c *Client) Generate(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions) (*fantasy.Response, error) {
	req, err := c.buildRequest(messages, opts)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Serialization Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	url := c.buildURL(false)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Creation Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.provider.httpClient.Do(httpReq)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Response Read Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Response Parse Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	if geminiResp.Error != nil {
		return nil, &fantasy.ProviderError{
			Title:      geminiResp.Error.Status,
			Message:    geminiResp.Error.Message,
			StatusCode: geminiResp.Error.Code,
			Provider:   Name,
		}
	}

	return c.parseResponse(&geminiResp)
}

// Stream performs a streaming generation.
func (c *Client) Stream(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions, callbacks fantasy.StreamCallbacks) (*fantasy.Response, error) {
	req, err := c.buildRequest(messages, opts)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Serialization Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	url := c.buildURL(true)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Creation Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.provider.httpClient.Do(httpReq)
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp geminiResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != nil {
			return nil, &fantasy.ProviderError{
				Title:      errResp.Error.Status,
				Message:    errResp.Error.Message,
				StatusCode: resp.StatusCode,
				Provider:   Name,
			}
		}
		return nil, &fantasy.ProviderError{
			Title:      "API Error",
			Message:    string(body),
			StatusCode: resp.StatusCode,
			Provider:   Name,
		}
	}

	return c.parseSSE(resp.Body, callbacks)
}

func (c *Client) buildURL(stream bool) string {
	if c.provider.useVertex {
		// Vertex AI URL format
		endpoint := "generateContent"
		if stream {
			endpoint = "streamGenerateContent?alt=sse"
		}
		return fmt.Sprintf("%s/projects/%s/locations/%s/publishers/google/models/%s:%s",
			c.provider.baseURL, c.provider.vertexProject, c.provider.vertexLocation, c.model, endpoint)
	}

	// Gemini API URL format
	if stream {
		return fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s",
			c.provider.baseURL, c.model, c.provider.apiKey)
	}
	return fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		c.provider.baseURL, c.model, c.provider.apiKey)
}

func (c *Client) buildRequest(messages []fantasy.Message, opts fantasy.GenerateOptions) (*geminiRequest, error) {
	req := &geminiRequest{}

	// Set generation config
	if opts.MaxOutputTokens != nil || opts.Temperature != nil || opts.TopP != nil || opts.TopK != nil {
		req.GenerationConfig = &geminiGenerationConfig{
			MaxOutputTokens: opts.MaxOutputTokens,
			Temperature:     opts.Temperature,
			TopP:            opts.TopP,
			TopK:            opts.TopK,
		}
	}

	// Convert tools
	if len(opts.Tools) > 0 {
		funcs := make([]geminiFunctionDeclaration, 0, len(opts.Tools))
		for _, tool := range opts.Tools {
			funcs = append(funcs, geminiFunctionDeclaration{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			})
		}
		req.Tools = []geminiTool{{FunctionDeclarations: funcs}}
	}

	// Convert messages
	contents := make([]geminiContent, 0, len(messages))
	for _, msg := range messages {
		// Handle system messages specially
		if msg.Role == fantasy.MessageRoleSystem {
			parts := make([]geminiPart, 0)
			for _, part := range msg.Content {
				if tp, ok := part.(fantasy.TextPart); ok {
					parts = append(parts, geminiPart{Text: tp.Text})
				}
			}
			req.SystemInstruction = &geminiContent{Parts: parts}
			continue
		}

		content, err := c.convertMessage(msg)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}
	req.Contents = contents

	return req, nil
}

func (c *Client) convertMessage(msg fantasy.Message) (geminiContent, error) {
	role := "user"
	switch msg.Role {
	case fantasy.MessageRoleUser:
		role = "user"
	case fantasy.MessageRoleAssistant:
		role = "model"
	case fantasy.MessageRoleTool:
		role = "user" // Tool results come from "user" in Gemini
	}

	parts := make([]geminiPart, 0, len(msg.Content))

	for _, part := range msg.Content {
		switch p := part.(type) {
		case fantasy.TextPart:
			parts = append(parts, geminiPart{Text: p.Text})

		case fantasy.FilePart:
			parts = append(parts, geminiPart{
				InlineData: &geminiInlineData{
					MimeType: p.MediaType,
					Data:     string(p.Data),
				},
			})

		case fantasy.ToolCallPart:
			// Parse the input as JSON
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(p.Input), &args); err != nil {
				args = map[string]interface{}{"input": p.Input}
			}
			parts = append(parts, geminiPart{
				FunctionCall: &geminiFunctionCall{
					Name: p.ToolName,
					Args: args,
				},
			})

		case fantasy.ToolResultPart:
			var resultMap map[string]interface{}
			if textOutput, ok := p.Output.(fantasy.ToolResultOutputContentText); ok {
				resultMap = map[string]interface{}{"result": textOutput.Text}
			} else if errOutput, ok := p.Output.(fantasy.ToolResultOutputContentError); ok {
				resultMap = map[string]interface{}{"error": errOutput.Error.Error()}
			} else {
				resultMap = map[string]interface{}{"result": ""}
			}
			parts = append(parts, geminiPart{
				FunctionResponse: &geminiFunctionResponse{
					Name:     p.ToolCallID, // Gemini uses the function name
					Response: resultMap,
				},
			})
		}
	}

	return geminiContent{
		Role:  role,
		Parts: parts,
	}, nil
}

func (c *Client) parseResponse(resp *geminiResponse) (*fantasy.Response, error) {
	if len(resp.Candidates) == 0 {
		return &fantasy.Response{
			Content:      fantasy.ResponseContent{Parts: []fantasy.MessagePart{}},
			FinishReason: fantasy.FinishReasonStop,
			Usage:        c.convertUsage(resp.UsageMetadata),
		}, nil
	}

	candidate := resp.Candidates[0]
	parts := make([]fantasy.MessagePart, 0)

	for _, p := range candidate.Content.Parts {
		if p.Text != "" {
			parts = append(parts, fantasy.TextPart{Text: p.Text})
		}
		if p.FunctionCall != nil {
			argsJSON, _ := json.Marshal(p.FunctionCall.Args)
			parts = append(parts, fantasy.ToolCallPart{
				ToolCallID: p.FunctionCall.Name, // Use name as ID
				ToolName:   p.FunctionCall.Name,
				Input:      string(argsJSON),
			})
		}
	}

	finishReason := fantasy.FinishReasonStop
	switch candidate.FinishReason {
	case "STOP":
		finishReason = fantasy.FinishReasonStop
	case "MAX_TOKENS":
		finishReason = fantasy.FinishReasonLength
	case "TOOL_CODE", "FUNCTION_CALL":
		finishReason = fantasy.FinishReasonToolCalls
	}

	return &fantasy.Response{
		Content:      fantasy.ResponseContent{Parts: parts},
		FinishReason: finishReason,
		Usage:        c.convertUsage(resp.UsageMetadata),
	}, nil
}

func (c *Client) convertUsage(usage *geminiUsage) fantasy.Usage {
	if usage == nil {
		return fantasy.Usage{}
	}
	return fantasy.Usage{
		InputTokens:  usage.PromptTokenCount,
		OutputTokens: usage.CandidatesTokenCount,
		TotalTokens:  usage.TotalTokenCount,
	}
}

func (c *Client) parseSSE(body io.Reader, callbacks fantasy.StreamCallbacks) (*fantasy.Response, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var fullText strings.Builder
	var toolCalls []fantasy.ToolCallPart
	var usage fantasy.Usage
	var finishReason fantasy.FinishReason = fantasy.FinishReasonStop

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "" || data == "[DONE]" {
			continue
		}

		var chunk geminiResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Update usage
		if chunk.UsageMetadata != nil {
			usage = c.convertUsage(chunk.UsageMetadata)
		}

		if len(chunk.Candidates) == 0 {
			continue
		}

		candidate := chunk.Candidates[0]

		for _, p := range candidate.Content.Parts {
			if p.Text != "" {
				fullText.WriteString(p.Text)
				if callbacks.OnTextDelta != nil {
					if err := callbacks.OnTextDelta("", p.Text); err != nil {
						return nil, err
					}
				}
			}

			if p.FunctionCall != nil {
				argsJSON, _ := json.Marshal(p.FunctionCall.Args)
				toolCall := fantasy.ToolCallPart{
					ToolCallID: p.FunctionCall.Name,
					ToolName:   p.FunctionCall.Name,
					Input:      string(argsJSON),
				}
				toolCalls = append(toolCalls, toolCall)

				if callbacks.OnToolCall != nil {
					if err := callbacks.OnToolCall(fantasy.ToolCallContent{
						ToolCallID: p.FunctionCall.Name,
						ToolName:   p.FunctionCall.Name,
						Input:      string(argsJSON),
					}); err != nil {
						return nil, err
					}
				}
			}
		}

		// Update finish reason
		if candidate.FinishReason != "" {
			switch candidate.FinishReason {
			case "STOP":
				finishReason = fantasy.FinishReasonStop
			case "MAX_TOKENS":
				finishReason = fantasy.FinishReasonLength
			case "TOOL_CODE", "FUNCTION_CALL":
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

	// Build response parts
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
