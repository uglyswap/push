// Package remote provides dynamic model fetching from provider APIs.
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/uglyswap/push/internal/catwalk"
	"github.com/uglyswap/push/internal/catwalk/embedded"
)

// ModelFetcher fetches models from provider APIs.
type ModelFetcher struct {
	client    *http.Client
	cache     map[string]cachedModels
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
	providers map[catwalk.InferenceProvider]providerConfig
}

type cachedModels struct {
	models    []catwalk.Model
	fetchedAt time.Time
}

type providerConfig struct {
	modelsEndpoint string
	parseFunc      func(body []byte) ([]catwalk.Model, error)
}

// NewModelFetcher creates a new model fetcher.
func NewModelFetcher() *ModelFetcher {
	return &ModelFetcher{
		client:   &http.Client{Timeout: 10 * time.Second},
		cache:    make(map[string]cachedModels),
		cacheTTL: 1 * time.Hour,
		providers: map[catwalk.InferenceProvider]providerConfig{
			catwalk.InferenceProviderAnthropic: {
				modelsEndpoint: "https://api.anthropic.com/v1/models",
				parseFunc:      parseAnthropicModels,
			},
			catwalk.InferenceProviderOpenAI: {
				modelsEndpoint: "https://api.openai.com/v1/models",
				parseFunc:      parseOpenAIModels,
			},
			catwalk.InferenceProviderGoogle: {
				modelsEndpoint: "https://generativelanguage.googleapis.com/v1beta/models",
				parseFunc:      parseGoogleModels,
			},
			catwalk.InferenceProviderOpenRouter: {
				modelsEndpoint: "https://openrouter.ai/api/v1/models",
				parseFunc:      parseOpenRouterModels,
			},
			catwalk.InferenceProviderGroq: {
				modelsEndpoint: "https://api.groq.com/openai/v1/models",
				parseFunc:      parseGroqModels,
			},
			catwalk.InferenceProviderZAI: {
				modelsEndpoint: "https://open.bigmodel.cn/api/paas/v4/models",
				parseFunc:      parseZhipuModels,
			},
		},
	}
}

// FetchModels fetches models for a provider, using cache if available.
func (f *ModelFetcher) FetchModels(ctx context.Context, providerID catwalk.InferenceProvider, apiKey string) ([]catwalk.Model, error) {
	// Check cache first
	cacheKey := string(providerID)
	f.cacheMu.RLock()
	if cached, ok := f.cache[cacheKey]; ok {
		if time.Since(cached.fetchedAt) < f.cacheTTL {
			f.cacheMu.RUnlock()
			return cached.models, nil
		}
	}
	f.cacheMu.RUnlock()

	// Get provider config
	config, ok := f.providers[providerID]
	if !ok {
		return nil, fmt.Errorf("no remote fetch support for provider: %s", providerID)
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, "GET", config.modelsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set auth header based on provider
	switch providerID {
	case catwalk.InferenceProviderAnthropic:
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	case catwalk.InferenceProviderOpenAI:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	case catwalk.InferenceProviderGoogle:
		// Google uses query param
		q := req.URL.Query()
		q.Set("key", apiKey)
		req.URL.RawQuery = q.Encode()
	case catwalk.InferenceProviderOpenRouter:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	case catwalk.InferenceProviderGroq:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	case catwalk.InferenceProviderZAI:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Make request
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse models
	models, err := config.parseFunc(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse models: %w", err)
	}

	// Enrich with embedded data (costs, etc.)
	models = enrichWithEmbedded(providerID, models)

	// Cache results
	f.cacheMu.Lock()
	f.cache[cacheKey] = cachedModels{
		models:    models,
		fetchedAt: time.Now(),
	}
	f.cacheMu.Unlock()

	return models, nil
}

// Anthropic models response
type anthropicModelsResponse struct {
	Data []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		CreatedAt   string `json:"created_at"`
		Type        string `json:"type"`
	} `json:"data"`
}

func parseAnthropicModels(body []byte) ([]catwalk.Model, error) {
	var resp anthropicModelsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var models []catwalk.Model
	for _, m := range resp.Data {
		// Only include claude models
		if !strings.HasPrefix(m.ID, "claude") {
			continue
		}
		name := m.DisplayName
		if name == "" {
			name = formatModelName(m.ID)
		}
		models = append(models, catwalk.Model{
			ID:   m.ID,
			Name: name,
		})
	}
	return models, nil
}

// OpenAI models response
type openAIModelsResponse struct {
	Data []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

func parseOpenAIModels(body []byte) ([]catwalk.Model, error) {
	var resp openAIModelsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var models []catwalk.Model
	for _, m := range resp.Data {
		// Filter to relevant models (GPT, o1, o3, o4)
		if !isRelevantOpenAIModel(m.ID) {
			continue
		}
		models = append(models, catwalk.Model{
			ID:   m.ID,
			Name: formatModelName(m.ID),
		})
	}
	return models, nil
}

func isRelevantOpenAIModel(id string) bool {
	relevantPrefixes := []string{"gpt-4", "gpt-3.5", "o1", "o3", "o4"}
	for _, prefix := range relevantPrefixes {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

// Groq models response
type groqModelsResponse struct {
	Data []struct {
		ID                  string `json:"id"`
		Object              string `json:"object"`
		Created             int64  `json:"created"`
		OwnedBy             string `json:"owned_by"`
		Active              bool   `json:"active"`
		ContextWindow       int64  `json:"context_window"`
		MaxCompletionTokens int64  `json:"max_completion_tokens"`
	} `json:"data"`
}

func parseGroqModels(body []byte) ([]catwalk.Model, error) {
	var resp groqModelsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var models []catwalk.Model
	for _, m := range resp.Data {
		// Only include active models
		if !m.Active {
			continue
		}
		models = append(models, catwalk.Model{
			ID:               m.ID,
			Name:             formatModelName(m.ID),
			ContextWindow:    m.ContextWindow,
			DefaultMaxTokens: m.MaxCompletionTokens,
		})
	}
	return models, nil
}

// Google models response
type googleModelsResponse struct {
	Models []struct {
		Name                       string   `json:"name"`
		DisplayName                string   `json:"displayName"`
		Description                string   `json:"description"`
		InputTokenLimit            int64    `json:"inputTokenLimit"`
		OutputTokenLimit           int64    `json:"outputTokenLimit"`
		SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	} `json:"models"`
}

func parseGoogleModels(body []byte) ([]catwalk.Model, error) {
	var resp googleModelsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var models []catwalk.Model
	for _, m := range resp.Models {
		// Filter to Gemini models with generateContent support
		if !strings.Contains(m.Name, "gemini") {
			continue
		}
		hasGenerateContent := false
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				hasGenerateContent = true
				break
			}
		}
		if !hasGenerateContent {
			continue
		}

		// Extract model ID from name (models/gemini-xxx -> gemini-xxx)
		id := strings.TrimPrefix(m.Name, "models/")

		models = append(models, catwalk.Model{
			ID:               id,
			Name:             m.DisplayName,
			ContextWindow:    m.InputTokenLimit,
			DefaultMaxTokens: m.OutputTokenLimit,
			SupportsImages:   strings.Contains(m.Description, "image") || strings.Contains(m.Description, "vision"),
		})
	}
	return models, nil
}

// OpenRouter models response
type openRouterModelsResponse struct {
	Data []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ContextLength int64  `json:"context_length"`
		Pricing       struct {
			Prompt     string `json:"prompt"`
			Completion string `json:"completion"`
		} `json:"pricing"`
		TopProvider struct {
			MaxCompletionTokens int64 `json:"max_completion_tokens"`
		} `json:"top_provider"`
	} `json:"data"`
}

func parseOpenRouterModels(body []byte) ([]catwalk.Model, error) {
	var resp openRouterModelsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var models []catwalk.Model
	for _, m := range resp.Data {
		// Parse pricing (comes as string like "0.000003")
		var costIn, costOut float64
		fmt.Sscanf(m.Pricing.Prompt, "%f", &costIn)
		fmt.Sscanf(m.Pricing.Completion, "%f", &costOut)
		// Convert from per-token to per-1M tokens
		costIn *= 1000000
		costOut *= 1000000

		models = append(models, catwalk.Model{
			ID:               m.ID,
			Name:             m.Name,
			ContextWindow:    m.ContextLength,
			DefaultMaxTokens: m.TopProvider.MaxCompletionTokens,
			CostPer1MIn:      costIn,
			CostPer1MOut:     costOut,
		})
	}
	return models, nil
}

// Zhipu models response
type zhipuModelsResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Object     string `json:"object"`
		OwnedBy    string `json:"owned_by"`
		Permission []any  `json:"permission"`
	} `json:"data"`
}

func parseZhipuModels(body []byte) ([]catwalk.Model, error) {
	var resp zhipuModelsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var models []catwalk.Model
	for _, m := range resp.Data {
		// Filter to GLM and CodeGeeX models
		if !strings.HasPrefix(m.ID, "glm") && !strings.HasPrefix(m.ID, "codegeex") {
			continue
		}
		models = append(models, catwalk.Model{
			ID:   m.ID,
			Name: formatModelName(m.ID),
		})
	}
	return models, nil
}

// enrichWithEmbedded adds cost and capability info from embedded models.
func enrichWithEmbedded(providerID catwalk.InferenceProvider, models []catwalk.Model) []catwalk.Model {
	// Get embedded provider
	var embeddedProvider *catwalk.Provider
	for _, p := range embedded.GetAll() {
		if p.ID == providerID {
			embeddedProvider = &p
			break
		}
	}
	if embeddedProvider == nil {
		return models
	}

	// Create lookup map
	embeddedMap := make(map[string]catwalk.Model)
	for _, m := range embeddedProvider.Models {
		embeddedMap[m.ID] = m
	}

	// Enrich fetched models
	for i := range models {
		if emb, ok := embeddedMap[models[i].ID]; ok {
			// Use embedded costs if fetched costs are zero
			if models[i].CostPer1MIn == 0 {
				models[i].CostPer1MIn = emb.CostPer1MIn
			}
			if models[i].CostPer1MOut == 0 {
				models[i].CostPer1MOut = emb.CostPer1MOut
			}
			// Use embedded context window if not set
			if models[i].ContextWindow == 0 {
				models[i].ContextWindow = emb.ContextWindow
			}
			if models[i].DefaultMaxTokens == 0 {
				models[i].DefaultMaxTokens = emb.DefaultMaxTokens
			}
			// Merge capabilities
			if emb.CanReason {
				models[i].CanReason = true
			}
			if emb.SupportsImages {
				models[i].SupportsImages = true
			}
			if len(emb.ReasoningLevels) > 0 && len(models[i].ReasoningLevels) == 0 {
				models[i].ReasoningLevels = emb.ReasoningLevels
			}
			if emb.DefaultReasoningEffort != "" && models[i].DefaultReasoningEffort == "" {
				models[i].DefaultReasoningEffort = emb.DefaultReasoningEffort
			}
		}
	}

	return models
}

// formatModelName converts model ID to display name.
func formatModelName(id string) string {
	// Replace common patterns
	name := strings.ReplaceAll(id, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Capitalize first letter of each word
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// GetProvidersWithModels fetches models for all providers that have API keys configured.
func (f *ModelFetcher) GetProvidersWithModels(ctx context.Context, apiKeys map[catwalk.InferenceProvider]string) []catwalk.Provider {
	// Start with embedded providers
	providers := embedded.GetAll()

	// Create a wait group for concurrent fetches
	var wg sync.WaitGroup
	results := make(chan struct {
		providerID catwalk.InferenceProvider
		models     []catwalk.Model
		err        error
	}, len(apiKeys))

	// Fetch models for each provider with an API key
	for providerID, apiKey := range apiKeys {
		if apiKey == "" {
			continue
		}
		wg.Add(1)
		go func(pid catwalk.InferenceProvider, key string) {
			defer wg.Done()
			models, err := f.FetchModels(ctx, pid, key)
			results <- struct {
				providerID catwalk.InferenceProvider
				models     []catwalk.Model
				err        error
			}{pid, models, err}
		}(providerID, apiKey)
	}

	// Close results channel when all fetches complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	fetchedModels := make(map[catwalk.InferenceProvider][]catwalk.Model)
	for result := range results {
		if result.err == nil && len(result.models) > 0 {
			fetchedModels[result.providerID] = result.models
		}
	}

	// Update providers with fetched models
	for i := range providers {
		if models, ok := fetchedModels[providers[i].ID]; ok {
			providers[i].Models = models
		}
	}

	return providers
}
