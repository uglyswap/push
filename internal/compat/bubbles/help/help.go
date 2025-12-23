// Package help provides compatibility shims for github.com/charmbracelet/bubbles/help
package help

import (
	"github.com/charmbracelet/bubbles/help"
)

// Re-export core types
type (
	Model  = help.Model
	KeyMap = help.KeyMap
	Styles = help.Styles
)

// Re-export functions
var (
	New = help.New
)
