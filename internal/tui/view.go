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

	colHeight := max(m.height-chromeHeight, 3)

	boardWidth := m.width
	var detail string
	if m.showDetail {
		detailWidth := clamp(m.width/3, 32, 48)
		boardWidth = max(m.width-detailWidth-4, minColWidth)
		detail = m.renderDetail(detailWidth, colHeight)
	}

	colWidth := clamp(boardWidth/len(m.board.Columns)-2, minColWidth, maxColWidth)

	cols := make([]string, len(m.board.Columns))
	for i, c := range m.board.Columns {
		cols[i] = m.renderColumn(c, colWidth, colHeight, i == m.colCursor)
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	if m.showDetail {
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, detail)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.styles.BoardTitle.Render(m.board.Name),
		body,
		m.styles.Help.Render(m.helpLine()),
	)
}

func (m Model) helpLine() string {
	if m.showDetail {
		return "j/k card · h/l column · enter/esc close · q quit"
	}
	return "h/l column · j/k card · enter detail · ? help · q quit"
}

// renderDetail draws the pane describing the selected card.
func (m Model) renderDetail(width, height int) string {
	frame := m.styles.Detail.Width(width).Height(height).MaxHeight(height + 2)

	card, ok := m.selectedCard()
	if !ok {
		return frame.Render(m.styles.DetailDim.Render("No card selected."))
	}

	inner := width - 4 // border + horizontal padding
	var b strings.Builder

	b.WriteString(m.styles.DetailTitle.Width(inner).Render(card.Title))
	b.WriteByte('\n')

	b.WriteString(m.styles.DetailLabel.Render("Priority  "))
	b.WriteString(priorityLabel(card.Priority))
	b.WriteByte('\n')

	b.WriteString(m.styles.DetailLabel.Render("Due       "))
	if card.DueDate != nil && *card.DueDate != "" {
		b.WriteString(*card.DueDate)
	} else {
		b.WriteString(m.styles.DetailDim.Render("—"))
	}

	if card.Body != "" {
		b.WriteString(m.styles.DetailBody.Width(inner).Render(card.Body))
	} else {
		b.WriteString(m.styles.DetailBody.Render(m.styles.DetailDim.Render("(no description)")))
	}

	return frame.Render(b.String())
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
