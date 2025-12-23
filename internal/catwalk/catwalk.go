// Package catwalk provides model configuration types.
// This is a minimal stub replacing github.com/charmbracelet/catwalk.
package catwalk

import (
	"context"
	"errors"
)

// ErrNotModified is returned when content hasn't changed since last request.
var ErrNotModified = errors.New("not modified")

// InferenceProvider identifies an LLM provider.
type InferenceProvider string

const (
	InferenceProviderAnthropic     InferenceProvider = "anthropic"
	InferenceProviderBedrock       InferenceProvider = "bedrock"
	InferenceProviderOpenAI        InferenceProvider = "openai"
	InferenceProviderGoogle        InferenceProvider = "google"
	InferenceProviderOpenRouter    InferenceProvider = "openrouter"
	InferenceProviderAzure         InferenceProvider = "azure"
	InferenceProviderGitHubCopilot InferenceProvider = "github-copilot"
	InferenceProviderCopilot       InferenceProvider = "copilot"
	InferenceProviderZAI           InferenceProvider = "z-ai"
	InferenceProviderOpenAICompat  InferenceProvider = "openai-compat"
	InferenceProviderHyper         InferenceProvider = "hyper"
	InferenceProviderVertexAI      InferenceProvider = "vertex-ai"
)

// Type represents a provider type.
type Type string

const (
	TypeOpenAI       Type = "openai"
	TypeOpenAICompat Type = "openai-compat"
	TypeAnthropic    Type = "anthropic"
	TypeGoogle       Type = "google"
	TypeAzure        Type = "azure"
	TypeBedrock      Type = "bedrock"
	TypeVertexAI     Type = "vertexai"
	TypeOpenRouter   Type = "openrouter"
)

// Model represents a model configuration.
type Model struct {
	ID                     string            `json:"id"`
	Name                   string            `json:"name"`
	Provider               InferenceProvider `json:"provider,omitempty"`
	APIKey                 string            `json:"api_key,omitempty"`
	BaseURL                string            `json:"base_url,omitempty"`
	MaxTokens              int64             `json:"max_tokens,omitempty"`
	Temperature            float64           `json:"temperature,omitempty"`
	CostPer1MIn            float64           `json:"cost_per_1m_in,omitempty"`
	CostPer1MOut           float64           `json:"cost_per_1m_out,omitempty"`
	CostPer1MInCached      float64           `json:"cost_per_1m_in_cached,omitempty"`
	CostPer1MOutCached     float64           `json:"cost_per_1m_out_cached,omitempty"`
	ContextWindow          int64             `json:"context_window,omitempty"`
	DefaultMaxTokens       int64             `json:"default_max_tokens,omitempty"`
	CanReason              bool              `json:"can_reason,omitempty"`
	ReasoningLevels        []string          `json:"reasoning_levels,omitempty"`
	DefaultReasoningEffort string            `json:"default_reasoning_effort,omitempty"`
	SupportsImages         bool              `json:"supports_images,omitempty"`
	SupportsAttachments    bool              `json:"supports_attachments,omitempty"`
}

// ModelList is a list of models.
type ModelList []Model

// GetModelByID finds a model by its ID.
func (ml ModelList) GetModelByID(id string) *Model {
	for i := range ml {
		if ml[i].ID == id {
			return &ml[i]
		}
	}
	return nil
}

// Provider represents a model provider with its models.
type Provider struct {
	ID                  InferenceProvider `json:"id"`
	Name                string            `json:"name"`
	Type                Type              `json:"type"`
	BaseURL             string            `json:"base_url,omitempty"`
	APIEndpoint         string            `json:"api_endpoint,omitempty"`
	APIKey              string            `json:"api_key,omitempty"`
	Models              ModelList         `json:"models,omitempty"`
	Disabled            bool              `json:"disabled,omitempty"`
	DefaultHeaders      map[string]string `json:"default_headers,omitempty"`
	DefaultLargeModelID string            `json:"default_large_model_id,omitempty"`
	DefaultSmallModelID string            `json:"default_small_model_id,omitempty"`
}

// ProviderList is a list of providers.
type ProviderList []Provider

// GetProviderByID finds a provider by its ID.
func (pl ProviderList) GetProviderByID(id string) *Provider {
	for i := range pl {
		if string(pl[i].ID) == id {
			return &pl[i]
		}
	}
	return nil
}

// KnownProviderTypes returns a list of known provider types.
func KnownProviderTypes() []Type {
	return []Type{
		TypeOpenAI,
		TypeOpenAICompat,
		TypeAnthropic,
		TypeGoogle,
		TypeAzure,
		TypeBedrock,
		TypeVertexAI,
		TypeOpenRouter,
	}
}

// Client is a Catwalk API client.
type Client struct {
	baseURL string
}

// NewWithURL creates a new Catwalk client with the given base URL.
func NewWithURL(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// GetProviders fetches providers from the Catwalk API.
func (c *Client) GetProviders(ctx context.Context, etag string) ([]Provider, error) {
	// This is a stub implementation. In production, this would make an HTTP request.
	// For now, return an empty list since we typically use embedded providers.
	return []Provider{}, nil
}
