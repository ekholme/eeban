package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
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
	if m.showHelp {
		return m.helpView()
	}
	if len(m.board.Columns) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.header(),
			m.styles.Empty.Render("(no columns yet — press N to add one)"),
			m.footer(),
		)
	}

	colHeight := max(m.height-chromeHeight, 3)

	boardWidth := m.width
	var detail string
	showRightPane := m.showDetail || m.editing
	if showRightPane {
		detailWidth := clamp(m.width/3, 32, 48)
		boardWidth = max(m.width-detailWidth-4, minColWidth)
		if m.editing {
			detail = m.renderEditForm(detailWidth, colHeight)
		} else {
			detail = m.renderDetail(detailWidth, colHeight)
		}
	}

	colWidth := clamp(boardWidth/len(m.board.Columns)-2, minColWidth, maxColWidth)

	cols := make([]string, len(m.board.Columns))
	for i, c := range m.board.Columns {
		cols[i] = m.renderColumn(c, colWidth, colHeight, i == m.colCursor)
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	if showRightPane {
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, detail)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.header(),
		body,
		m.footer(),
	)
}

// header renders the board title, with the transient toast pinned to the
// top-right when one is showing.
func (m Model) header() string {
	title := m.styles.BoardTitle.Render(m.board.Name)
	if m.toast == "" {
		return title
	}
	toast := m.styles.Toast.Render(m.toast)
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(toast)
	if gap < 1 {
		return lipgloss.JoinVertical(lipgloss.Left, title, toast)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, title, strings.Repeat(" ", gap), toast)
}

// helpView renders the full-screen keybinding overlay.
func (m Model) helpView() string {
	groups := []struct {
		name     string
		bindings []key.Binding
	}{
		{"Navigate", []key.Binding{m.keys.Left, m.keys.Right, m.keys.Up, m.keys.Down, m.keys.Detail, m.keys.Back}},
		{"Cards", []key.Binding{m.keys.New, m.keys.Edit, m.keys.Delete, m.keys.MoveLeft, m.keys.MoveRight, m.keys.ReorderUp, m.keys.ReorderDown}},
		{"Columns", []key.Binding{m.keys.NewColumn, m.keys.RenameColumn, m.keys.DeleteColumn, m.keys.ColumnLeft, m.keys.ColumnRight}},
		{"General", []key.Binding{m.keys.Help, m.keys.Quit}},
	}

	var b strings.Builder
	b.WriteString(m.styles.HelpHeading.Render("eeban — keybindings"))
	for _, g := range groups {
		b.WriteByte('\n')
		b.WriteString(m.styles.HelpGroup.Render(g.name))
		b.WriteByte('\n')
		for _, bind := range g.bindings {
			h := bind.Help()
			b.WriteString(fmt.Sprintf("  %-12s %s\n", h.Key, h.Desc))
		}
	}
	b.WriteString(m.styles.DetailDim.Render("\npress ? or esc to close"))

	box := m.styles.HelpOverlay.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// footer renders the title input while adding a card, the last mutation
// error if one occurred, or the contextual help line.
func (m Model) footer() string {
	switch {
	case m.adding:
		col := m.board.Columns[m.colCursor].Name
		return m.styles.Prompt.Render(fmt.Sprintf("New card in %s: ", col)) + m.titleInput.View()
	case m.addingColumn:
		return m.styles.Prompt.Render("New column: ") + m.columnInput.View()
	case m.renamingColumn:
		return m.styles.Prompt.Render("Rename column: ") + m.columnInput.View()
	case m.editing:
		return m.styles.Help.Render("tab/shift+tab field · ←/→ priority · ctrl+s save · esc cancel")
	case m.confirm != nil:
		return m.styles.Prompt.Render(m.confirm.prompt + "  (y/n)")
	default:
		return m.styles.Help.Render(m.helpLine())
	}
}

func (m Model) helpLine() string {
	if m.showDetail {
		return "j/k card · h/l column · enter/esc close · ? help · q quit"
	}
	return "h/l·j/k nav · H/L·J/K move/reorder · n/e/d card · N/R/D/[/] col · ? help · q quit"
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
	b.WriteString(m.renderDue(card))

	if card.Body != "" {
		b.WriteString(m.styles.DetailBody.Width(inner).Render(card.Body))
	} else {
		b.WriteString(m.styles.DetailBody.Render(m.styles.DetailDim.Render("(no description)")))
	}

	return frame.Render(b.String())
}

// renderDue formats a card's due date, coloured by how close (or overdue) it is.
func (m Model) renderDue(card domain.Card) string {
	if card.DueDate == nil || *card.DueDate == "" {
		return m.styles.DetailDim.Render("—")
	}
	switch domain.DueStatusFor(card.DueDate, time.Now()) {
	case domain.DueOverdue:
		return m.styles.DueOverdue.Render(*card.DueDate + " (overdue)")
	case domain.DueSoon:
		return m.styles.DueSoon.Render(*card.DueDate + " (soon)")
	default:
		return *card.DueDate
	}
}

// renderEditForm draws the card edit form: title, body, priority, due date.
func (m Model) renderEditForm(width, height int) string {
	frame := m.styles.Detail.Width(width).Height(height).MaxHeight(height + 2)

	inner := width - 4 // border + horizontal padding
	f := m.form
	f.setWidth(inner)
	f.body.SetHeight(clamp(height-10, 3, 10))

	var b strings.Builder
	b.WriteString(m.fieldLabel("Title", f.focus == fieldTitle))
	b.WriteByte('\n')
	b.WriteString(f.title.View())
	b.WriteString("\n\n")

	b.WriteString(m.fieldLabel("Body", f.focus == fieldBody))
	b.WriteByte('\n')
	b.WriteString(f.body.View())
	b.WriteByte('\n')

	b.WriteString(m.fieldLabel("Priority", f.focus == fieldPriority))
	b.WriteString("  " + priorityLabel(f.priority))
	b.WriteByte('\n')

	b.WriteString(m.fieldLabel("Due", f.focus == fieldDueDate))
	b.WriteByte('\n')
	b.WriteString(f.dueDate.View())

	return frame.Render(b.String())
}

// fieldLabel renders an edit-form field label, highlighted when active.
func (m Model) fieldLabel(label string, active bool) string {
	if active {
		return m.styles.Prompt.Render("> " + label)
	}
	return m.styles.DetailLabel.Render("  " + label)
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
	now := time.Now()
	for j, card := range col.Cards {
		style := m.styles.Card
		if active && j == m.cardCursor {
			style = m.styles.CardActive
		}
		cell := card.Title
		if tag := m.dueTag(card, now); tag != "" {
			cell = lipgloss.JoinVertical(lipgloss.Left, card.Title, tag)
		}
		b.WriteString(style.Width(width - 3).Render(cell))
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

// dueTag returns a short coloured due-date marker for the compact card view,
// or "" when the card is not due soon or overdue.
func (m Model) dueTag(card domain.Card, now time.Time) string {
	switch domain.DueStatusFor(card.DueDate, now) {
	case domain.DueOverdue:
		return m.styles.DueOverdue.Render("⏰ overdue")
	case domain.DueSoon:
		return m.styles.DueSoon.Render("⏰ due soon")
	default:
		return ""
	}
}
