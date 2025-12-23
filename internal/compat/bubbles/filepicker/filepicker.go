// Package filepicker provides compatibility shims for github.com/charmbracelet/bubbles/filepicker
package filepicker

import (
	"github.com/charmbracelet/bubbles/filepicker"
)

// Re-export core types
type (
	Model  = filepicker.Model
	KeyMap = filepicker.KeyMap
	Styles = filepicker.Styles
)

// Re-export functions
var (
	New = filepicker.New
)

// GetHighlightedPath returns the currently highlighted path in the filepicker.
// This is a compatibility helper for filepicker.Model.Path which is a field
// in the standard bubbles/filepicker package.
func GetHighlightedPath(fp *filepicker.Model) string {
	// In standard bubbles filepicker, Path is a field containing the selected path
	return fp.Path
}
