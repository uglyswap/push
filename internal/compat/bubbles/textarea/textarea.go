// Package textarea provides compatibility shims for github.com/charmbracelet/bubbles/textarea
package textarea

import (
	"image/color"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// Re-export core types
type (
	Model  = textarea.Model
	KeyMap = textarea.KeyMap
)

// Re-export functions
var (
	New           = textarea.New
	DefaultKeyMap = textarea.DefaultKeyMap
)

// CursorShape represents the shape of the cursor.
type CursorShape int

const (
	CursorBlock CursorShape = iota
	CursorLine
	CursorUnderline
)

// StyleState holds style configuration for a specific state (focused/blurred).
type StyleState struct {
	Base             lipgloss.Style
	Text             lipgloss.Style
	LineNumber       lipgloss.Style
	CursorLine       lipgloss.Style
	CursorLineNumber lipgloss.Style
	Placeholder      lipgloss.Style
	Prompt           lipgloss.Style
}

// CursorStyle holds cursor styling options.
type CursorStyle struct {
	Color color.Color
	Shape CursorShape
	Blink bool
}

// Styles holds all style configurations for the textarea component.
type Styles struct {
	Focused StyleState
	Blurred StyleState
	Cursor  CursorStyle
}
