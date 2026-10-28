package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ekholme/eeban/internal/domain"
)

// openUpcoming shows the cross-board list of overdue and soon-due cards.
func (m Model) openUpcoming() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	m.showUpcoming = true
	m.upcomingCursor = 0
	return m, m.loadUpcomingCmd()
}

// updateUpcoming handles keys while the upcoming view is open. enter jumps
// to the highlighted card on its own board.
func (m Model) updateUpcoming(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Upcoming), key.Matches(msg, m.keys.Back):
		m.showUpcoming = false
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.upcomingCursor = clamp(m.upcomingCursor-1, 0, max(len(m.upcoming)-1, 0))
	case key.Matches(msg, m.keys.Down):
		m.upcomingCursor = clamp(m.upcomingCursor+1, 0, max(len(m.upcoming)-1, 0))
	case key.Matches(msg, m.keys.Detail):
		if m.upcomingCursor < 0 || m.upcomingCursor >= len(m.upcoming) {
			return m, nil
		}
		c := m.upcoming[m.upcomingCursor]
		m.showUpcoming = false
		return m, m.jumpToCardCmd(c.BoardID, c.ID)
	}
	return m, nil
}

// upcomingSection names the group a card falls in by days until due.
func upcomingSection(days int) string {
	switch {
	case days < 0:
		return "Overdue"
	case days == 0:
		return "Today"
	default:
		return fmt.Sprintf("Next %d days", domain.UpcomingDays)
	}
}

// upcomingDue describes how far off (or overdue) a card is, coloured to
// match the board's due-date highlighting.
func (m Model) upcomingDue(c domain.DueCard) string {
	switch {
	case c.DaysUntil < 0:
		return m.styles.DueOverdue.Render(fmt.Sprintf("%s · %dd overdue", *c.DueDate, -c.DaysUntil))
	case c.DaysUntil == 0:
		return m.styles.DueSoon.Render(*c.DueDate + " · today")
	case c.DaysUntil == 1:
		return m.styles.DueSoon.Render(*c.DueDate + " · tomorrow")
	case c.DaysUntil <= domain.SoonDays:
		return m.styles.DueSoon.Render(fmt.Sprintf("%s · in %dd", *c.DueDate, c.DaysUntil))
	default:
		return fmt.Sprintf("%s · in %dd", *c.DueDate, c.DaysUntil)
	}
}

// upcomingView renders the full-screen cross-board due list, grouped into
// overdue / today / next few days.
func (m Model) upcomingView() string {
	var b strings.Builder
	b.WriteString(m.styles.HelpHeading.Render("upcoming — all boards"))

	b.WriteByte('\n')
	if len(m.upcoming) == 0 {
		b.WriteString(m.styles.DetailDim.Render(
			fmt.Sprintf("Nothing overdue or due in the next %d days.", domain.UpcomingDays)))
		b.WriteByte('\n')
	}

	// Pad title and location to their widest value so the due dates line up.
	titles := make([]string, len(m.upcoming))
	locs := make([]string, len(m.upcoming))
	titleW, locW := 0, 0
	for i, c := range m.upcoming {
		titles[i] = truncateText(c.Title, 50)
		locs[i] = c.BoardName + " › " + c.ColumnName
		titleW = max(titleW, lipgloss.Width(titles[i]))
		locW = max(locW, lipgloss.Width(locs[i]))
	}

	section := ""
	for i, c := range m.upcoming {
		if s := upcomingSection(c.DaysUntil); s != section {
			section = s
			b.WriteString(m.styles.HelpGroup.Render(s))
			b.WriteByte('\n')
		}

		line := fmt.Sprintf("%s  %s  %s",
			padRight(titles[i], titleW),
			m.styles.DetailDim.Render(padRight(locs[i], locW)),
			m.upcomingDue(c))
		if chips := m.renderLabelChips(c.Labels); chips != "" {
			line += "  " + chips
		}
		if i == m.upcomingCursor {
			b.WriteString(m.styles.OverlayRowActive.Render("> " + line))
		} else {
			b.WriteString(m.styles.OverlayRow.Render(line))
		}
		b.WriteByte('\n')
	}

	b.WriteString(m.styles.DetailDim.Render("\nenter go to card · j/k move · U/esc close"))

	box := m.styles.HelpOverlay.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
