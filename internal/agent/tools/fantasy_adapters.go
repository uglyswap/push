// Package tools provides fantasy.AgentTool adapters for internal tools.
package tools

import (
	"context"
	"encoding/json"

	"github.com/uglyswap/push/pkg/fantasy"
)

// WebFetchToolAdapter wraps WebFetchTool to implement fantasy.AgentTool.
type WebFetchToolAdapter struct {
	*WebFetchTool
}

// NewWebFetchToolAdapter creates a new adapter that wraps WebFetchTool.
func NewWebFetchToolAdapter(tool *WebFetchTool) *WebFetchToolAdapter {
	return &WebFetchToolAdapter{WebFetchTool: tool}
}

// Execute implements fantasy.AgentTool.
func (a *WebFetchToolAdapter) Execute(ctx context.Context, input string) (fantasy.ToolResultOutput, error) {
	result, err := a.WebFetchTool.Execute(ctx, json.RawMessage(input))
	if err != nil {
		return fantasy.ToolResultOutputContentError{Error: err}, nil
	}
	return fantasy.ToolResultOutputContentText{Text: result}, nil
}

// SetProviderOptions implements fantasy.AgentTool (no-op for this tool).
func (a *WebFetchToolAdapter) SetProviderOptions(opts fantasy.ProviderOptions) {}

// WebSearchToolAdapter wraps WebSearchTool to implement fantasy.AgentTool.
type WebSearchToolAdapter struct {
	*WebSearchTool
}

// NewWebSearchToolAdapter creates a new adapter that wraps WebSearchTool.
func NewWebSearchToolAdapter(tool *WebSearchTool) *WebSearchToolAdapter {
	return &WebSearchToolAdapter{WebSearchTool: tool}
}

// Execute implements fantasy.AgentTool.
func (a *WebSearchToolAdapter) Execute(ctx context.Context, input string) (fantasy.ToolResultOutput, error) {
	result, err := a.WebSearchTool.Execute(ctx, json.RawMessage(input))
	if err != nil {
		return fantasy.ToolResultOutputContentError{Error: err}, nil
	}
	return fantasy.ToolResultOutputContentText{Text: result}, nil
}

// SetProviderOptions implements fantasy.AgentTool (no-op for this tool).
func (a *WebSearchToolAdapter) SetProviderOptions(opts fantasy.ProviderOptions) {}

// Verify adapters implement fantasy.AgentTool
var _ fantasy.AgentTool = (*WebFetchToolAdapter)(nil)
var _ fantasy.AgentTool = (*WebSearchToolAdapter)(nil)
