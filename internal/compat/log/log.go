// Package log provides compatibility shims for github.com/charmbracelet/log
// using the public github.com/charmbracelet/log package.
package log

import (
	"io"

	"github.com/charmbracelet/log"
)

// Re-export core types
type (
	Logger  = log.Logger
	Level   = log.Level
	Styles  = log.Styles
	Options = log.Options
)

// Re-export functions
var (
	New         = log.New
	NewWithOptions = log.NewWithOptions
	Default     = log.Default
	SetDefault  = log.SetDefault
	With        = log.With
	WithPrefix  = log.WithPrefix
	Debug       = log.Debug
	Info        = log.Info
	Warn        = log.Warn
	Error       = log.Error
	Fatal       = log.Fatal
	Print       = log.Print
	Debugf      = log.Debugf
	Infof       = log.Infof
	Warnf       = log.Warnf
	Errorf      = log.Errorf
	Fatalf      = log.Fatalf
	Printf      = log.Printf
	DefaultStyles = log.DefaultStyles
	SetLevel    = log.SetLevel
	GetLevel    = log.GetLevel
	SetOutput   = log.SetOutput
	SetReportCaller = log.SetReportCaller
	SetReportTimestamp = log.SetReportTimestamp
	SetTimeFormat = log.SetTimeFormat
	SetFormatter = log.SetFormatter
	StandardLog = log.StandardLog
)

// Level constants
const (
	DebugLevel = log.DebugLevel
	InfoLevel  = log.InfoLevel
	WarnLevel  = log.WarnLevel
	ErrorLevel = log.ErrorLevel
	FatalLevel = log.FatalLevel
)

// Helper function to create a new logger with specific output
func NewLogger(w io.Writer) *log.Logger {
	return log.NewWithOptions(w, log.Options{})
}
