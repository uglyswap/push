package lsp

import (
	"context"
	"fmt"
)

// Operation represents an LSP operation type.
type Operation string

const (
	OperationGoToDefinition     Operation = "goToDefinition"
	OperationFindReferences     Operation = "findReferences"
	OperationHover              Operation = "hover"
	OperationDocumentSymbol     Operation = "documentSymbol"
	OperationWorkspaceSymbol    Operation = "workspaceSymbol"
	OperationGoToImplementation Operation = "goToImplementation"
	OperationPrepareCallHierarchy Operation = "prepareCallHierarchy"
	OperationIncomingCalls      Operation = "incomingCalls"
	OperationOutgoingCalls      Operation = "outgoingCalls"
)

// Position represents a position in a text document.
type Position struct {
	Line      int `json:"line"`      // 0-based line number
	Character int `json:"character"` // 0-based character offset
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// LocationLink represents a link to a location.
type LocationLink struct {
	OriginSelectionRange *Range `json:"originSelectionRange,omitempty"`
	TargetURI            string `json:"targetUri"`
	TargetRange          Range  `json:"targetRange"`
	TargetSelectionRange Range  `json:"targetSelectionRange"`
}

// SymbolKind represents the kind of a symbol.
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// DocumentSymbol represents a symbol in a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Tags           []int            `json:"tags,omitempty"`
	Deprecated     bool             `json:"deprecated,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation represents information about a symbol.
type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Tags          []int      `json:"tags,omitempty"`
	Deprecated    bool       `json:"deprecated,omitempty"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

// WorkspaceSymbol represents a symbol in the workspace.
type WorkspaceSymbol struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Tags          []int      `json:"tags,omitempty"`
	ContainerName string     `json:"containerName,omitempty"`
	Location      Location   `json:"location"`
}

// Hover represents hover information.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents markup content.
type MarkupContent struct {
	Kind  string `json:"kind"` // "plaintext" or "markdown"
	Value string `json:"value"`
}

// CallHierarchyItem represents an item in the call hierarchy.
type CallHierarchyItem struct {
	Name           string     `json:"name"`
	Kind           SymbolKind `json:"kind"`
	Tags           []int      `json:"tags,omitempty"`
	Detail         string     `json:"detail,omitempty"`
	URI            string     `json:"uri"`
	Range          Range      `json:"range"`
	SelectionRange Range      `json:"selectionRange"`
	Data           any        `json:"data,omitempty"`
}

// CallHierarchyIncomingCall represents an incoming call.
type CallHierarchyIncomingCall struct {
	From       CallHierarchyItem `json:"from"`
	FromRanges []Range           `json:"fromRanges"`
}

// CallHierarchyOutgoingCall represents an outgoing call.
type CallHierarchyOutgoingCall struct {
	To         CallHierarchyItem `json:"to"`
	FromRanges []Range           `json:"fromRanges"`
}

// LSPService defines the interface for LSP operations.
type LSPService interface {
	// GoToDefinition finds the definition of a symbol.
	GoToDefinition(ctx context.Context, filePath string, line, character int) ([]Location, error)

	// FindReferences finds all references to a symbol.
	FindReferences(ctx context.Context, filePath string, line, character int, includeDeclaration bool) ([]Location, error)

	// Hover returns hover information for a position.
	Hover(ctx context.Context, filePath string, line, character int) (*Hover, error)

	// DocumentSymbol returns all symbols in a document.
	DocumentSymbol(ctx context.Context, filePath string) ([]DocumentSymbol, error)

	// WorkspaceSymbol searches for symbols in the workspace.
	WorkspaceSymbol(ctx context.Context, query string) ([]WorkspaceSymbol, error)

	// GoToImplementation finds implementations of an interface or abstract method.
	GoToImplementation(ctx context.Context, filePath string, line, character int) ([]Location, error)

	// PrepareCallHierarchy prepares a call hierarchy for a position.
	PrepareCallHierarchy(ctx context.Context, filePath string, line, character int) ([]CallHierarchyItem, error)

	// IncomingCalls finds all incoming calls to a call hierarchy item.
	IncomingCalls(ctx context.Context, item CallHierarchyItem) ([]CallHierarchyIncomingCall, error)

	// OutgoingCalls finds all outgoing calls from a call hierarchy item.
	OutgoingCalls(ctx context.Context, item CallHierarchyItem) ([]CallHierarchyOutgoingCall, error)
}

// FormatSymbolKind returns a human-readable name for a symbol kind.
func FormatSymbolKind(kind SymbolKind) string {
	switch kind {
	case SymbolKindFile:
		return "File"
	case SymbolKindModule:
		return "Module"
	case SymbolKindNamespace:
		return "Namespace"
	case SymbolKindPackage:
		return "Package"
	case SymbolKindClass:
		return "Class"
	case SymbolKindMethod:
		return "Method"
	case SymbolKindProperty:
		return "Property"
	case SymbolKindField:
		return "Field"
	case SymbolKindConstructor:
		return "Constructor"
	case SymbolKindEnum:
		return "Enum"
	case SymbolKindInterface:
		return "Interface"
	case SymbolKindFunction:
		return "Function"
	case SymbolKindVariable:
		return "Variable"
	case SymbolKindConstant:
		return "Constant"
	case SymbolKindString:
		return "String"
	case SymbolKindNumber:
		return "Number"
	case SymbolKindBoolean:
		return "Boolean"
	case SymbolKindArray:
		return "Array"
	case SymbolKindObject:
		return "Object"
	case SymbolKindKey:
		return "Key"
	case SymbolKindNull:
		return "Null"
	case SymbolKindEnumMember:
		return "EnumMember"
	case SymbolKindStruct:
		return "Struct"
	case SymbolKindEvent:
		return "Event"
	case SymbolKindOperator:
		return "Operator"
	case SymbolKindTypeParameter:
		return "TypeParameter"
	default:
		return fmt.Sprintf("Unknown(%d)", kind)
	}
}

// FormatLocation formats a location for display.
func FormatLocation(loc Location) string {
	return fmt.Sprintf("%s:%d:%d", loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
}

// FormatHover formats hover information for display.
func FormatHover(hover *Hover) string {
	if hover == nil {
		return "No hover information available."
	}
	return hover.Contents.Value
}
