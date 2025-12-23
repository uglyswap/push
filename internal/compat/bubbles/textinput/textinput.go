// Package textinput provides compatibility shims for github.com/charmbracelet/bubbles/textinput
package textinput

import (
	"image/color"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Re-export core types
type (
	Model      = textinput.Model
	EchoMode   = textinput.EchoMode
	CursorMode = textinput.CursorMode
)

// Re-export functions
var (
	New   = textinput.New
	Blink = textinput.Blink
)

// Echo mode constants
const (
	EchoNormal   = textinput.EchoNormal
	EchoPassword = textinput.EchoPassword
	EchoNone     = textinput.EchoNone
)

// Cursor mode constants
const (
	CursorBlink  = textinput.CursorBlink
	CursorStatic = textinput.CursorStatic
	CursorHide   = textinput.CursorHide
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
	Text        lipgloss.Style
	Placeholder lipgloss.Style
	Prompt      lipgloss.Style
	Suggestion  lipgloss.Style
}

// CursorStyle holds cursor styling options.
type CursorStyle struct {
	Color color.Color
	Shape CursorShape
	Blink bool
}

// Styles holds all style configurations for the textinput component.
type Styles struct {
	Focused StyleState
	Blurred StyleState
	Cursor  CursorStyle
}

// ExtendedModel wraps textinput.Model with additional methods.
type ExtendedModel struct {
	textinput.Model
	virtualCursor bool
	styles        *Styles
}

// NewExtended creates a new ExtendedModel.
func NewExtended() ExtendedModel {
	return ExtendedModel{Model: textinput.New()}
}

// SetVirtualCursor sets whether to use a virtual cursor.
func (m *ExtendedModel) SetVirtualCursor(v bool) {
	m.virtualCursor = v
}

// SetStyles sets the styles for the textinput.
func (m *ExtendedModel) SetStyles(s Styles) {
	m.styles = &s
	// Apply what we can to the underlying model
	// The standard textinput doesn't have full style support
}

// SetWidth sets the width of the textinput.
func (m *ExtendedModel) SetWidth(w int) {
	m.Model.Width = w
}

// SetVirtualCursorOnModel sets virtual cursor on a standard Model.
// This is a no-op since standard textinput doesn't support it.
func SetVirtualCursorOnModel(m *textinput.Model, v bool) {
	// No-op for standard textinput
}

// SetStylesOnModel sets styles on a standard Model.
// This is a no-op since standard textinput doesn't support custom Styles type.
func SetStylesOnModel(m *textinput.Model, s Styles) {
	// Apply what we can - standard textinput has some style fields
	m.PromptStyle = s.Focused.Prompt
	m.TextStyle = s.Focused.Text
	m.PlaceholderStyle = s.Focused.Placeholder
}

// SetWidthOnModel sets width on a standard Model.
func SetWidthOnModel(m *textinput.Model, w int) {
	m.Width = w
}

// Cursor represents a cursor position.
type Cursor struct {
	X int
	Y int
}

// GetCursorPosition returns the cursor position for a textinput.Model.
// Returns nil if the input is not focused.
func GetCursorPosition(m *textinput.Model) *Cursor {
	if !m.Focused() {
		return nil
	}
	return &Cursor{
		X: len(m.Prompt) + m.Position(),
		Y: 0, // textinput is single line
	}
}
