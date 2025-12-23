// Package key provides compatibility shims for github.com/charmbracelet/bubbles/key
// using the public github.com/charmbracelet/bubbles/key package.
package key

import (
	"github.com/charmbracelet/bubbles/key"
)

// Re-export core types
type (
	Binding = key.Binding
)

// Re-export functions
var (
	NewBinding   = key.NewBinding
	WithKeys     = key.WithKeys
	WithHelp     = key.WithHelp
	WithDisabled = key.WithDisabled
)

// Matches wraps key.Matches - checks if the given key message matches any of the bindings.
func Matches(msg interface{ String() string }, bindings ...Binding) bool {
	return key.Matches(msg, bindings...)
}
