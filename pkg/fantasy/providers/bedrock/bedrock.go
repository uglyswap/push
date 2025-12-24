// Package bedrock provides AWS Bedrock integration for fantasy.
package bedrock

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/uglyswap/push/pkg/fantasy"
	"github.com/uglyswap/push/pkg/fantasy/providers/anthropic"
)

// Name is the provider name for Bedrock.
const Name = "bedrock"

// Re-export Anthropic cache control options for Bedrock compatibility.
type ProviderCacheControlOptions = anthropic.ProviderCacheControlOptions

// Option configures the provider.
type Option func(*provider)

// provider implements fantasy.Provider for AWS Bedrock.
type provider struct {
	region       string
	accessKey    string
	secretKey    string
	sessionToken string
	headers      map[string]string
	httpClient   *http.Client
}

// WithRegion sets the AWS region.
func WithRegion(region string) Option {
	return func(p *provider) {
		p.region = region
	}
}

// WithCredentials sets AWS credentials.
func WithCredentials(accessKey, secretKey, sessionToken string) Option {
	return func(p *provider) {
		p.accessKey = accessKey
		p.secretKey = secretKey
		p.sessionToken = sessionToken
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

// WithAPIKey sets a bearer token for authentication.
// This is an alternative to IAM credentials, using AWS bearer token auth.
func WithAPIKey(bearerToken string) Option {
	return func(p *provider) {
		// Bearer token auth would require different request signing
		// For now, store it as session token for compatibility
		p.sessionToken = bearerToken
	}
}

// New creates a new Bedrock provider.
func New(opts ...Option) (fantasy.Provider, error) {
	p := &provider{
		region:       "us-east-1",
		accessKey:    os.Getenv("AWS_ACCESS_KEY_ID"),
		secretKey:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		sessionToken: os.Getenv("AWS_SESSION_TOKEN"),
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
		region:       p.region,
		model:        modelID,
		accessKey:    p.accessKey,
		secretKey:    p.secretKey,
		sessionToken: p.sessionToken,
		httpClient:   p.httpClient,
	}, nil
}

// Client is an AWS Bedrock client.
type Client struct {
	region       string
	model        string
	accessKey    string
	secretKey    string
	sessionToken string
	httpClient   *http.Client
}

// NewClient creates a new Bedrock client.
func NewClient(region string, model string) *Client {
	return &Client{
		region:       region,
		model:        model,
		accessKey:    os.Getenv("AWS_ACCESS_KEY_ID"),
		secretKey:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		sessionToken: os.Getenv("AWS_SESSION_TOKEN"),
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// WithCredentials sets AWS credentials explicitly.
func (c *Client) WithCredentials(accessKey, secretKey, sessionToken string) *Client {
	c.accessKey = accessKey
	c.secretKey = secretKey
	c.sessionToken = sessionToken
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

// Bedrock uses Anthropic's message format for Claude models
type bedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens        int64            `json:"max_tokens"`
	Messages         []bedrockMessage `json:"messages"`
	System           string           `json:"system,omitempty"`
	Temperature      *float64         `json:"temperature,omitempty"`
	TopP             *float64         `json:"top_p,omitempty"`
	TopK             *int64           `json:"top_k,omitempty"`
}

type bedrockMessage struct {
	Role    string               `json:"role"`
	Content []bedrockContentPart `json:"content"`
}

type bedrockContentPart struct {
	Type      string             `json:"type"`
	Text      string             `json:"text,omitempty"`
	Source    *bedrockImageSource `json:"source,omitempty"`
	ID        string             `json:"id,omitempty"`
	Name      string             `json:"name,omitempty"`
	Input     interface{}        `json:"input,omitempty"`
	ToolUseID string             `json:"tool_use_id,omitempty"`
	Content   string             `json:"content,omitempty"`
}

type bedrockImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type bedrockResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Content      []bedrockContentPart `json:"content"`
	Model        string               `json:"model"`
	StopReason   string               `json:"stop_reason"`
	StopSequence string               `json:"stop_sequence"`
	Usage        bedrockUsage         `json:"usage"`
}

type bedrockUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

type bedrockStreamEvent struct {
	Type         string `json:"type"`
	Message      *bedrockResponse `json:"message,omitempty"`
	Index        int    `json:"index,omitempty"`
	ContentBlock *bedrockContentPart `json:"content_block,omitempty"`
	Delta        *bedrockDelta `json:"delta,omitempty"`
	Usage        *bedrockUsage `json:"usage,omitempty"`
}

type bedrockDelta struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	PartialJSON  string `json:"partial_json,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
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

	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke", c.region, c.model)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Creation Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Sign the request with AWS SigV4
	if err := c.signRequest(httpReq, reqBody); err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Signing Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	resp, err := c.httpClient.Do(httpReq)
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

	if resp.StatusCode != http.StatusOK {
		return nil, &fantasy.ProviderError{
			Title:      "API Error",
			Message:    string(body),
			StatusCode: resp.StatusCode,
			Provider:   Name,
		}
	}

	var bedrockResp bedrockResponse
	if err := json.Unmarshal(body, &bedrockResp); err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Response Parse Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	return c.parseResponse(&bedrockResp)
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

	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke-with-response-stream", c.region, c.model)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Creation Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/vnd.amazon.eventstream")

	// Sign the request with AWS SigV4
	if err := c.signRequest(httpReq, reqBody); err != nil {
		return nil, &fantasy.ProviderError{
			Title:    "Request Signing Failed",
			Message:  err.Error(),
			Provider: Name,
		}
	}

	resp, err := c.httpClient.Do(httpReq)
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
		return nil, &fantasy.ProviderError{
			Title:      "API Error",
			Message:    string(body),
			StatusCode: resp.StatusCode,
			Provider:   Name,
		}
	}

	return c.parseStream(resp.Body, callbacks)
}

func (c *Client) buildRequest(messages []fantasy.Message, opts fantasy.GenerateOptions) (*bedrockRequest, error) {
	req := &bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Temperature:      opts.Temperature,
		TopP:             opts.TopP,
		TopK:             opts.TopK,
	}

	if opts.MaxOutputTokens != nil {
		req.MaxTokens = *opts.MaxOutputTokens
	}

	bedrockMsgs := make([]bedrockMessage, 0, len(messages))

	for _, msg := range messages {
		// Handle system messages
		if msg.Role == fantasy.MessageRoleSystem {
			for _, part := range msg.Content {
				if tp, ok := part.(fantasy.TextPart); ok {
					req.System = tp.Text
				}
			}
			continue
		}

		bedrockMsg, err := c.convertMessage(msg)
		if err != nil {
			return nil, err
		}
		bedrockMsgs = append(bedrockMsgs, bedrockMsg)
	}

	req.Messages = bedrockMsgs
	return req, nil
}

func (c *Client) convertMessage(msg fantasy.Message) (bedrockMessage, error) {
	role := "user"
	if msg.Role == fantasy.MessageRoleAssistant {
		role = "assistant"
	}

	parts := make([]bedrockContentPart, 0, len(msg.Content))

	for _, part := range msg.Content {
		switch p := part.(type) {
		case fantasy.TextPart:
			parts = append(parts, bedrockContentPart{
				Type: "text",
				Text: p.Text,
			})

		case fantasy.FilePart:
			parts = append(parts, bedrockContentPart{
				Type: "image",
				Source: &bedrockImageSource{
					Type:      "base64",
					MediaType: p.MediaType,
					Data:      string(p.Data),
				},
			})

		case fantasy.ToolCallPart:
			var input interface{}
			if err := json.Unmarshal([]byte(p.Input), &input); err != nil {
				input = p.Input
			}
			parts = append(parts, bedrockContentPart{
				Type:  "tool_use",
				ID:    p.ToolCallID,
				Name:  p.ToolName,
				Input: input,
			})

		case fantasy.ToolResultPart:
			var content string
			if textOutput, ok := p.Output.(fantasy.ToolResultOutputContentText); ok {
				content = textOutput.Text
			} else if errOutput, ok := p.Output.(fantasy.ToolResultOutputContentError); ok {
				content = errOutput.Error.Error()
			}
			parts = append(parts, bedrockContentPart{
				Type:      "tool_result",
				ToolUseID: p.ToolCallID,
				Content:   content,
			})
		}
	}

	return bedrockMessage{
		Role:    role,
		Content: parts,
	}, nil
}

func (c *Client) parseResponse(resp *bedrockResponse) (*fantasy.Response, error) {
	parts := make([]fantasy.MessagePart, 0)

	for _, content := range resp.Content {
		switch content.Type {
		case "text":
			parts = append(parts, fantasy.TextPart{Text: content.Text})
		case "tool_use":
			inputJSON, _ := json.Marshal(content.Input)
			parts = append(parts, fantasy.ToolCallPart{
				ToolCallID: content.ID,
				ToolName:   content.Name,
				Input:      string(inputJSON),
			})
		}
	}

	finishReason := fantasy.FinishReasonStop
	switch resp.StopReason {
	case "end_turn", "stop_sequence":
		finishReason = fantasy.FinishReasonStop
	case "max_tokens":
		finishReason = fantasy.FinishReasonLength
	case "tool_use":
		finishReason = fantasy.FinishReasonToolCalls
	}

	return &fantasy.Response{
		Content:      fantasy.ResponseContent{Parts: parts},
		FinishReason: finishReason,
		Usage: fantasy.Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}, nil
}

func (c *Client) parseStream(body io.Reader, callbacks fantasy.StreamCallbacks) (*fantasy.Response, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var fullText strings.Builder
	var toolCalls []fantasy.ToolCallPart
	var usage fantasy.Usage
	var finishReason fantasy.FinishReason = fantasy.FinishReasonStop
	var currentToolCall *fantasy.ToolCallPart
	var toolInputBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Parse event stream format
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "" {
			continue
		}

		var event bedrockStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			// Try parsing as raw bytes (Bedrock event stream format)
			continue
		}

		switch event.Type {
		case "content_block_start":
			if event.ContentBlock != nil {
				if event.ContentBlock.Type == "tool_use" {
					currentToolCall = &fantasy.ToolCallPart{
						ToolCallID: event.ContentBlock.ID,
						ToolName:   event.ContentBlock.Name,
					}
					toolInputBuilder.Reset()
				}
			}

		case "content_block_delta":
			if event.Delta != nil {
				if event.Delta.Text != "" {
					fullText.WriteString(event.Delta.Text)
					if callbacks.OnTextDelta != nil {
						if err := callbacks.OnTextDelta("", event.Delta.Text); err != nil {
							return nil, err
						}
					}
				}
				if event.Delta.PartialJSON != "" && currentToolCall != nil {
					toolInputBuilder.WriteString(event.Delta.PartialJSON)
				}
			}

		case "content_block_stop":
			if currentToolCall != nil {
				currentToolCall.Input = toolInputBuilder.String()
				toolCalls = append(toolCalls, *currentToolCall)
				if callbacks.OnToolCall != nil {
					if err := callbacks.OnToolCall(fantasy.ToolCallContent{
						ToolCallID: currentToolCall.ToolCallID,
						ToolName:   currentToolCall.ToolName,
						Input:      currentToolCall.Input,
					}); err != nil {
						return nil, err
					}
				}
				currentToolCall = nil
			}

		case "message_delta":
			if event.Delta != nil && event.Delta.StopReason != "" {
				switch event.Delta.StopReason {
				case "end_turn", "stop_sequence":
					finishReason = fantasy.FinishReasonStop
				case "max_tokens":
					finishReason = fantasy.FinishReasonLength
				case "tool_use":
					finishReason = fantasy.FinishReasonToolCalls
				}
			}
			if event.Usage != nil {
				usage = fantasy.Usage{
					InputTokens:  event.Usage.InputTokens,
					OutputTokens: event.Usage.OutputTokens,
					TotalTokens:  event.Usage.InputTokens + event.Usage.OutputTokens,
				}
			}

		case "message_stop":
			// End of stream
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

// AWS SigV4 signing implementation
func (c *Client) signRequest(req *http.Request, body []byte) error {
	if c.accessKey == "" || c.secretKey == "" {
		return fmt.Errorf("AWS credentials not configured")
	}

	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	// Set required headers
	req.Header.Set("Host", req.Host)
	req.Header.Set("X-Amz-Date", amzDate)
	if c.sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", c.sessionToken)
	}

	// Create canonical request
	payloadHash := sha256Hex(body)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalQueryString := req.URL.RawQuery

	// Build signed headers
	signedHeaders := []string{}
	canonicalHeaders := ""
	for key := range req.Header {
		lowerKey := strings.ToLower(key)
		signedHeaders = append(signedHeaders, lowerKey)
	}
	sort.Strings(signedHeaders)
	signedHeadersStr := strings.Join(signedHeaders, ";")

	for _, key := range signedHeaders {
		canonicalHeaders += key + ":" + strings.TrimSpace(req.Header.Get(key)) + "\n"
	}

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeadersStr,
		payloadHash,
	}, "\n")

	// Create string to sign
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := strings.Join([]string{dateStamp, c.region, "bedrock", "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		algorithm,
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	// Calculate signature
	signingKey := getSignatureKey(c.secretKey, dateStamp, c.region, "bedrock")
	signature := hmacSHA256Hex(signingKey, stringToSign)

	// Build authorization header
	authHeader := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, c.accessKey, credentialScope, signedHeadersStr, signature)
	req.Header.Set("Authorization", authHeader)

	return nil
}

func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hmacSHA256Hex(key []byte, data string) string {
	return hex.EncodeToString(hmacSHA256(key, data))
}

func getSignatureKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}

// Verify Client implements LanguageModel
var _ fantasy.LanguageModel = (*Client)(nil)

// Verify provider implements Provider
var _ fantasy.Provider = (*provider)(nil)
