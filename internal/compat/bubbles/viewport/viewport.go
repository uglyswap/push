// Package viewport provides compatibility shims for github.com/charmbracelet/bubbles/viewport
package viewport

import (
	"github.com/charmbracelet/bubbles/viewport"
)

// Re-export core types
type (
	Model  = viewport.Model
	KeyMap = viewport.KeyMap
)

// Re-export functions
var (
	New = viewport.New
	DefaultKeyMap = viewport.DefaultKeyMap
)
