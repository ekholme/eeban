package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds every Lipgloss style the board view uses.
type Styles struct {
	BoardTitle      lipgloss.Style
	Column          lipgloss.Style
	ColumnActive    lipgloss.Style
	ColumnWarn      lipgloss.Style
	ColumnTitle     lipgloss.Style
	ColumnTitleWarn lipgloss.Style
	Card            lipgloss.Style
	CardActive      lipgloss.Style
	Empty           lipgloss.Style
	Help            lipgloss.Style
	Detail          lipgloss.Style
	DetailTitle     lipgloss.Style
	DetailLabel     lipgloss.Style
	DetailBody      lipgloss.Style
	DetailDim       lipgloss.Style
	Prompt          lipgloss.Style
	Toast           lipgloss.Style
	FilterBar       lipgloss.Style
	DueSoon         lipgloss.Style
	DueOverdue      lipgloss.Style
	HelpOverlay     lipgloss.Style
	HelpHeading     lipgloss.Style
	HelpGroup       lipgloss.Style
}

// DefaultStyles returns styles that adapt to a light or dark terminal.
func DefaultStyles() Styles {
	subtle := lipgloss.AdaptiveColor{Light: "245", Dark: "241"}
	accent := lipgloss.AdaptiveColor{Light: "63", Dark: "63"}
	warn := lipgloss.AdaptiveColor{Light: "166", Dark: "214"}

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
		ColumnWarn: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warn).
			Padding(0, 1),
		ColumnTitle:     lipgloss.NewStyle().Bold(true).MarginBottom(1),
		ColumnTitleWarn: lipgloss.NewStyle().Bold(true).MarginBottom(1).Foreground(warn),
		Card:            cardBase.BorderForeground(subtle),
		CardActive:      cardBase.BorderForeground(accent).Bold(true),
		Empty:           lipgloss.NewStyle().Faint(true),
		Help:            lipgloss.NewStyle().Foreground(subtle).MarginTop(1),
		Detail: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 1).
			MarginLeft(1),
		DetailTitle: lipgloss.NewStyle().Bold(true).MarginBottom(1),
		DetailLabel: lipgloss.NewStyle().Foreground(subtle),
		DetailBody:  lipgloss.NewStyle().MarginTop(1),
		DetailDim:   lipgloss.NewStyle().Faint(true),
		Prompt:      lipgloss.NewStyle().Foreground(accent).MarginTop(1),
		Toast: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(lipgloss.Color("231")).
			Background(lipgloss.AdaptiveColor{Light: "160", Dark: "124"}),
		FilterBar: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		DueSoon: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "166", Dark: "214"}),
		DueOverdue: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "203"}).
			Bold(true),
		HelpOverlay: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 3),
		HelpHeading: lipgloss.NewStyle().Bold(true).MarginBottom(1),
		HelpGroup: lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			MarginTop(1),
	}
}
