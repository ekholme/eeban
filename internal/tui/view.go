package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ekholme/eeban/internal/domain"
)

const (
	minColWidth = 22
	maxColWidth = 40
	// chromeHeight is the vertical space taken by the board title and help
	// line plus the column border rows.
	chromeHeight = 6
)

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "loading…"
	}
	if len(m.board.Columns) == 0 {
		return m.board.Name + "\n\n(no columns yet)\n\npress q to quit"
	}

	colWidth := clamp(m.width/len(m.board.Columns)-2, minColWidth, maxColWidth)
	colHeight := max(m.height-chromeHeight, 3)

	cols := make([]string, len(m.board.Columns))
	for i, c := range m.board.Columns {
		cols[i] = m.renderColumn(c, colWidth, colHeight, i == m.colCursor)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.styles.BoardTitle.Render(m.board.Name),
		lipgloss.JoinHorizontal(lipgloss.Top, cols...),
		m.styles.Help.Render("h/l column · j/k card · ? help · q quit"),
	)
}

func (m Model) renderColumn(col domain.Column, width, height int, active bool) string {
	var b strings.Builder

	count := fmt.Sprintf("%d", len(col.Cards))
	if col.WIPLimit != nil {
		count = fmt.Sprintf("%d/%d", len(col.Cards), *col.WIPLimit)
	}
	b.WriteString(m.styles.ColumnTitle.Render(fmt.Sprintf("%s  %s", col.Name, count)))
	b.WriteByte('\n')

	if len(col.Cards) == 0 {
		b.WriteString(m.styles.Empty.Render("(empty)"))
	}
	for j, card := range col.Cards {
		style := m.styles.Card
		if active && j == m.cardCursor {
			style = m.styles.CardActive
		}
		b.WriteString(style.Width(width - 3).Render(card.Title))
		if j < len(col.Cards)-1 {
			b.WriteByte('\n')
		}
	}

	frame := m.styles.Column
	if active {
		frame = m.styles.ColumnActive
	}
	return frame.Width(width).Height(height).MaxHeight(height + 2).Render(b.String())
}
