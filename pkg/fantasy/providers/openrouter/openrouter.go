// Package openrouter provides OpenRouter API integration for fantasy.
package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uglyswap/push/pkg/fantasy"
	"github.com/uglyswap/push/pkg/fantasy/providers/openai"
)

// Name is the provider name for OpenRouter.
const Name = "openrouter"

const baseURL = "https://openrouter.ai/api/v1"

// Options holds OpenRouter-specific options.
type Options struct {
	Reasoning *ReasoningOptions `json:"reasoning,omitempty"`
}

// ReasoningOptions holds reasoning configuration.
type ReasoningOptions struct {
	Enabled bool   `json:"enabled,omitempty"`
	Effort  string `json:"effort,omitempty"`
}

// ProviderMetadata holds OpenRouter-specific metadata from responses.
type ProviderMetadata struct {
	Usage ProviderUsage `json:"usage,omitempty"`
}

// ProviderUsage holds usage information from OpenRouter.
type ProviderUsage struct {
	Cost float64 `json:"cost,omitempty"`
}

// ParseOptions parses provider options into OpenRouter options.
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

// provider implements fantasy.Provider for OpenRouter.
type provider struct {
	apiKey     string
	headers    map[string]string
	httpClient *http.Client
}

// WithAPIKey sets the API key.
func WithAPIKey(key string) Option {
	return func(p *provider) {
		p.apiKey = key
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

// New creates a new OpenRouter provider.
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
	// Create an OpenAI client configured for OpenRouter
	client := openai.NewClient(p.apiKey, modelID)
	client.WithBaseURL(baseURL)
	return &Client{
		openaiClient: client,
		model:        modelID,
	}, nil
}

// Client is an OpenRouter API client.
type Client struct {
	openaiClient *openai.Client
	model        string
}

// NewClient creates a new OpenRouter client.
func NewClient(apiKey string, model string) *Client {
	openaiClient := openai.NewClient(apiKey, model)
	openaiClient.WithBaseURL(baseURL)

	return &Client{
		openaiClient: openaiClient,
		model:        model,
	}
}

// WithHTTPClient sets a custom HTTP client.
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	return c
}

// WithTimeout sets the request timeout.
func (c *Client) WithTimeout(timeout time.Duration) *Client {
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

// Generate performs a non-streaming generation.
func (c *Client) Generate(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions) (*fantasy.Response, error) {
	return c.openaiClient.Generate(ctx, messages, opts)
}

// Stream performs a streaming generation.
func (c *Client) Stream(ctx context.Context, messages []fantasy.Message, opts fantasy.GenerateOptions, callbacks fantasy.StreamCallbacks) (*fantasy.Response, error) {
	return c.openaiClient.Stream(ctx, messages, opts, callbacks)
}

// Verify Client implements LanguageModel
var _ fantasy.LanguageModel = (*Client)(nil)

// Verify provider implements Provider
var _ fantasy.Provider = (*provider)(nil)
