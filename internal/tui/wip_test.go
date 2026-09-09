package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWIPLimitSetAndWarn(t *testing.T) {
	m := loadModel(t)

	m = send(m, press("w"))
	if !m.settingWIP {
		t.Fatalf("'w' did not open the WIP input")
	}
	m = typeRunes(m, "1")
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})

	col := m.board.Columns[0]
	if col.WIPLimit == nil || *col.WIPLimit != 1 {
		t.Fatalf("WIPLimit = %v, want 1", col.WIPLimit)
	}
	if !col.OverWIP() {
		t.Fatalf("Backlog has 2 cards over a limit of 1 but OverWIP() is false")
	}
	if !strings.Contains(m.View(), "2/1") {
		t.Fatalf("column header should show the over-limit count 2/1:\n%s", m.View())
	}

	// A blank value clears it.
	m = send(m, press("w"))
	m.wipInput.SetValue("")
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.board.Columns[0].WIPLimit != nil {
		t.Fatalf("WIP limit not cleared: %v", m.board.Columns[0].WIPLimit)
	}
}

func TestWIPLimitRejectsNonNumber(t *testing.T) {
	m := loadModel(t)
	m = send(m, press("w"))
	m = typeRunes(m, "abc")
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.board.Columns[0].WIPLimit != nil {
		t.Fatalf("non-numeric WIP input should not set a limit")
	}
	if m.settingWIP {
		t.Fatalf("WIP input still open after enter")
	}
	if !strings.Contains(m.toast, "number") {
		t.Fatalf("expected a toast about the value not being a number, got %q", m.toast)
	}
}
