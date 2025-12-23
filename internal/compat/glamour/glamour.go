// Package glamour provides compatibility shims for github.com/charmbracelet/glamour
// using the public github.com/charmbracelet/glamour package.
package glamour

import (
	"github.com/charmbracelet/glamour"
)

// Re-export core types
type (
	TermRenderer    = glamour.TermRenderer
	TermRendererOption = glamour.TermRendererOption
)

// Re-export functions
var (
	Render         = glamour.Render
	RenderWithEnvironmentConfig = glamour.RenderWithEnvironmentConfig
	NewTermRenderer = glamour.NewTermRenderer
	WithAutoStyle   = glamour.WithAutoStyle
	WithStandardStyle = glamour.WithStandardStyle
	WithStylePath   = glamour.WithStylePath
	WithWordWrap    = glamour.WithWordWrap
	WithEmoji       = glamour.WithEmoji
	WithBaseURL     = glamour.WithBaseURL
	WithPreservedNewLines = glamour.WithPreservedNewLines
)

// Style presets - These are string constants for style names.
// Some versions of glamour may not have all these styles.
const (
	DarkStyle    = "dark"
	LightStyle   = "light"
	DraculaStyle = "dracula"
	NoTTYStyle   = "notty"
	ASCIIStyle   = "ascii"
	PinkStyle    = "pink"
)
