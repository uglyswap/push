package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// NotebookEditTool edits Jupyter notebook cells.
type NotebookEditTool struct{}

// NewNotebookEditTool creates a new NotebookEdit tool.
func NewNotebookEditTool() *NotebookEditTool {
	return &NotebookEditTool{}
}

// Name returns the tool name.
func (t *NotebookEditTool) Name() string {
	return "NotebookEdit"
}

// Description returns the tool description.
func (t *NotebookEditTool) Description() string {
	return `Completely replaces the contents of a specific cell in a Jupyter notebook (.ipynb file) with new source.

Jupyter notebooks are interactive documents that combine code, text, and visualizations, commonly used for data analysis and scientific computing.

The notebook_path parameter must be an absolute path, not a relative path. The cell_number is 0-indexed.

Use edit_mode=insert to add a new cell at the index specified by cell_number.
Use edit_mode=delete to delete the cell at the index specified by cell_number.`
}

// NotebookEditParams represents the parameters for NotebookEdit.
type NotebookEditParams struct {
	NotebookPath string `json:"notebook_path"`
	CellID       string `json:"cell_id,omitempty"`
	CellType     string `json:"cell_type,omitempty"` // code or markdown
	EditMode     string `json:"edit_mode,omitempty"` // replace, insert, delete
	NewSource    string `json:"new_source"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *NotebookEditTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"notebook_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the Jupyter notebook file to edit (must be absolute, not relative)",
			},
			"cell_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the cell to edit. When inserting, the new cell will be inserted after this cell.",
			},
			"cell_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"code", "markdown"},
				"description": "The type of the cell (code or markdown). Required for insert mode.",
			},
			"edit_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"replace", "insert", "delete"},
				"description": "The type of edit to make. Defaults to replace.",
			},
			"new_source": map[string]interface{}{
				"type":        "string",
				"description": "The new source for the cell",
			},
		},
		"required": []string{"notebook_path", "new_source"},
	}
}

// NotebookCell represents a cell in a Jupyter notebook.
type NotebookCell struct {
	CellType       string            `json:"cell_type"`
	Source         []string          `json:"source"`
	Metadata       map[string]interface{} `json:"metadata"`
	ID             string            `json:"id,omitempty"`
	ExecutionCount *int              `json:"execution_count,omitempty"`
	Outputs        []interface{}     `json:"outputs,omitempty"`
}

// Notebook represents a Jupyter notebook.
type Notebook struct {
	Cells         []NotebookCell         `json:"cells"`
	Metadata      map[string]interface{} `json:"metadata"`
	Nbformat      int                    `json:"nbformat"`
	NbformatMinor int                    `json:"nbformat_minor"`
}

// Execute runs the NotebookEdit tool.
func (t *NotebookEditTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p NotebookEditParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if p.NotebookPath == "" {
		return "", fmt.Errorf("notebook_path is required")
	}

	// Default edit mode
	if p.EditMode == "" {
		p.EditMode = "replace"
	}

	// Read notebook
	data, err := os.ReadFile(p.NotebookPath)
	if err != nil {
		return "", fmt.Errorf("failed to read notebook: %w", err)
	}

	var notebook Notebook
	if err := json.Unmarshal(data, &notebook); err != nil {
		return "", fmt.Errorf("failed to parse notebook: %w", err)
	}

	// Find cell by ID or index
	cellIndex := -1
	if p.CellID != "" {
		for i, cell := range notebook.Cells {
			if cell.ID == p.CellID {
				cellIndex = i
				break
			}
		}
		if cellIndex == -1 && p.EditMode != "insert" {
			return "", fmt.Errorf("cell with ID '%s' not found", p.CellID)
		}
	}

	// Convert source string to lines
	sourceLines := splitSource(p.NewSource)

	switch p.EditMode {
	case "replace":
		if cellIndex == -1 {
			return "", fmt.Errorf("cell_id is required for replace mode")
		}
		notebook.Cells[cellIndex].Source = sourceLines
		if p.CellType != "" {
			notebook.Cells[cellIndex].CellType = p.CellType
		}

	case "insert":
		if p.CellType == "" {
			p.CellType = "code"
		}
		newCell := NotebookCell{
			CellType: p.CellType,
			Source:   sourceLines,
			Metadata: make(map[string]interface{}),
			ID:       generateCellID(),
		}
		if p.CellType == "code" {
			newCell.Outputs = []interface{}{}
		}
		
		// Insert after specified cell or at beginning
		insertIndex := cellIndex + 1
		if cellIndex == -1 {
			insertIndex = 0
		}
		
		// Insert cell
		notebook.Cells = append(
			notebook.Cells[:insertIndex],
			append([]NotebookCell{newCell}, notebook.Cells[insertIndex:]...)...,
		)

	case "delete":
		if cellIndex == -1 {
			return "", fmt.Errorf("cell_id is required for delete mode")
		}
		notebook.Cells = append(notebook.Cells[:cellIndex], notebook.Cells[cellIndex+1:]...)

	default:
		return "", fmt.Errorf("unknown edit_mode: %s", p.EditMode)
	}

	// Write notebook back
	output, err := json.MarshalIndent(notebook, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal notebook: %w", err)
	}

	if err := os.WriteFile(p.NotebookPath, output, 0644); err != nil {
		return "", fmt.Errorf("failed to write notebook: %w", err)
	}

	return fmt.Sprintf("Successfully %sed cell in %s", p.EditMode, p.NotebookPath), nil
}

// splitSource splits source content into lines for notebook format.
func splitSource(source string) []string {
	if source == "" {
		return []string{}
	}

	lines := strings.Split(source, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		if i < len(lines)-1 {
			result[i] = line + "\n"
		} else {
			result[i] = line
		}
	}

	return result
}

// generateCellID generates a simple cell ID.
func generateCellID() string {
	// Generate a simple random-ish ID
	return fmt.Sprintf("cell-%d", randInt())
}

var cellIDCounter = 0

func randInt() int {
	cellIDCounter++
	return cellIDCounter
}

// RequiresApproval returns whether this tool requires user approval.
func (t *NotebookEditTool) RequiresApproval() bool {
	return false
}
