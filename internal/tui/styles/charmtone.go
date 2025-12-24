package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/uglyswap/push/internal/charmtone"
)

func NewCharmtoneTheme() *Theme {
	t := &Theme{
		Name:   "charmtone",
		IsDark: true,

		Primary:   charmtone.Charple,
		Secondary: charmtone.Dolly,
		Tertiary:  charmtone.Bok,
		Accent:    charmtone.Zest,

		// Backgrounds
		BgBase:        charmtone.Pepper,
		BgBaseLighter: charmtone.BBQ,
		BgSubtle:      charmtone.Charcoal,
		BgOverlay:     charmtone.Iron,

		// Foregrounds
		FgBase:      charmtone.Ash,
		FgMuted:     charmtone.Squid,
		FgHalfMuted: charmtone.Smoke,
		FgSubtle:    charmtone.Oyster,
		FgSelected:  charmtone.Salt,

		// Borders
		Border:      charmtone.Charcoal,
		BorderFocus: charmtone.Charple,

		// Status
		Success: charmtone.Guac,
		Error:   charmtone.Sriracha,
		Warning: charmtone.Zest,
		Info:    charmtone.Malibu,

		// Colors
		White: charmtone.Butter,

		BlueLight: charmtone.Sardine,
		BlueDark:  charmtone.Damson,
		Blue:      charmtone.Malibu,

		Yellow: charmtone.Mustard,
		Citron: charmtone.Citron,

		Green:      charmtone.Julep,
		GreenDark:  charmtone.Guac,
		GreenLight: charmtone.Bok,

		Red:      charmtone.Coral,
		RedDark:  charmtone.Sriracha,
		RedLight: charmtone.Salmon,
		Cherry:   charmtone.Cherry,
	}

	// Text selection.
	t.TextSelection = lipgloss.NewStyle().Foreground(charmtone.Salt.Lipgloss()).Background(charmtone.Charple.Lipgloss())

	// LSP and MCP status.
	t.ItemOfflineIcon = lipgloss.NewStyle().Foreground(charmtone.Squid.Lipgloss()).SetString("●")
	t.ItemBusyIcon = t.ItemOfflineIcon.Foreground(charmtone.Citron.Lipgloss())
	t.ItemErrorIcon = t.ItemOfflineIcon.Foreground(charmtone.Coral.Lipgloss())
	t.ItemOnlineIcon = t.ItemOfflineIcon.Foreground(charmtone.Guac.Lipgloss())

	// Editor: Yolo Mode.
	t.YoloIconFocused = lipgloss.NewStyle().Foreground(charmtone.Oyster.Lipgloss()).Background(charmtone.Citron.Lipgloss()).Bold(true).SetString(" ! ")
	t.YoloIconBlurred = t.YoloIconFocused.Foreground(charmtone.Pepper.Lipgloss()).Background(charmtone.Squid.Lipgloss())
	t.YoloDotsFocused = lipgloss.NewStyle().Foreground(charmtone.Zest.Lipgloss()).SetString(":::")
	t.YoloDotsBlurred = t.YoloDotsFocused.Foreground(charmtone.Squid.Lipgloss())

	// oAuth Chooser.
	t.AuthBorderSelected = lipgloss.NewStyle().BorderForeground(charmtone.Guac.Lipgloss())
	t.AuthTextSelected = lipgloss.NewStyle().Foreground(charmtone.Julep.Lipgloss())
	t.AuthBorderUnselected = lipgloss.NewStyle().BorderForeground(charmtone.Iron.Lipgloss())
	t.AuthTextUnselected = lipgloss.NewStyle().Foreground(charmtone.Squid.Lipgloss())

	return t
}
