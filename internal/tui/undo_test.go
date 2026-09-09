package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUndoMoveCard(t *testing.T) {
	m := loadModel(t)
	card, _ := m.selectedCard()

	m = runMutation(m, press("L")) // move right into In Progress
	if m.colCursor != 1 {
		t.Fatalf("card did not move right: colCursor=%d", m.colCursor)
	}
	if m.undo == nil {
		t.Fatalf("move did not record an undo step")
	}

	m = runMutation(m, press("u"))
	if m.undo != nil {
		t.Fatalf("undo step should be consumed")
	}
	back := false
	for _, c := range m.board.Columns[0].Cards {
		if c.ID == card.ID {
			back = true
		}
	}
	if !back {
		t.Fatalf("undo did not return the card to Backlog: %+v", m.board.Columns[0].Cards)
	}
}

func TestUndoArchive(t *testing.T) {
	m := loadModel(t)
	card, _ := m.selectedCard()

	m = runMutation(m, press("a"))
	if m.undo == nil {
		t.Fatalf("archive did not record an undo step")
	}
	m = runMutation(m, press("u"))

	back := false
	for _, c := range m.board.Columns[0].Cards {
		if c.ID == card.ID {
			back = true
		}
	}
	if !back {
		t.Fatalf("undo did not un-archive the card")
	}
}

func TestUndoCreateCard(t *testing.T) {
	m := loadModel(t)
	before := len(m.board.Columns[0].Cards)

	m = send(m, press("n"))
	m = typeRunes(m, "Throwaway")
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.board.Columns[0].Cards) != before+1 {
		t.Fatalf("card not created")
	}
	if m.undo == nil {
		t.Fatalf("create did not record an undo step")
	}

	m = runMutation(m, press("u"))
	if len(m.board.Columns[0].Cards) != before {
		t.Fatalf("undo did not remove the created card: %d, want %d", len(m.board.Columns[0].Cards), before)
	}
	if strings.Contains(m.View(), "Throwaway") {
		t.Fatalf("undone card still visible")
	}
}

func TestUndoWithNothingToUndo(t *testing.T) {
	m := loadModel(t)
	m = send(m, press("u"))
	if !strings.Contains(m.toast, "nothing to undo") {
		t.Fatalf("expected a 'nothing to undo' toast, got %q", m.toast)
	}
}
