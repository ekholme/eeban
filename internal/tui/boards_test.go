package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBoardSwitcherCreateAndSwitch(t *testing.T) {
	m := loadModel(t)

	m = runMutation(m, press("b"))
	if !m.boardSwitcher {
		t.Fatalf("'b' did not open the board switcher")
	}
	if len(m.boards) != 1 {
		t.Fatalf("board list = %+v", m.boards)
	}

	m = send(m, press("n"))
	if !m.boardNaming {
		t.Fatalf("'n' did not start naming a board")
	}
	m = typeRunes(m, "Side Project")
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.boardSwitcher {
		t.Fatalf("switcher should close after creating a board")
	}
	if m.board.Name != "Side Project" {
		t.Fatalf("did not switch to the new board: %q", m.board.Name)
	}
	if m.boardID == 1 {
		t.Fatalf("boardID still 1 after switching to the new board")
	}
	// The new board is seeded with the default columns and no cards.
	if len(m.board.Columns) != 3 {
		t.Fatalf("new board columns = %d, want 3", len(m.board.Columns))
	}
	if len(m.currentCards()) != 0 {
		t.Fatalf("new board should have no cards: %+v", m.currentCards())
	}
}

func TestBoardSwitcherSwitchesBack(t *testing.T) {
	m := loadModel(t)
	m = runMutation(m, press("b"))
	m = send(m, press("n"))
	m = typeRunes(m, "Second")
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Re-open, move to the first board, press enter.
	m = runMutation(m, press("b"))
	m = send(m, tea.KeyMsg{Type: tea.KeyUp})
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.boardID != 1 || m.board.Name != "My Board" {
		t.Fatalf("did not switch back to board 1: id=%d name=%q", m.boardID, m.board.Name)
	}
	if !strings.Contains(m.View(), "Welcome to eeban") {
		t.Fatalf("board 1 content not shown after switching back:\n%s", m.View())
	}
}
