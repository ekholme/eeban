package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ekholme/eeban/internal/domain"
)

// openBoardSwitcher opens the board list, refreshing it from the store.
func (m Model) openBoardSwitcher() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	m.boardSwitcher = true
	m.boardNaming = false
	m.boardRenaming = false
	return m, m.loadBoardsCmd()
}

// syncBoardCursor points boardCursor at the currently open board.
func (m *Model) syncBoardCursor() {
	for i, b := range m.boards {
		if b.ID == m.boardID {
			m.boardCursor = i
			return
		}
	}
	m.boardCursor = clamp(m.boardCursor, 0, max(len(m.boards)-1, 0))
}

// updateBoardSwitcher handles keys while the board switcher is open.
func (m Model) updateBoardSwitcher(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.boardNaming || m.boardRenaming {
		return m.updateBoardNameInput(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Boards), key.Matches(msg, m.keys.Back):
		m.boardSwitcher = false
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.boardCursor = clamp(m.boardCursor-1, 0, max(len(m.boards)-1, 0))
	case key.Matches(msg, m.keys.Down):
		m.boardCursor = clamp(m.boardCursor+1, 0, max(len(m.boards)-1, 0))
	case key.Matches(msg, m.keys.Detail): // enter: open the highlighted board
		b, ok := m.highlightedBoard()
		m.boardSwitcher = false
		if !ok || b.ID == m.boardID {
			return m, nil
		}
		return m, m.switchBoardCmd(b.ID)
	case key.Matches(msg, m.keys.New):
		ti := textinput.New()
		ti.Placeholder = "Board name"
		ti.CharLimit = 100
		ti.Focus()
		m.boardNameInput = ti
		m.boardNaming = true
		return m, textinput.Blink
	case key.Matches(msg, m.keys.RenameColumn):
		b, ok := m.highlightedBoard()
		if !ok {
			return m, nil
		}
		ti := textinput.New()
		ti.Placeholder = "Board name"
		ti.CharLimit = 100
		ti.SetValue(b.Name)
		ti.CursorEnd()
		ti.Focus()
		m.boardNameInput = ti
		m.boardRenaming = true
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Delete):
		b, ok := m.highlightedBoard()
		if !ok {
			return m, nil
		}
		if len(m.boards) < 2 {
			return m, m.setToast("can't delete the only board")
		}
		switchTo := m.boardID
		if b.ID == m.boardID {
			switchTo = m.firstOtherBoard(b.ID)
		}
		m.confirm = &confirmState{
			prompt: fmt.Sprintf("Delete board %q and everything on it?", b.Name),
			action: m.deleteBoardCmd(b.ID, switchTo),
		}
		return m, nil
	}
	return m, nil
}

func (m Model) updateBoardNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.boardNaming = false
		m.boardRenaming = false
		return m, nil
	case tea.KeyEnter:
		name := strings.TrimSpace(m.boardNameInput.Value())
		naming := m.boardNaming
		m.boardNaming = false
		m.boardRenaming = false
		if name == "" {
			return m, nil
		}
		if naming {
			m.boardSwitcher = false
			return m, m.createBoardCmd(name)
		}
		b, ok := m.highlightedBoard()
		if !ok {
			return m, nil
		}
		return m, m.renameBoardCmd(b.ID, name)
	}

	var cmd tea.Cmd
	m.boardNameInput, cmd = m.boardNameInput.Update(msg)
	return m, cmd
}

func (m Model) highlightedBoard() (domain.Board, bool) {
	if m.boardCursor < 0 || m.boardCursor >= len(m.boards) {
		return domain.Board{}, false
	}
	return m.boards[m.boardCursor], true
}

func (m Model) firstOtherBoard(exclude int64) int64 {
	for _, b := range m.boards {
		if b.ID != exclude {
			return b.ID
		}
	}
	return exclude
}

// boardSwitcherView renders the full-screen board list.
func (m Model) boardSwitcherView() string {
	var b strings.Builder
	b.WriteString(m.styles.HelpHeading.Render("boards"))
	b.WriteByte('\n')

	for i, bd := range m.boards {
		row := bd.Name
		if bd.ID == m.boardID {
			row += m.styles.DetailDim.Render("  (current)")
		}
		if i == m.boardCursor {
			b.WriteString(m.styles.OverlayRowActive.Render("> " + row))
		} else {
			b.WriteString(m.styles.OverlayRow.Render(row))
		}
		b.WriteByte('\n')
	}

	if m.boardNaming {
		b.WriteString("\n" + m.styles.Prompt.Render("New board: ") + m.boardNameInput.View() + "\n")
	} else if m.boardRenaming {
		b.WriteString("\n" + m.styles.Prompt.Render("Rename board: ") + m.boardNameInput.View() + "\n")
	}

	b.WriteString(m.styles.DetailDim.Render("\nenter open · n new · R rename · d delete · b/esc close"))
	b.WriteString(m.confirmLine())

	box := m.styles.HelpOverlay.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
