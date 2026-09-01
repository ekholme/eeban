package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the board-view key bindings.
type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// DefaultKeyMap returns the vim + arrow-key defaults.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:    key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "prev card")),
		Down:  key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "next card")),
		Left:  key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/←", "prev column")),
		Right: key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/→", "next column")),
		Help:  key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
