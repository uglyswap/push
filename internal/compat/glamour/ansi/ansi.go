// Package ansi provides compatibility shims for github.com/charmbracelet/glamour/ansi
// using the public github.com/charmbracelet/glamour/ansi package.
package ansi

import (
	"github.com/charmbracelet/glamour/ansi"
)

// Re-export core types
type (
	StyleConfig = ansi.StyleConfig
	StyleBlock  = ansi.StyleBlock
	StylePrimitive = ansi.StylePrimitive
	StyleTask   = ansi.StyleTask
	StyleCodeBlock = ansi.StyleCodeBlock
	StyleList   = ansi.StyleList
	StyleTable  = ansi.StyleTable
)

// Re-export functions
var (
	NewRenderer = ansi.NewRenderer
)

// DefaultStyles provides default style configurations.
// This is a stub for compatibility - the actual implementation may vary by glamour version.
var DefaultStyles = map[string]StyleConfig{}
