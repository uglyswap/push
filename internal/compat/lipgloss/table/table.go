// Package table provides compatibility shims for github.com/charmbracelet/lipgloss/table
// using the public github.com/charmbracelet/lipgloss/table package.
package table

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Re-export core types
type (
	Table      = table.Table
	StringData = table.StringData
	StyleFunc  = table.StyleFunc
)

// HeaderStyleFunc is a function that returns a style for a header cell.
// This type may not exist in all versions of lipgloss/table.
type HeaderStyleFunc = func(row, col int) lipgloss.Style

// Re-export functions
var (
	New = table.New
)
