// Package progress provides compatibility shims for github.com/charmbracelet/bubbles/progress
package progress

import (
	"github.com/charmbracelet/bubbles/progress"
)

// Re-export core types
type (
	Model      = progress.Model
	Option     = progress.Option
	FrameMsg   = progress.FrameMsg
)

// Re-export functions
var (
	New               = progress.New
	WithWidth         = progress.WithWidth
	WithGradient      = progress.WithGradient
	WithDefaultGradient = progress.WithDefaultGradient
	WithSolidFill     = progress.WithSolidFill
	WithDefaultScaledGradient = progress.WithDefaultScaledGradient
	WithScaledGradient = progress.WithScaledGradient
	WithoutPercentage = progress.WithoutPercentage
)
