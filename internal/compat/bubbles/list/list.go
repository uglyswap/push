// Package list provides compatibility shims for github.com/charmbracelet/bubbles/list
package list

import (
	"github.com/charmbracelet/bubbles/list"
)

// Re-export core types
type (
	Model        = list.Model
	Item         = list.Item
	DefaultItem  = list.DefaultItem
	Styles       = list.Styles
	ItemDelegate = list.ItemDelegate
	DefaultDelegate = list.DefaultDelegate
	FilterFunc   = list.FilterFunc
	FilterState  = list.FilterState
	KeyMap       = list.KeyMap
)

// Re-export functions
var (
	New               = list.New
	NewDefaultDelegate = list.NewDefaultDelegate
	DefaultFilter      = list.DefaultFilter
	DefaultStyles      = list.DefaultStyles
)

// Filter state constants
const (
	Unfiltered = list.Unfiltered
	Filtering  = list.Filtering
	FilterApplied = list.FilterApplied
)
