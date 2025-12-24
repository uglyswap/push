// Package uiutil provides utility functions for UI message handling.
// TODO: Move to internal/ui/<appropriate_location> once the new UI migration
// is finalized.
package uiutil

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"time"

	tea "github.com/uglyswap/crush/internal/compat/bubbletea"
	"mvdan.cc/sh/v3/shell"
)

// CursorPosition represents a cursor position in the terminal.
// This provides compatibility with bubbletea v2's Cursor type for v1.
type CursorPosition struct {
	X, Y  int
	Shape CursorShape
	Blink bool
}

// CursorShape represents the shape of the cursor.
type CursorShape int

// Cursor shapes.
const (
	CursorBlock CursorShape = iota
	CursorUnderline
	CursorBar
)

type Cursor interface {
	Cursor() *CursorPosition
}

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func ReportError(err error) tea.Cmd {
	slog.Error("Error reported", "error", err)
	return CmdHandler(InfoMsg{
		Type: InfoTypeError,
		Msg:  err.Error(),
	})
}

type InfoType int

const (
	InfoTypeInfo InfoType = iota
	InfoTypeSuccess
	InfoTypeWarn
	InfoTypeError
	InfoTypeUpdate
)

func ReportInfo(info string) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeInfo,
		Msg:  info,
	})
}

func ReportWarn(warn string) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeWarn,
		Msg:  warn,
	})
}

type (
	InfoMsg struct {
		Type InfoType
		Msg  string
		TTL  time.Duration
	}
	ClearStatusMsg struct{}
)

// ExecShell parses a shell command string and executes it with exec.Command.
// Uses shell.Fields for proper handling of shell syntax like quotes and
// arguments while preserving TTY handling for terminal editors.
func ExecShell(ctx context.Context, cmdStr string, callback tea.ExecCallback) tea.Cmd {
	fields, err := shell.Fields(cmdStr, nil)
	if err != nil {
		return ReportError(err)
	}
	if len(fields) == 0 {
		return ReportError(errors.New("empty command"))
	}

	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)
	return tea.ExecProcess(cmd, callback)
}
