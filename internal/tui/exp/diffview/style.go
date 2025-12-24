package diffview

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/uglyswap/push/internal/charmtone"
)

// LineStyle defines the styles for a given line type in the diff view.
type LineStyle struct {
	LineNumber lipgloss.Style
	Symbol     lipgloss.Style
	Code       lipgloss.Style
}

// Style defines the overall style for the diff view, including styles for
// different line types such as divider, missing, equal, insert, and delete
// lines.
type Style struct {
	DividerLine LineStyle
	MissingLine LineStyle
	EqualLine   LineStyle
	InsertLine  LineStyle
	DeleteLine  LineStyle
}

// DefaultLightStyle provides a default light theme style for the diff view.
func DefaultLightStyle() Style {
	return Style{
		DividerLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Iron.Lipgloss()).
				Background(charmtone.Thunder.Lipgloss()),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Oyster.Lipgloss()).
				Background(charmtone.Anchovy.Lipgloss()),
		},
		MissingLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Background(charmtone.Ash.Lipgloss()),
			Code: lipgloss.NewStyle().
				Background(charmtone.Ash.Lipgloss()),
		},
		EqualLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Charcoal.Lipgloss()).
				Background(charmtone.Ash.Lipgloss()),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Pepper.Lipgloss()).
				Background(charmtone.Salt.Lipgloss()),
		},
		InsertLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Turtle.Lipgloss()).
				Background(lipgloss.Color("#c8e6c9")),
			Symbol: lipgloss.NewStyle().
				Foreground(charmtone.Turtle.Lipgloss()).
				Background(lipgloss.Color("#e8f5e9")),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Pepper.Lipgloss()).
				Background(lipgloss.Color("#e8f5e9")),
		},
		DeleteLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Cherry.Lipgloss()).
				Background(lipgloss.Color("#ffcdd2")),
			Symbol: lipgloss.NewStyle().
				Foreground(charmtone.Cherry.Lipgloss()).
				Background(lipgloss.Color("#ffebee")),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Pepper.Lipgloss()).
				Background(lipgloss.Color("#ffebee")),
		},
	}
}

// DefaultDarkStyle provides a default dark theme style for the diff view.
func DefaultDarkStyle() Style {
	return Style{
		DividerLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Smoke.Lipgloss()).
				Background(charmtone.Sapphire.Lipgloss()),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Smoke.Lipgloss()).
				Background(charmtone.Ox.Lipgloss()),
		},
		MissingLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Background(charmtone.Charcoal.Lipgloss()),
			Code: lipgloss.NewStyle().
				Background(charmtone.Charcoal.Lipgloss()),
		},
		EqualLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Ash.Lipgloss()).
				Background(charmtone.Charcoal.Lipgloss()),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Salt.Lipgloss()).
				Background(charmtone.Pepper.Lipgloss()),
		},
		InsertLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Turtle.Lipgloss()).
				Background(lipgloss.Color("#293229")),
			Symbol: lipgloss.NewStyle().
				Foreground(charmtone.Turtle.Lipgloss()).
				Background(lipgloss.Color("#303a30")),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Salt.Lipgloss()).
				Background(lipgloss.Color("#303a30")),
		},
		DeleteLine: LineStyle{
			LineNumber: lipgloss.NewStyle().
				Foreground(charmtone.Cherry.Lipgloss()).
				Background(lipgloss.Color("#332929")),
			Symbol: lipgloss.NewStyle().
				Foreground(charmtone.Cherry.Lipgloss()).
				Background(lipgloss.Color("#3a3030")),
			Code: lipgloss.NewStyle().
				Foreground(charmtone.Salt.Lipgloss()).
				Background(lipgloss.Color("#3a3030")),
		},
	}
}
