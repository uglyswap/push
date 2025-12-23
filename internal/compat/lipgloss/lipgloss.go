// Package lipgloss provides compatibility shims for github.com/charmbracelet/lipgloss
// using the public github.com/charmbracelet/lipgloss package.
package lipgloss

import (
	"github.com/charmbracelet/lipgloss"
)

// Re-export core types
type (
	Style           = lipgloss.Style
	Position        = lipgloss.Position
	TerminalColor   = lipgloss.TerminalColor
	Color           = lipgloss.Color
	AdaptiveColor   = lipgloss.AdaptiveColor
	CompleteColor   = lipgloss.CompleteColor
	Border          = lipgloss.Border
	Renderer        = lipgloss.Renderer
	WhitespaceOption = lipgloss.WhitespaceOption
)

// Re-export functions
var (
	NewStyle      = lipgloss.NewStyle
	NewRenderer   = lipgloss.NewRenderer
	DefaultRenderer = lipgloss.DefaultRenderer

	// Layout functions
	JoinHorizontal = lipgloss.JoinHorizontal
	JoinVertical   = lipgloss.JoinVertical
	Place          = lipgloss.Place
	PlaceHorizontal = lipgloss.PlaceHorizontal
	PlaceVertical   = lipgloss.PlaceVertical
	Width          = lipgloss.Width
	Height         = lipgloss.Height
	Size           = lipgloss.Size

	// Borders
	NormalBorder    = lipgloss.NormalBorder
	RoundedBorder   = lipgloss.RoundedBorder
	BlockBorder     = lipgloss.BlockBorder
	OuterHalfBlockBorder = lipgloss.OuterHalfBlockBorder
	InnerHalfBlockBorder = lipgloss.InnerHalfBlockBorder
	ThickBorder     = lipgloss.ThickBorder
	DoubleBorder    = lipgloss.DoubleBorder
	HiddenBorder    = lipgloss.HiddenBorder

	// Colors - NoColor is a function that returns NoColor{}
	NoColor = func() lipgloss.NoColor { return lipgloss.NoColor{} }
)

// Position constants
const (
	Top    = lipgloss.Top
	Bottom = lipgloss.Bottom
	Center = lipgloss.Center
	Left   = lipgloss.Left
	Right  = lipgloss.Right
)

// HasDarkBackground returns true if the terminal has a dark background.
func HasDarkBackground() bool {
	return lipgloss.HasDarkBackground()
}

// Layer represents a compositing layer for rendering.
// This is a stub for charm.land/x/exp/lipgloss.Layer.
type Layer struct {
	content string
	x, y    int
}

// NewLayer creates a new Layer with optional content.
func NewLayer(content ...string) *Layer {
	l := &Layer{}
	if len(content) > 0 {
		l.content = content[0]
	}
	return l
}

// String returns the layer content.
func (l *Layer) String() string {
	return l.content
}

// X sets the x position.
func (l *Layer) X(x int) *Layer {
	l.x = x
	return l
}

// Y sets the y position.
func (l *Layer) Y(y int) *Layer {
	l.y = y
	return l
}

// Fill fills the layer with the given style and content.
func (l *Layer) Fill(s Style, content string) *Layer {
	l.content = s.Render(content)
	return l
}

// PlaceOverlay places content at the given position.
func (l *Layer) PlaceOverlay(x, y int, content string, opts ...any) *Layer {
	if l.content == "" {
		l.content = content
	}
	return l
}

// SetString sets the layer content directly.
func (l *Layer) SetString(content string) *Layer {
	l.content = content
	return l
}

// HyperlinkStyle wraps a lipgloss.Style to add Hyperlink support.
type HyperlinkStyle struct {
	Style
	url string
	id  string
}

// NewHyperlinkStyle creates a style that can add hyperlinks.
func NewHyperlinkStyle() HyperlinkStyle {
	return HyperlinkStyle{Style: lipgloss.NewStyle()}
}

// Hyperlink sets the hyperlink URL and returns the style for chaining.
// Note: Terminal hyperlink support (OSC 8) is terminal-dependent.
func (s HyperlinkStyle) Hyperlink(url string, id string) HyperlinkStyle {
	s.url = url
	s.id = id
	return s
}

// Render renders the text, optionally with OSC 8 hyperlink escape codes.
func (s HyperlinkStyle) Render(text ...string) string {
	content := ""
	for _, t := range text {
		content += t
	}
	rendered := s.Style.Render(content)

	// Add OSC 8 hyperlink if URL is set
	if s.url != "" {
		idPart := ""
		if s.id != "" {
			idPart = "id=" + s.id
		}
		// OSC 8 format: \x1b]8;params;url\x07text\x1b]8;;\x07
		return "\x1b]8;" + idPart + ";" + s.url + "\x07" + rendered + "\x1b]8;;\x07"
	}
	return rendered
}

// StyleWithHyperlink extends lipgloss.Style with Hyperlink method.
// This is a drop-in replacement for lipgloss.NewStyle() when hyperlinks are needed.
func StyleWithHyperlink(s Style) HyperlinkStyle {
	return HyperlinkStyle{Style: s}
}

// Range represents a styled range of characters in a string.
// This is a compatibility type for lipgloss.Range from newer versions.
type Range struct {
	Start int
	End   int
	Style Style
}

// NewRange creates a new Range with the given start, end, and style.
func NewRange(start, end int, style Style) Range {
	return Range{Start: start, End: end, Style: style}
}

// StyleRanges applies styles to specific character ranges in a string.
// This is a simplified implementation of lipgloss.StyleRanges.
func StyleRanges(text string, ranges ...Range) string {
	if len(ranges) == 0 {
		return text
	}

	runes := []rune(text)
	result := make([]rune, 0, len(runes))

	// Sort ranges by start position
	// Simple bubble sort since we typically have few ranges
	for i := 0; i < len(ranges)-1; i++ {
		for j := 0; j < len(ranges)-i-1; j++ {
			if ranges[j].Start > ranges[j+1].Start {
				ranges[j], ranges[j+1] = ranges[j+1], ranges[j]
			}
		}
	}

	pos := 0
	for _, r := range ranges {
		if r.Start > pos {
			// Add unstyled text before this range
			result = append(result, runes[pos:r.Start]...)
		}

		if r.Start < len(runes) && r.End <= len(runes) {
			// Apply style to the range
			styledText := r.Style.Render(string(runes[r.Start:r.End]))
			result = append(result, []rune(styledText)...)
		}

		pos = r.End
	}

	// Add remaining text after the last range
	if pos < len(runes) {
		result = append(result, runes[pos:]...)
	}

	return string(result)
}
