package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the board-view key bindings.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Detail      key.Binding
	Back        key.Binding
	New         key.Binding
	Edit        key.Binding
	Delete      key.Binding
	MoveLeft    key.Binding
	MoveRight   key.Binding
	ReorderUp   key.Binding
	ReorderDown key.Binding

	NewColumn    key.Binding
	RenameColumn key.Binding
	DeleteColumn key.Binding
	ColumnLeft   key.Binding
	ColumnRight  key.Binding
	WIPLimit     key.Binding

	Labels      key.Binding
	Search      key.Binding
	Archive     key.Binding
	ArchiveView key.Binding
	Restore     key.Binding

	Help key.Binding
	Quit key.Binding
}

// DefaultKeyMap returns the vim + arrow-key defaults.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:          key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "prev card")),
		Down:        key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "next card")),
		Left:        key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/←", "prev column")),
		Right:       key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/→", "next column")),
		Detail:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle detail")),
		Back:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close / clear filter")),
		New:         key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new card")),
		Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit card")),
		Delete:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete card")),
		MoveLeft:    key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "move card left")),
		MoveRight:   key.NewBinding(key.WithKeys("L"), key.WithHelp("L", "move card right")),
		ReorderUp:   key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "move card up")),
		ReorderDown: key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "move card down")),

		NewColumn:    key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "new column")),
		RenameColumn: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "rename column")),
		DeleteColumn: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "delete column")),
		ColumnLeft:   key.NewBinding(key.WithKeys("["), key.WithHelp("[", "reorder column left")),
		ColumnRight:  key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "reorder column right")),
		WIPLimit:     key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "set WIP limit")),

		Labels:      key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "labels")),
		Search:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Archive:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive card")),
		ArchiveView: key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "archive view")),
		Restore:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "restore card")),

		Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
