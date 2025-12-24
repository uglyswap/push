package util

import (
	tea "github.com/uglyswap/crush/internal/compat/bubbletea"
	compat_tea "github.com/uglyswap/crush/internal/compat/bubbletea"
	"github.com/uglyswap/crush/internal/uiutil"
)

// CursorState is the uiutil Cursor type (Mode, Style).
type CursorState = uiutil.Cursor

// Cursor is a position-based cursor with X and Y coordinates.
type Cursor = compat_tea.Cursor

// CursorProvider is an interface for components that provide cursor information.
type CursorProvider interface {
	Cursor() *Cursor
}

type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (Model, tea.Cmd)
	View() string
}

func CmdHandler(msg tea.Msg) tea.Cmd {
	return uiutil.CmdHandler(msg)
}

func ReportError(err error) tea.Cmd {
	return uiutil.ReportError(err)
}

type InfoType = uiutil.InfoType

const (
	InfoTypeInfo    = uiutil.InfoTypeInfo
	InfoTypeSuccess = uiutil.InfoTypeSuccess
	InfoTypeWarn    = uiutil.InfoTypeWarn
	InfoTypeError   = uiutil.InfoTypeError
	InfoTypeUpdate  = uiutil.InfoTypeUpdate
)

func ReportInfo(info string) tea.Cmd {
	return uiutil.ReportInfo(info)
}

func ReportWarn(warn string) tea.Cmd {
	return uiutil.ReportWarn(warn)
}

type (
	InfoMsg        = uiutil.InfoMsg
	ClearStatusMsg = uiutil.ClearStatusMsg
)
