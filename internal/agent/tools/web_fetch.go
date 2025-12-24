package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// WebFetchCache stores cached fetch results.
type WebFetchCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

type cacheEntry struct {
	content   string
	fetchedAt time.Time
}

const cacheTTL = 15 * time.Minute

// NewWebFetchCache creates a new cache.
func NewWebFetchCache() *WebFetchCache {
	return &WebFetchCache{
		entries: make(map[string]*cacheEntry),
	}
}

// Get retrieves a cached entry if it exists and is not expired.
func (c *WebFetchCache) Get(url string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[url]
	if !ok {
		return "", false
	}

	if time.Since(entry.fetchedAt) > cacheTTL {
		return "", false
	}

	return entry.content, true
}

// Set stores a cache entry.
func (c *WebFetchCache) Set(url, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[url] = &cacheEntry{
		content:   content,
		fetchedAt: time.Now(),
	}
}

// Clean removes expired entries.
func (c *WebFetchCache) Clean() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for url, entry := range c.entries {
		if now.Sub(entry.fetchedAt) > cacheTTL {
			delete(c.entries, url)
		}
	}
}

// WebFetchTool fetches and processes web content.
type WebFetchTool struct {
	httpClient *http.Client
	cache      *WebFetchCache
}

// NewWebFetchTool creates a new WebFetch tool.
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 10 redirects
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		cache: NewWebFetchCache(),
	}
}

// NewWebFetchToolWithClient creates a new WebFetch tool with custom directory and HTTP client.
// The tmpDir parameter is accepted for API compatibility but not used (kept for future use).
func NewWebFetchToolWithClient(tmpDir string, client *http.Client) *WebFetchTool {
	t := NewWebFetchTool()
	if client != nil {
		t.httpClient = client
	}
	return t
}

// Name returns the tool name.
func (t *WebFetchTool) Name() string {
	return "WebFetch"
}

// Description returns the tool description.
func (t *WebFetchTool) Description() string {
	return `Fetches content from a specified URL and processes it.

Usage:
- Takes a URL and a prompt as input
- Fetches the URL content, converts HTML to markdown
- Processes the content with the prompt using a small, fast model
- Returns the model's response about the content
- Use this tool when you need to retrieve and analyze web content

Usage notes:
- The URL must be a fully-formed valid URL
- HTTP URLs will be automatically upgraded to HTTPS
- The prompt should describe what information you want to extract from the page
- Results may be summarized if the content is very large
- Includes a self-cleaning 15-minute cache for faster responses
- When a URL redirects to a different host, the tool will inform you`
}

// WebFetchToolParams represents the parameters for WebFetch tool with prompt.
type WebFetchToolParams struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *WebFetchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to fetch content from",
				"format":      "uri",
			},
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "The prompt to run on the fetched content",
			},
		},
		"required": []string{"url", "prompt"},
	}
}

// Execute runs the WebFetch tool.
func (t *WebFetchTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p WebFetchToolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if p.URL == "" {
		return "", fmt.Errorf("url is required")
	}
	if p.Prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}

	// Validate and normalize URL
	parsedURL, err := url.Parse(p.URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Upgrade HTTP to HTTPS
	if parsedURL.Scheme == "http" {
		parsedURL.Scheme = "https"
	}

	normalizedURL := parsedURL.String()

	// Check cache
	if cached, ok := t.cache.Get(normalizedURL); ok {
		return t.processContent(cached, p.Prompt), nil
	}

	// Fetch content
	req, err := http.NewRequestWithContext(ctx, "GET", normalizedURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "Crush/1.0 (AI Assistant; +https://github.com/uglyswap/push)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check for redirects to different host
	finalURL := resp.Request.URL
	if finalURL.Host != parsedURL.Host {
		return fmt.Sprintf("The URL redirected to a different host: %s\n\nPlease make a new WebFetch request with this URL to fetch the content.", finalURL.String()), nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Read body with limit
	limitedReader := io.LimitReader(resp.Body, 5*1024*1024) // 5MB limit
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Convert to markdown
	contentType := resp.Header.Get("Content-Type")
	var content string

	if strings.Contains(contentType, "text/html") {
		content = t.htmlToMarkdown(string(body))
	} else if strings.Contains(contentType, "text/plain") {
		content = string(body)
	} else if strings.Contains(contentType, "application/json") {
		content = "```json\n" + string(body) + "\n```"
	} else {
		content = string(body)
	}

	// Cache the content
	t.cache.Set(normalizedURL, content)

	// Clean expired cache entries periodically
	go t.cache.Clean()

	return t.processContent(content, p.Prompt), nil
}

// processContent processes the fetched content with the prompt.
func (t *WebFetchTool) processContent(content, prompt string) string {
	// Truncate if too long
	const maxContentLen = 50000
	if len(content) > maxContentLen {
		content = content[:maxContentLen] + "\n\n... (content truncated)"
	}

	return fmt.Sprintf("## Fetched Content\n\n%s\n\n---\n\n**User Prompt**: %s\n\nPlease analyze the above content according to the prompt.", content, prompt)
}

// htmlToMarkdown converts HTML to a simplified markdown format.
func (t *WebFetchTool) htmlToMarkdown(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// Fallback: strip tags with regex
		return t.stripHTMLTags(htmlContent)
	}

	var result strings.Builder
	t.extractText(doc, &result, 0)

	// Clean up the result
	text := result.String()

	// Remove excessive whitespace
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

func (t *WebFetchTool) extractText(n *html.Node, w *strings.Builder, depth int) {
	// Skip certain elements
	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style", "noscript", "iframe", "svg":
			return
		case "h1":
			w.WriteString("\n# ")
		case "h2":
			w.WriteString("\n## ")
		case "h3":
			w.WriteString("\n### ")
		case "h4":
			w.WriteString("\n#### ")
		case "h5":
			w.WriteString("\n##### ")
		case "h6":
			w.WriteString("\n###### ")
		case "p", "div", "article", "section":
			w.WriteString("\n")
		case "br":
			w.WriteString("\n")
		case "li":
			w.WriteString("\n- ")
		case "a":
			// Extract href for links
			for _, attr := range n.Attr {
				if attr.Key == "href" && !strings.HasPrefix(attr.Val, "#") && !strings.HasPrefix(attr.Val, "javascript:") {
					w.WriteString("[")
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						t.extractText(c, w, depth+1)
					}
					w.WriteString("](" + attr.Val + ")")
					return
				}
			}
		case "code":
			w.WriteString("`")
		case "pre":
			w.WriteString("\n```\n")
		case "strong", "b":
			w.WriteString("**")
		case "em", "i":
			w.WriteString("*")
		}
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			w.WriteString(text)
			w.WriteString(" ")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.extractText(c, w, depth+1)
	}

	// Closing tags
	if n.Type == html.ElementNode {
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			w.WriteString("\n")
		case "p", "div", "article", "section":
			w.WriteString("\n")
		case "code":
			w.WriteString("`")
		case "pre":
			w.WriteString("\n```\n")
		case "strong", "b":
			w.WriteString("**")
		case "em", "i":
			w.WriteString("*")
		}
	}
}

func (t *WebFetchTool) stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// RequiresApproval returns whether this tool requires user approval.
func (t *WebFetchTool) RequiresApproval() bool {
	return false
}
