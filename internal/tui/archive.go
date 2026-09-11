package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ekholme/eeban/internal/domain"
)

// openArchive shows the archived-card list for the current board.
func (m Model) openArchive() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	m.showArchive = true
	m.archiveCursor = 0
	return m, m.loadArchiveCmd()
}

// updateArchive handles keys while the archive view is open.
func (m Model) updateArchive(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.ArchiveView), key.Matches(msg, m.keys.Back):
		m.showArchive = false
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.archiveCursor = clamp(m.archiveCursor-1, 0, max(len(m.archive)-1, 0))
	case key.Matches(msg, m.keys.Down):
		m.archiveCursor = clamp(m.archiveCursor+1, 0, max(len(m.archive)-1, 0))
	case key.Matches(msg, m.keys.Restore):
		if c, ok := m.selectedArchived(); ok {
			return m, m.unarchiveCardCmd(c.ID)
		}
	case key.Matches(msg, m.keys.Delete):
		if c, ok := m.selectedArchived(); ok {
			m.confirm = &confirmState{
				prompt: fmt.Sprintf("Permanently delete %q?", truncateText(c.Title, 40)),
				action: m.deleteCardCmd(c.ID),
			}
		}
	}
	return m, nil
}

func (m Model) selectedArchived() (domain.Card, bool) {
	if m.archiveCursor < 0 || m.archiveCursor >= len(m.archive) {
		return domain.Card{}, false
	}
	return m.archive[m.archiveCursor], true
}

// archiveView renders the full-screen archived-card list.
func (m Model) archiveView() string {
	var b strings.Builder
	b.WriteString(m.styles.HelpHeading.Render(fmt.Sprintf("%s — archive", m.board.Name)))
	b.WriteByte('\n')

	if len(m.archive) == 0 {
		b.WriteString(m.styles.DetailDim.Render("No archived cards."))
	}
	for i, c := range m.archive {
		line := c.Title
		if c.ArchivedAt != nil && *c.ArchivedAt != "" {
			line = fmt.Sprintf("%s  %s", c.Title, m.styles.DetailDim.Render("archived "+*c.ArchivedAt))
		}
		if chips := m.renderLabelChips(c.Labels); chips != "" {
			line += "  " + chips
		}
		if i == m.archiveCursor {
			b.WriteString(m.styles.OverlayRowActive.Render("> " + line))
		} else {
			b.WriteString(m.styles.OverlayRow.Render(line))
		}
		b.WriteByte('\n')
	}

	b.WriteString(m.styles.DetailDim.Render("\nr restore · d delete · j/k move · A/esc close"))
	b.WriteString(m.confirmLine())

	box := m.styles.HelpOverlay.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
