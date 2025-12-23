// Package bubbletea provides compatibility shims for github.com/charmbracelet/bubbletea
// using the public github.com/charmbracelet/bubbletea package.
package bubbletea

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Re-export core types
type (
	Cmd       = tea.Cmd
	Model     = tea.Model
	Msg       = tea.Msg
	Program   = tea.Program
	ProgramOption = tea.ProgramOption

	// Window messages
	WindowSizeMsg = tea.WindowSizeMsg
	KeyMsg        = tea.KeyMsg
	MouseMsg      = tea.MouseMsg

	// Commands
	BatchMsg = tea.BatchMsg
)

// Re-export functions
var (
	NewProgram = tea.NewProgram
	Batch      = tea.Batch
	Quit       = tea.Quit
	ClearScreen = tea.ClearScreen
	EnterAltScreen = tea.EnterAltScreen
	ExitAltScreen  = tea.ExitAltScreen
	EnableMouseAllMotion = tea.EnableMouseAllMotion
	EnableMouseCellMotion = tea.EnableMouseCellMotion
	DisableMouse = tea.DisableMouse
	HideCursor = tea.HideCursor
	ShowCursor = tea.ShowCursor
	Printf     = tea.Printf
	Println    = tea.Println
	Sequence   = tea.Sequence
	Every      = tea.Every
	Tick       = tea.Tick

	// Program options
	WithAltScreen       = tea.WithAltScreen
	WithMouseAllMotion  = tea.WithMouseAllMotion
	WithMouseCellMotion = tea.WithMouseCellMotion
	WithInput           = tea.WithInput
	WithOutput          = tea.WithOutput
	WithoutSignalHandler = tea.WithoutSignalHandler
	WithoutCatchPanics   = tea.WithoutCatchPanics
	WithANSICompressor   = tea.WithANSICompressor
)

// QuitMsg is a message that tells the program to quit.
type QuitMsg = tea.QuitMsg

// KeyType is the type of a key press.
type KeyType = tea.KeyType

// Key type constants
const (
	KeyCtrlC     = tea.KeyCtrlC
	KeyCtrlD     = tea.KeyCtrlD
	KeyEnter     = tea.KeyEnter
	KeyEsc       = tea.KeyEscape
	KeyEscape    = tea.KeyEscape
	KeyUp        = tea.KeyUp
	KeyDown      = tea.KeyDown
	KeyLeft      = tea.KeyLeft
	KeyRight     = tea.KeyRight
	KeyTab       = tea.KeyTab
	KeyShiftTab  = tea.KeyShiftTab
	KeyBackspace = tea.KeyBackspace
	KeyDelete    = tea.KeyDelete
	KeySpace     = tea.KeySpace
	KeyRunes     = tea.KeyRunes
	KeyHome      = tea.KeyHome
	KeyEnd       = tea.KeyEnd
	KeyPgUp      = tea.KeyPgUp
	KeyPgDown    = tea.KeyPgDown
	KeyF1        = tea.KeyF1
	KeyF2        = tea.KeyF2
	KeyF3        = tea.KeyF3
	KeyF4        = tea.KeyF4
	KeyF5        = tea.KeyF5
	KeyF6        = tea.KeyF6
	KeyF7        = tea.KeyF7
	KeyF8        = tea.KeyF8
	KeyF9        = tea.KeyF9
	KeyF10       = tea.KeyF10
	KeyF11       = tea.KeyF11
	KeyF12       = tea.KeyF12
)

// KeyPressMsg is an alias for KeyMsg for compatibility.
type KeyPressMsg = tea.KeyMsg

// Cursor represents a cursor position.
type Cursor struct {
	X int
	Y int
}

// SetClipboard returns a command that sets the system clipboard.
func SetClipboard(text string) Cmd {
	return func() Msg {
		// Clipboard operations are handled at the terminal level
		// This is a no-op stub that returns nil
		return nil
	}
}

// ClipboardMsg is sent when clipboard content is available.
type ClipboardMsg string

// MouseWheelMsg is an alias for MouseMsg for mouse wheel events.
// In standard bubbletea, mouse wheel events are MouseMsg with Button set to MouseWheelUp/Down.
type MouseWheelMsg = tea.MouseMsg

// MouseButton is the type for mouse buttons.
type MouseButton = tea.MouseButton

// Mouse button constants
const (
	MouseWheelUp   = tea.MouseWheelUp
	MouseWheelDown = tea.MouseWheelDown
)

// PasteMsg is sent when text is pasted from clipboard.
// Note: In standard bubbletea, paste events may come through KeyMsg.
// This type is provided for compatibility with code that expects PasteMsg.
type PasteMsg string

// String returns the pasted text.
func (p PasteMsg) String() string {
	return string(p)
}
