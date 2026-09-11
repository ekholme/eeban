package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

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
	if m.showArchive {
		return m.archiveView()
	}
	if m.labelPicker {
		return m.labelPickerView()
	}
	if m.boardSwitcher {
		return m.boardSwitcherView()
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

// header renders the board title, an active-filter chip, and the transient
// toast pinned to the top-right when one is showing.
func (m Model) header() string {
	left := m.styles.BoardTitle.Render(m.board.Name)
	if fc := m.filterChip(); fc != "" {
		left = lipgloss.JoinHorizontal(lipgloss.Bottom, left, "  ", fc)
	}
	if m.toast == "" {
		return left
	}
	toast := m.styles.Toast.Render(m.toast)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(toast)
	if gap < 1 {
		return lipgloss.JoinVertical(lipgloss.Left, left, toast)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), toast)
}

// filterChip describes the active filter(s), or "" when none is set.
func (m Model) filterChip() string {
	var parts []string
	if m.filter != "" {
		parts = append(parts, fmt.Sprintf("search:%q", m.filter))
	}
	if m.filterLabel != 0 {
		name := fmt.Sprintf("#%d", m.filterLabel)
		for _, l := range m.board.Labels {
			if l.ID == m.filterLabel {
				name = l.Name
				break
			}
		}
		parts = append(parts, "label:"+name)
	}
	if len(parts) == 0 {
		return ""
	}
	return m.styles.FilterBar.Render("filter " + strings.Join(parts, " ") + " · esc clears")
}

// helpView renders the full-screen keybinding overlay.
func (m Model) helpView() string {
	groups := []struct {
		name     string
		bindings []key.Binding
	}{
		{"Navigate", []key.Binding{m.keys.Left, m.keys.Right, m.keys.Up, m.keys.Down, m.keys.Detail, m.keys.Back}},
		{"Cards", []key.Binding{m.keys.New, m.keys.Edit, m.keys.Delete, m.keys.Archive, m.keys.MoveLeft, m.keys.MoveRight, m.keys.ReorderUp, m.keys.ReorderDown}},
		{"Columns", []key.Binding{m.keys.NewColumn, m.keys.RenameColumn, m.keys.DeleteColumn, m.keys.ColumnLeft, m.keys.ColumnRight, m.keys.WIPLimit}},
		{"Organise", []key.Binding{m.keys.Labels, m.keys.Search, m.keys.ArchiveView, m.keys.Boards, m.keys.Undo}},
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

// footer renders the active input prompt, a pending confirmation, or the
// contextual help line.
func (m Model) footer() string {
	switch {
	case m.adding:
		col := m.board.Columns[m.colCursor].Name
		return m.styles.Prompt.Render(fmt.Sprintf("New card in %s: ", col)) + m.titleInput.View()
	case m.addingColumn:
		return m.styles.Prompt.Render("New column: ") + m.columnInput.View()
	case m.renamingColumn:
		return m.styles.Prompt.Render("Rename column: ") + m.columnInput.View()
	case m.settingWIP:
		col := m.board.Columns[m.colCursor].Name
		return m.styles.Prompt.Render(fmt.Sprintf("WIP limit for %s: ", col)) + m.wipInput.View()
	case m.filtering:
		return m.styles.Prompt.Render("Search: ") + m.filterInput.View()
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
		return "j/k card · h/l column · t labels · enter/esc close · ? help · q quit"
	}
	return "h/l·j/k nav · n/e/d/a card · t tag · / search · A archive · b boards · u undo · ? help"
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
	b.WriteByte('\n')

	b.WriteString(m.styles.DetailLabel.Render("Labels    "))
	if chips := m.renderLabelChips(card.Labels); chips != "" {
		b.WriteString(chips)
	} else {
		b.WriteString(m.styles.DetailDim.Render("—"))
	}

	if card.Body != "" {
		b.WriteString(m.styles.DetailBody.Render(renderMarkdown(card.Body, inner)))
	} else {
		b.WriteString(m.styles.DetailBody.Render(m.styles.DetailDim.Render("(no description)")))
	}

	return frame.Render(b.String())
}

// renderMarkdown renders a card body as markdown for the detail pane,
// word-wrapped to width and styled for the terminal's light/dark background.
// It falls back to the raw text if glamour fails to construct a renderer or
// render the input, which can happen for pathological input.
//
// This deliberately avoids glamour.WithAutoStyle(): it re-detects the
// background via termenv.HasDarkBackground() on every call, and that
// termenv lookup is uncached, so it re-queries the terminal (an OSC query
// round-trip) each time. Bubble Tea already owns stdin in raw mode while
// the program runs, so that query races Bubble Tea's own input reader for
// the response and reliably loses, blocking for termenv's ~5s OSCTimeout
// on every re-render of the pane. markdownStyle below reimplements the
// same choice glamour's auto style makes, swapping in
// lipgloss.HasDarkBackground(), which is safe: it's cached (sync.Once) and
// pre-warmed by Bubble Tea's own init() before the terminal is acquired —
// see charmbracelet/bubbletea's tea_init.go for the same workaround.
func renderMarkdown(body string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(markdownStyle()),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return body
	}
	out, err := r.Render(body)
	if err != nil {
		return body
	}
	return strings.TrimRight(out, "\n")
}

// markdownStyle picks a glamour standard style name without ever querying
// the terminal at render time (see renderMarkdown). term.IsTerminal is a
// local ioctl, not a terminal round-trip, so it's always safe to call.
func markdownStyle() string {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return "notty"
	}
	if lipgloss.HasDarkBackground() {
		return "dark"
	}
	return "light"
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

	cards := m.visibleCards(col)
	count := fmt.Sprintf("%d", len(col.Cards))
	if col.WIPLimit != nil {
		count = fmt.Sprintf("%d/%d", len(col.Cards), *col.WIPLimit)
	}
	titleStyle := m.styles.ColumnTitle
	if col.OverWIP() {
		titleStyle = m.styles.ColumnTitleWarn
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("%s  %s", col.Name, count)))
	b.WriteByte('\n')

	switch {
	case len(col.Cards) == 0:
		b.WriteString(m.styles.Empty.Render("(empty)"))
	case len(cards) == 0:
		b.WriteString(m.styles.Empty.Render("(no matches)"))
	}
	now := time.Now()
	for j, card := range cards {
		style := m.styles.Card
		if active && j == m.cardCursor {
			style = m.styles.CardActive
		}

		title := card.Title
		if dots := m.renderLabelDots(card.Labels); dots != "" {
			title += " " + dots
		}
		cell := title
		if tag := m.dueTag(card, now); tag != "" {
			cell = lipgloss.JoinVertical(lipgloss.Left, title, tag)
		}
		b.WriteString(style.Width(width - 3).Render(cell))
		if j < len(cards)-1 {
			b.WriteByte('\n')
		}
	}

	frame := m.styles.Column
	switch {
	case active:
		frame = m.styles.ColumnActive
	case col.OverWIP():
		frame = m.styles.ColumnWarn
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
