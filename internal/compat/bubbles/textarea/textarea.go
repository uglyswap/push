// Package textarea provides compatibility shims for github.com/charmbracelet/bubbles/textarea
package textarea

import (
	"image/color"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	compat_tea "github.com/uglyswap/crush/internal/compat/bubbletea"
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

// PromptInfo contains information about a line for the prompt function.
// This provides v2 compatibility.
type PromptInfo struct {
	LineNumber int
	Focused    bool
}

// GetWord returns the word at the current cursor position in the textarea.
// Returns empty string if there is no word at the cursor.
func GetWord(m *textarea.Model) string {
	value := m.Value()
	if len(value) == 0 {
		return ""
	}

	// Get cursor position - in v1, we use Line() and LineInfo()
	line := m.Line()
	lines := strings.Split(value, "\n")
	if line >= len(lines) {
		return ""
	}

	currentLine := lines[line]
	info := m.LineInfo()
	col := info.CharOffset

	if col > len(currentLine) {
		col = len(currentLine)
	}

	// Find word boundaries
	start := col
	for start > 0 && !unicode.IsSpace(rune(currentLine[start-1])) {
		start--
	}

	end := col
	for end < len(currentLine) && !unicode.IsSpace(rune(currentLine[end])) {
		end++
	}

	return currentLine[start:end]
}

// MoveToEnd moves the cursor to the end of the textarea content.
// In v1, we do this by setting the cursor position.
func MoveToEnd(m *textarea.Model) {
	value := m.Value()
	lines := strings.Split(value, "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		m.SetCursor(len(lastLine))
		// Move to last line by going down from current position
		for i := m.Line(); i < len(lines)-1; i++ {
			m.CursorDown()
		}
		m.CursorEnd()
	}
}

// GetCursorPosition returns the cursor position for a textarea.Model.
// Returns nil if the textarea is not focused.
func GetCursorPosition(m *textarea.Model) *compat_tea.Cursor {
	if !m.Focused() {
		return nil
	}
	info := m.LineInfo()
	return &compat_tea.Cursor{
		X: info.CharOffset + len(m.Prompt),
		Y: m.Line(),
	}
}

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

// SetStylesOnModel sets styles on a textarea model.
// In v1, textarea.Model uses FocusedStyle and BlurredStyle fields instead of SetStyles method.
// This function accepts our compat Styles type and converts it to the v1 Style type.
func SetStylesOnModel(m *textarea.Model, styles Styles) {
	m.FocusedStyle = convertStyleState(styles.Focused)
	m.BlurredStyle = convertStyleState(styles.Blurred)
}

// convertStyleState converts our compat StyleState to bubbles textarea.Style
func convertStyleState(ss StyleState) textarea.Style {
	return textarea.Style{
		Base:             ss.Base,
		Text:             ss.Text,
		LineNumber:       ss.LineNumber,
		CursorLine:       ss.CursorLine,
		CursorLineNumber: ss.CursorLineNumber,
		Placeholder:      ss.Placeholder,
		Prompt:           ss.Prompt,
		EndOfBuffer:      lipgloss.Style{}, // v1 has this field, v2 might not use it
	}
}

// PromptFuncV2 is the v2-style prompt function signature.
type PromptFuncV2 func(info PromptInfo) string

// SetPromptFuncCompat sets a v2-style prompt function on a textarea model.
// It adapts the v2 function signature to v1's simpler (lineIdx int) -> string signature.
func SetPromptFuncCompat(m *textarea.Model, width int, fn PromptFuncV2) {
	m.SetPromptFunc(width, func(lineIdx int) string {
		return fn(PromptInfo{
			LineNumber: lineIdx,
			Focused:    m.Focused(),
		})
	})
}

// SetVirtualCursorCompat is a no-op for v1 compatibility.
// v1 doesn't have SetVirtualCursor method.
func SetVirtualCursorCompat(m *textarea.Model, enabled bool) {
	// No-op in v1 - virtual cursor is not supported
}
