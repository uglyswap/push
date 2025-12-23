// Package charmtone provides color palette constants.
// This is a stub replacing github.com/charmbracelet/x/exp/charmtone.
package charmtone

import (
	"image/color"

	"github.com/charmbracelet/lipgloss"
)

// Color represents a color that can be used with lipgloss and provides hex string access.
type Color struct {
	hex string
}

// NewColor creates a new Color from a hex string.
func NewColor(hex string) Color {
	return Color{hex: hex}
}

// Hex returns the hex string representation of the color.
func (c Color) Hex() string {
	return c.hex
}

// String returns the color as a string (implements fmt.Stringer).
func (c Color) String() string {
	return c.hex
}

// Lipgloss returns the color as a lipgloss.Color for use with lipgloss styles.
func (c Color) Lipgloss() lipgloss.Color {
	return lipgloss.Color(c.hex)
}

// RGBA implements the color.Color interface.
func (c Color) RGBA() (r, g, b, a uint32) {
	hex := c.hex
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 0, 0, 0, 0xFFFF
	}

	var ri, gi, bi uint8
	parseHexByte(hex[0:2], &ri)
	parseHexByte(hex[2:4], &gi)
	parseHexByte(hex[4:6], &bi)

	return uint32(ri) * 257, uint32(gi) * 257, uint32(bi) * 257, 0xFFFF
}

func parseHexByte(s string, v *uint8) {
	var val uint8
	for _, ch := range s {
		val *= 16
		switch {
		case ch >= '0' && ch <= '9':
			val += uint8(ch - '0')
		case ch >= 'a' && ch <= 'f':
			val += uint8(ch - 'a' + 10)
		case ch >= 'A' && ch <= 'F':
			val += uint8(ch - 'A' + 10)
		}
	}
	*v = val
}

// Ensure Color implements color.Color
var _ color.Color = Color{}

// Note: charmtone.Color cannot directly implement lipgloss.TerminalColor
// because that interface has an unexported method. Use .Lipgloss() to convert
// to lipgloss.Color when needed.

// Primary colors
var (
	Charple = NewColor("#7B61FF") // Purple-ish primary
	Dolly   = NewColor("#FFD93D") // Yellow secondary
	Bok     = NewColor("#6BCB77") // Green tertiary
	Zest    = NewColor("#FF6B6B") // Orange-red accent
)

// Background colors
var (
	Pepper   = NewColor("#0D0D0D") // Darkest background
	BBQ      = NewColor("#1A1A1A") // Slightly lighter background
	Charcoal = NewColor("#262626") // Subtle background
	Iron     = NewColor("#333333") // Overlay background
)

// Foreground/Text colors
var (
	Ash    = NewColor("#E0E0E0") // Base text
	Squid  = NewColor("#6B6B6B") // Muted text
	Smoke  = NewColor("#B3B3B3") // Half-muted text
	Oyster = NewColor("#8A8A8A") // Subtle text
	Salt   = NewColor("#FFFFFF") // Selected/highlight text
	Butter = NewColor("#FFF8DC") // White-ish
)

// Status colors
var (
	Guac     = NewColor("#6BCB77") // Success green
	Sriracha = NewColor("#FF4757") // Error red
	Malibu   = NewColor("#54A0FF") // Info blue
	Citron   = NewColor("#FFC312") // Warning yellow
)

// Blue variants
var (
	Sardine = NewColor("#74B9FF") // Light blue
	Damson  = NewColor("#4834D4") // Dark blue
)

// Yellow variants
var (
	Mustard = NewColor("#F9CA24") // Yellow
)

// Green variants
var (
	Julep = NewColor("#26DE81") // Green
)

// Red variants
var (
	Coral  = NewColor("#FF6B6B") // Coral red
	Salmon = NewColor("#FF7979") // Light red
	Cherry = NewColor("#EB3B5A") // Dark red
)

// Syntax highlighting colors
var (
	Bengal = NewColor("#F79F1F") // Preprocessor comments
	Pony   = NewColor("#A55EEA") // Reserved keywords, namespaces
	Guppy  = NewColor("#45AAF2") // Type keywords
	Cheeky = NewColor("#FC5C65") // Builtin names
	Mauve  = NewColor("#D980FA") // Tags
	Hazy   = NewColor("#7D5FFF") // Attributes
	Cumin  = NewColor("#F8B500") // Strings
	Zinc   = NewColor("#778CA3") // Links
)

// Additional grays
var (
	Thunder = NewColor("#2F2F2F")
	Anchovy = NewColor("#4F4F4F")
	White   = NewColor("#FFFFFF")
	Black   = NewColor("#000000")
	Gray    = NewColor("#808080")
	Silver  = NewColor("#C0C0C0")
)

// Legacy names for compatibility
var (
	Turtle   = NewColor("#2ECC71")
	Sapphire = NewColor("#3498DB")
	Ox       = NewColor("#9B59B6")
	Gold     = NewColor("#F1C40F")
	Orange   = NewColor("#E67E22")
	Mint     = NewColor("#1ABC9C")
	Sky      = NewColor("#74B9FF")
	Lavender = NewColor("#A29BFE")
)
