package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/uglyswap/push/internal/lsp"
)

// LSPTool provides LSP operations for code intelligence.
type LSPTool struct {
	service lsp.LSPService
}

// NewLSPTool creates a new LSP tool.
func NewLSPTool(service lsp.LSPService) *LSPTool {
	return &LSPTool{
		service: service,
	}
}

// Name returns the tool name.
func (t *LSPTool) Name() string {
	return "LSP"
}

// Description returns the tool description.
func (t *LSPTool) Description() string {
	return `Interact with Language Server Protocol (LSP) servers to get code intelligence features.

Supported operations:
- goToDefinition: Find where a symbol is defined
- findReferences: Find all references to a symbol
- hover: Get hover information (documentation, type info) for a symbol
- documentSymbol: Get all symbols (functions, classes, variables) in a document
- workspaceSymbol: Search for symbols across the entire workspace
- goToImplementation: Find implementations of an interface or abstract method
- prepareCallHierarchy: Get call hierarchy item at a position (functions/methods)
- incomingCalls: Find all functions/methods that call the function at a position
- outgoingCalls: Find all functions/methods called by the function at a position

All operations require:
- filePath: The file to operate on
- line: The line number (1-based, as shown in editors)
- character: The character offset (1-based, as shown in editors)

Note: LSP servers must be configured for the file type. If no server is available, an error will be returned.`
}

// LSPParams represents the parameters for LSP operations.
type LSPParams struct {
	Operation lsp.Operation `json:"operation"`
	FilePath  string        `json:"filePath"`
	Line      int           `json:"line"`
	Character int           `json:"character"`
	Query     string        `json:"query,omitempty"` // For workspaceSymbol
}

// Parameters returns the JSON schema for the tool parameters.
func (t *LSPTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The LSP operation to perform",
				"enum": []string{
					"goToDefinition",
					"findReferences",
					"hover",
					"documentSymbol",
					"workspaceSymbol",
					"goToImplementation",
					"prepareCallHierarchy",
					"incomingCalls",
					"outgoingCalls",
				},
			},
			"filePath": map[string]interface{}{
				"type":        "string",
				"description": "The absolute or relative path to the file",
			},
			"line": map[string]interface{}{
				"type":             "integer",
				"description":      "The line number (1-based, as shown in editors)",
				"exclusiveMinimum": 0,
			},
			"character": map[string]interface{}{
				"type":             "integer",
				"description":      "The character offset (1-based, as shown in editors)",
				"exclusiveMinimum": 0,
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query for workspaceSymbol operation",
			},
		},
		"required": []string{"operation", "filePath", "line", "character"},
	}
}

// Execute runs the LSP tool.
func (t *LSPTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p LSPParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Convert 1-based to 0-based
	line := p.Line - 1
	character := p.Character - 1

	if line < 0 {
		return "", fmt.Errorf("line must be at least 1")
	}
	if character < 0 {
		return "", fmt.Errorf("character must be at least 1")
	}

	switch p.Operation {
	case lsp.OperationGoToDefinition:
		return t.goToDefinition(ctx, p.FilePath, line, character)

	case lsp.OperationFindReferences:
		return t.findReferences(ctx, p.FilePath, line, character)

	case lsp.OperationHover:
		return t.hover(ctx, p.FilePath, line, character)

	case lsp.OperationDocumentSymbol:
		return t.documentSymbol(ctx, p.FilePath)

	case lsp.OperationWorkspaceSymbol:
		return t.workspaceSymbol(ctx, p.Query)

	case lsp.OperationGoToImplementation:
		return t.goToImplementation(ctx, p.FilePath, line, character)

	case lsp.OperationPrepareCallHierarchy:
		return t.prepareCallHierarchy(ctx, p.FilePath, line, character)

	case lsp.OperationIncomingCalls:
		return t.incomingCalls(ctx, p.FilePath, line, character)

	case lsp.OperationOutgoingCalls:
		return t.outgoingCalls(ctx, p.FilePath, line, character)

	default:
		return "", fmt.Errorf("unknown operation: %s", p.Operation)
	}
}

func (t *LSPTool) goToDefinition(ctx context.Context, filePath string, line, character int) (string, error) {
	locations, err := t.service.GoToDefinition(ctx, filePath, line, character)
	if err != nil {
		return "", err
	}

	if len(locations) == 0 {
		return "No definition found.", nil
	}

	var result strings.Builder
	result.WriteString("## Definition(s) Found\n\n")
	for _, loc := range locations {
		result.WriteString(fmt.Sprintf("- `%s`\n", lsp.FormatLocation(loc)))
	}
	return result.String(), nil
}

func (t *LSPTool) findReferences(ctx context.Context, filePath string, line, character int) (string, error) {
	locations, err := t.service.FindReferences(ctx, filePath, line, character, true)
	if err != nil {
		return "", err
	}

	if len(locations) == 0 {
		return "No references found.", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## References Found (%d)\n\n", len(locations)))

	// Group by file
	byFile := make(map[string][]lsp.Location)
	for _, loc := range locations {
		byFile[loc.URI] = append(byFile[loc.URI], loc)
	}

	for uri, locs := range byFile {
		result.WriteString(fmt.Sprintf("### %s\n", uri))
		for _, loc := range locs {
			result.WriteString(fmt.Sprintf("- Line %d, Col %d\n", loc.Range.Start.Line+1, loc.Range.Start.Character+1))
		}
		result.WriteString("\n")
	}
	return result.String(), nil
}

func (t *LSPTool) hover(ctx context.Context, filePath string, line, character int) (string, error) {
	hover, err := t.service.Hover(ctx, filePath, line, character)
	if err != nil {
		return "", err
	}

	if hover == nil {
		return "No hover information available.", nil
	}

	var result strings.Builder
	result.WriteString("## Hover Information\n\n")
	result.WriteString(hover.Contents.Value)
	return result.String(), nil
}

func (t *LSPTool) documentSymbol(ctx context.Context, filePath string) (string, error) {
	symbols, err := t.service.DocumentSymbol(ctx, filePath)
	if err != nil {
		return "", err
	}

	if len(symbols) == 0 {
		return "No symbols found in document.", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## Document Symbols (%d)\n\n", countSymbols(symbols)))

	writeSymbols(&result, symbols, 0)
	return result.String(), nil
}

func countSymbols(symbols []lsp.DocumentSymbol) int {
	count := len(symbols)
	for _, s := range symbols {
		count += countSymbols(s.Children)
	}
	return count
}

func writeSymbols(w *strings.Builder, symbols []lsp.DocumentSymbol, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, s := range symbols {
		w.WriteString(fmt.Sprintf("%s- **%s** (%s) - Line %d\n",
			prefix, s.Name, lsp.FormatSymbolKind(s.Kind), s.Range.Start.Line+1))
		if len(s.Children) > 0 {
			writeSymbols(w, s.Children, indent+1)
		}
	}
}

func (t *LSPTool) workspaceSymbol(ctx context.Context, query string) (string, error) {
	symbols, err := t.service.WorkspaceSymbol(ctx, query)
	if err != nil {
		return "", err
	}

	if len(symbols) == 0 {
		return fmt.Sprintf("No symbols found matching '%s'.", query), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## Workspace Symbols matching '%s' (%d)\n\n", query, len(symbols)))

	for _, s := range symbols {
		container := ""
		if s.ContainerName != "" {
			container = fmt.Sprintf(" in %s", s.ContainerName)
		}
		result.WriteString(fmt.Sprintf("- **%s** (%s)%s\n  `%s`\n",
			s.Name, lsp.FormatSymbolKind(s.Kind), container, lsp.FormatLocation(s.Location)))
	}
	return result.String(), nil
}

func (t *LSPTool) goToImplementation(ctx context.Context, filePath string, line, character int) (string, error) {
	locations, err := t.service.GoToImplementation(ctx, filePath, line, character)
	if err != nil {
		return "", err
	}

	if len(locations) == 0 {
		return "No implementations found.", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## Implementation(s) Found (%d)\n\n", len(locations)))
	for _, loc := range locations {
		result.WriteString(fmt.Sprintf("- `%s`\n", lsp.FormatLocation(loc)))
	}
	return result.String(), nil
}

func (t *LSPTool) prepareCallHierarchy(ctx context.Context, filePath string, line, character int) (string, error) {
	items, err := t.service.PrepareCallHierarchy(ctx, filePath, line, character)
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "No call hierarchy available at this position.", nil
	}

	var result strings.Builder
	result.WriteString("## Call Hierarchy Items\n\n")
	for _, item := range items {
		detail := ""
		if item.Detail != "" {
			detail = fmt.Sprintf(" - %s", item.Detail)
		}
		result.WriteString(fmt.Sprintf("- **%s** (%s)%s\n  `%s:%d`\n",
			item.Name, lsp.FormatSymbolKind(item.Kind), detail,
			item.URI, item.Range.Start.Line+1))
	}
	return result.String(), nil
}

func (t *LSPTool) incomingCalls(ctx context.Context, filePath string, line, character int) (string, error) {
	// First prepare call hierarchy to get the item
	items, err := t.service.PrepareCallHierarchy(ctx, filePath, line, character)
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "No call hierarchy available at this position.", nil
	}

	// Get incoming calls for the first item
	calls, err := t.service.IncomingCalls(ctx, items[0])
	if err != nil {
		return "", err
	}

	if len(calls) == 0 {
		return fmt.Sprintf("No incoming calls to '%s'.", items[0].Name), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## Incoming Calls to '%s' (%d)\n\n", items[0].Name, len(calls)))

	for _, call := range calls {
		result.WriteString(fmt.Sprintf("- **%s** (%s)\n  `%s:%d`\n",
			call.From.Name, lsp.FormatSymbolKind(call.From.Kind),
			call.From.URI, call.From.Range.Start.Line+1))
		if len(call.FromRanges) > 1 {
			result.WriteString(fmt.Sprintf("  Called from %d locations\n", len(call.FromRanges)))
		}
	}
	return result.String(), nil
}

func (t *LSPTool) outgoingCalls(ctx context.Context, filePath string, line, character int) (string, error) {
	// First prepare call hierarchy to get the item
	items, err := t.service.PrepareCallHierarchy(ctx, filePath, line, character)
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "No call hierarchy available at this position.", nil
	}

	// Get outgoing calls for the first item
	calls, err := t.service.OutgoingCalls(ctx, items[0])
	if err != nil {
		return "", err
	}

	if len(calls) == 0 {
		return fmt.Sprintf("No outgoing calls from '%s'.", items[0].Name), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## Outgoing Calls from '%s' (%d)\n\n", items[0].Name, len(calls)))

	for _, call := range calls {
		result.WriteString(fmt.Sprintf("- **%s** (%s)\n  `%s:%d`\n",
			call.To.Name, lsp.FormatSymbolKind(call.To.Kind),
			call.To.URI, call.To.Range.Start.Line+1))
		if len(call.FromRanges) > 1 {
			result.WriteString(fmt.Sprintf("  Called %d times\n", len(call.FromRanges)))
		}
	}
	return result.String(), nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *LSPTool) RequiresApproval() bool {
	return false
}
