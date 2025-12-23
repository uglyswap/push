// Package spinner provides compatibility shims for github.com/charmbracelet/bubbles/spinner
package spinner

import (
	"github.com/charmbracelet/bubbles/spinner"
)

// Re-export core types
type (
	Model   = spinner.Model
	Spinner = spinner.Spinner
	TickMsg = spinner.TickMsg
)

// Re-export functions
var (
	New  = spinner.New
	Tick = spinner.Tick
)

// Spinner type constants
var (
	Line       = spinner.Line
	Dot        = spinner.Dot
	MiniDot    = spinner.MiniDot
	Jump       = spinner.Jump
	Pulse      = spinner.Pulse
	Points     = spinner.Points
	Globe      = spinner.Globe
	Moon       = spinner.Moon
	Monkey     = spinner.Monkey
	Meter      = spinner.Meter
	Hamburger  = spinner.Hamburger
)
