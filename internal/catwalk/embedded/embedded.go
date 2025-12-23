// Package embedded provides embedded provider configurations.
// This is a stub replacing github.com/charmbracelet/catwalk/pkg/embedded.
package embedded

import "github.com/uglyswap/crush/internal/catwalk"

// GetAll returns all embedded providers.
func GetAll() []catwalk.Provider {
	return []catwalk.Provider{
		{
			ID:      catwalk.InferenceProviderAnthropic,
			Name:    "Anthropic",
			Type:    catwalk.TypeAnthropic,
			BaseURL: "https://api.anthropic.com/v1",
			Models: catwalk.ModelList{
				{
					ID:               "claude-sonnet-4-20250514",
					Name:             "Claude Sonnet 4",
					CostPer1MIn:      3.0,
					CostPer1MOut:     15.0,
					ContextWindow:    200000,
					DefaultMaxTokens: 8192,
					CanReason:        true,
					SupportsImages:   true,
				},
				{
					ID:               "claude-3-5-sonnet-20241022",
					Name:             "Claude 3.5 Sonnet",
					CostPer1MIn:      3.0,
					CostPer1MOut:     15.0,
					ContextWindow:    200000,
					DefaultMaxTokens: 8192,
					SupportsImages:   true,
				},
				{
					ID:               "claude-3-5-haiku-20241022",
					Name:             "Claude 3.5 Haiku",
					CostPer1MIn:      0.8,
					CostPer1MOut:     4.0,
					ContextWindow:    200000,
					DefaultMaxTokens: 8192,
					SupportsImages:   true,
				},
				{
					ID:               "claude-3-opus-20240229",
					Name:             "Claude 3 Opus",
					CostPer1MIn:      15.0,
					CostPer1MOut:     75.0,
					ContextWindow:    200000,
					DefaultMaxTokens: 4096,
					SupportsImages:   true,
				},
			},
		},
		{
			ID:      catwalk.InferenceProviderOpenAI,
			Name:    "OpenAI",
			Type:    catwalk.TypeOpenAI,
			BaseURL: "https://api.openai.com/v1",
			Models: catwalk.ModelList{
				{
					ID:               "gpt-4o",
					Name:             "GPT-4o",
					CostPer1MIn:      2.5,
					CostPer1MOut:     10.0,
					ContextWindow:    128000,
					DefaultMaxTokens: 4096,
					SupportsImages:   true,
				},
				{
					ID:               "gpt-4o-mini",
					Name:             "GPT-4o Mini",
					CostPer1MIn:      0.15,
					CostPer1MOut:     0.6,
					ContextWindow:    128000,
					DefaultMaxTokens: 4096,
					SupportsImages:   true,
				},
				{
					ID:                     "o1",
					Name:                   "o1",
					CostPer1MIn:            15.0,
					CostPer1MOut:           60.0,
					ContextWindow:          200000,
					DefaultMaxTokens:       100000,
					CanReason:              true,
					ReasoningLevels:        []string{"low", "medium", "high"},
					DefaultReasoningEffort: "medium",
				},
				{
					ID:                     "o1-mini",
					Name:                   "o1-mini",
					CostPer1MIn:            3.0,
					CostPer1MOut:           12.0,
					ContextWindow:          128000,
					DefaultMaxTokens:       65536,
					CanReason:              true,
					ReasoningLevels:        []string{"low", "medium", "high"},
					DefaultReasoningEffort: "medium",
				},
			},
		},
		{
			ID:      catwalk.InferenceProviderGoogle,
			Name:    "Google AI",
			Type:    catwalk.TypeGoogle,
			BaseURL: "https://generativelanguage.googleapis.com/v1beta",
			Models: catwalk.ModelList{
				{
					ID:               "gemini-2.0-flash",
					Name:             "Gemini 2.0 Flash",
					CostPer1MIn:      0.1,
					CostPer1MOut:     0.4,
					ContextWindow:    1000000,
					DefaultMaxTokens: 8192,
					SupportsImages:   true,
				},
				{
					ID:               "gemini-1.5-pro",
					Name:             "Gemini 1.5 Pro",
					CostPer1MIn:      1.25,
					CostPer1MOut:     5.0,
					ContextWindow:    2000000,
					DefaultMaxTokens: 8192,
					SupportsImages:   true,
				},
				{
					ID:               "gemini-1.5-flash",
					Name:             "Gemini 1.5 Flash",
					CostPer1MIn:      0.075,
					CostPer1MOut:     0.3,
					ContextWindow:    1000000,
					DefaultMaxTokens: 8192,
					SupportsImages:   true,
				},
			},
		},
	}
}
