package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SearchProvider defines the interface for web search providers.
type SearchProvider interface {
	Search(ctx context.Context, query string, options SearchOptions) ([]SearchResult, error)
}

// SearchOptions contains options for search.
type SearchOptions struct {
	AllowedDomains  []string `json:"allowed_domains,omitempty"`
	BlockedDomains  []string `json:"blocked_domains,omitempty"`
	MaxResults      int      `json:"max_results,omitempty"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
}

// WebSearchTool performs web searches.
type WebSearchTool struct {
	provider SearchProvider
}

// NewWebSearchTool creates a new WebSearch tool.
func NewWebSearchTool(provider SearchProvider) *WebSearchTool {
	return &WebSearchTool{
		provider: provider,
	}
}

// NewWebSearchToolWithHTTPClient creates a WebSearch tool with an HTTP client.
// It wraps the http.Client with a DuckDuckGo provider.
func NewWebSearchToolWithHTTPClient(client *http.Client) *WebSearchTool {
	if client == nil {
		client = http.DefaultClient
	}
	return NewWebSearchTool(NewDuckDuckGoProvider(&httpClientAdapter{client: client}))
}

// httpClientAdapter wraps http.Client to implement HTTPClient interface.
type httpClientAdapter struct {
	client *http.Client
}

func (a *httpClientAdapter) Get(url string) ([]byte, error) {
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Name returns the tool name.
func (t *WebSearchTool) Name() string {
	return "WebSearch"
}

// Description returns the tool description.
func (t *WebSearchTool) Description() string {
	return `Allows searching the web and using the results to inform responses.

Usage:
- Provides up-to-date information for current events and recent data
- Returns search result information formatted as search result blocks, including links as markdown hyperlinks
- Use this tool for accessing information beyond knowledge cutoff
- Searches are performed automatically within a single API call

CRITICAL REQUIREMENT:
- After answering the user's question, you MUST include a "Sources:" section at the end
- In the Sources section, list all relevant URLs from the search results as markdown hyperlinks
- Example format:

  [Your answer here]

  Sources:
  - [Source Title 1](https://example.com/1)
  - [Source Title 2](https://example.com/2)

Usage notes:
- Domain filtering is supported to include or block specific websites
- Use the correct year in search queries for recent information`
}

// WebSearchParams represents the parameters for WebSearch.
type WebSearchParams struct {
	Query          string   `json:"query"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	BlockedDomains []string `json:"blocked_domains,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *WebSearchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query to use",
				"minLength":   2,
			},
			"allowed_domains": map[string]interface{}{
				"type":        "array",
				"description": "Only include search results from these domains",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"blocked_domains": map[string]interface{}{
				"type":        "array",
				"description": "Never include search results from these domains",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"query"},
	}
}

// Execute runs the WebSearch tool.
func (t *WebSearchTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p WebSearchParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if len(p.Query) < 2 {
		return "", fmt.Errorf("query must be at least 2 characters")
	}

	if t.provider == nil {
		return t.fallbackResponse(p.Query), nil
	}

	options := SearchOptions{
		AllowedDomains: p.AllowedDomains,
		BlockedDomains: p.BlockedDomains,
		MaxResults:     10,
	}

	results, err := t.provider.Search(ctx, p.Query, options)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	return t.formatResults(p.Query, results), nil
}

func (t *WebSearchTool) formatResults(query string, results []SearchResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No search results found for: %s", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Search Results for: %s\n\n", query))
	sb.WriteString(fmt.Sprintf("Found %d results:\n\n", len(results)))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("### %d. [%s](%s)\n", i+1, r.Title, r.URL))
		if r.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n", r.Description))
		}
		if r.Source != "" {
			sb.WriteString(fmt.Sprintf("*Source: %s*\n", r.Source))
		}
		sb.WriteString("\n")
	}

	// Add sources section reminder
	sb.WriteString("---\n\n")
	sb.WriteString("**Remember**: Include a Sources section in your response with the relevant URLs.\n")

	return sb.String()
}

func (t *WebSearchTool) fallbackResponse(query string) string {
	return fmt.Sprintf(`## Search Provider Not Configured

The web search tool requires a search provider to be configured.

To enable web search, you can:
1. Configure a search API (e.g., DuckDuckGo, Brave Search, SerpAPI)
2. Set the appropriate API keys in your configuration

**Query attempted**: %s

For now, please use the WebFetch tool to fetch specific URLs if you know them, or ask the user to provide URLs directly.`, query)
}

// RequiresApproval returns whether this tool requires user approval.
func (t *WebSearchTool) RequiresApproval() bool {
	return false
}

// DuckDuckGoProvider implements a basic DuckDuckGo search.
type DuckDuckGoProvider struct {
	httpClient HTTPClient
}

// HTTPClient interface for making HTTP requests.
type HTTPClient interface {
	Get(url string) ([]byte, error)
}

// NewDuckDuckGoProvider creates a new DuckDuckGo provider.
func NewDuckDuckGoProvider(client HTTPClient) *DuckDuckGoProvider {
	return &DuckDuckGoProvider{
		httpClient: client,
	}
}

// Search performs a DuckDuckGo search.
func (p *DuckDuckGoProvider) Search(ctx context.Context, query string, options SearchOptions) ([]SearchResult, error) {
	// Build search URL
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_redirect=1&no_html=1",
		url.QueryEscape(query))

	if len(options.AllowedDomains) > 0 {
		searchURL += "&site:" + strings.Join(options.AllowedDomains, "+OR+site:")
	}

	data, err := p.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search results: %w", err)
	}

	// Parse DuckDuckGo response
	var ddgResponse struct {
		Abstract       string `json:"Abstract"`
		AbstractURL    string `json:"AbstractURL"`
		AbstractSource string `json:"AbstractSource"`
		RelatedTopics  []struct {
			Text      string `json:"Text"`
			FirstURL  string `json:"FirstURL"`
		} `json:"RelatedTopics"`
		Results []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(data, &ddgResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	var results []SearchResult

	// Add abstract if present
	if ddgResponse.Abstract != "" && ddgResponse.AbstractURL != "" {
		results = append(results, SearchResult{
			Title:       ddgResponse.AbstractSource,
			URL:         ddgResponse.AbstractURL,
			Description: ddgResponse.Abstract,
			Source:      "DuckDuckGo",
		})
	}

	// Add direct results
	for _, r := range ddgResponse.Results {
		if r.FirstURL == "" {
			continue
		}
		results = append(results, SearchResult{
			Title:       extractTitle(r.Text),
			URL:         r.FirstURL,
			Description: r.Text,
			Source:      "DuckDuckGo",
		})
	}

	// Add related topics
	for _, r := range ddgResponse.RelatedTopics {
		if r.FirstURL == "" {
			continue
		}
		results = append(results, SearchResult{
			Title:       extractTitle(r.Text),
			URL:         r.FirstURL,
			Description: r.Text,
			Source:      "DuckDuckGo",
		})
	}

	// Apply domain filtering
	if len(options.BlockedDomains) > 0 {
		var filtered []SearchResult
		for _, r := range results {
			blocked := false
			for _, domain := range options.BlockedDomains {
				if strings.Contains(r.URL, domain) {
					blocked = true
					break
				}
			}
			if !blocked {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	// Limit results
	if options.MaxResults > 0 && len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results, nil
}

func extractTitle(text string) string {
	// Extract title from text (usually before the first dash or period)
	if idx := strings.Index(text, " - "); idx > 0 && idx < 100 {
		return strings.TrimSpace(text[:idx])
	}
	if idx := strings.Index(text, ". "); idx > 0 && idx < 100 {
		return strings.TrimSpace(text[:idx])
	}
	if len(text) > 80 {
		return strings.TrimSpace(text[:80]) + "..."
	}
	return text
}
