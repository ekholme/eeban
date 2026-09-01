// Package domain holds the core kanban types and the pure rules that operate
// on them. It has no dependency on storage or the TUI.
package domain

// Priority levels for a card.
const (
	PriorityNone = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
)

// PositionGap is the spacing left between adjacent items so that inserting or
// moving an item only rewrites a single row's position.
const PositionGap = 1000

// Board is a single kanban board together with its ordered columns.
type Board struct {
	ID      int64
	Name    string
	Columns []Column
}

// Column is a vertical lane on a board holding an ordered list of cards.
type Column struct {
	ID       int64
	BoardID  int64
	Name     string
	Position int64
	WIPLimit *int
	Cards    []Card
}

// Card is a single work item within a column.
type Card struct {
	ID       int64
	ColumnID int64
	Title    string
	Body     string
	Position int64
	Priority int
	DueDate  *string // RFC3339 date, nil when unset
}
