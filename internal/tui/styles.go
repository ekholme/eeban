package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds every Lipgloss style the board view uses.
type Styles struct {
	BoardTitle   lipgloss.Style
	Column       lipgloss.Style
	ColumnActive lipgloss.Style
	ColumnTitle  lipgloss.Style
	Card         lipgloss.Style
	CardActive   lipgloss.Style
	Empty        lipgloss.Style
	Help         lipgloss.Style
}

// DefaultStyles returns styles that adapt to a light or dark terminal.
func DefaultStyles() Styles {
	subtle := lipgloss.AdaptiveColor{Light: "245", Dark: "241"}
	accent := lipgloss.AdaptiveColor{Light: "63", Dark: "63"}

	cardBase := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		PaddingLeft(1)

	return Styles{
		BoardTitle: lipgloss.NewStyle().Bold(true).MarginBottom(1),
		Column: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Padding(0, 1),
		ColumnActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 1),
		ColumnTitle: lipgloss.NewStyle().Bold(true).MarginBottom(1),
		Card:        cardBase.BorderForeground(subtle),
		CardActive:  cardBase.BorderForeground(accent).Bold(true),
		Empty:       lipgloss.NewStyle().Faint(true),
		Help:        lipgloss.NewStyle().Foreground(subtle).MarginTop(1),
	}
}
