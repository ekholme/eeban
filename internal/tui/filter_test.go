package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchFilterNarrowsBoard(t *testing.T) {
	m := loadModel(t)

	m = send(m, press("/"))
	if !m.filtering {
		t.Fatalf("'/' did not open the search input")
	}
	m = typeRunes(m, "welcome")
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.filter != "welcome" {
		t.Fatalf("committed filter = %q, want welcome", m.filter)
	}
	if got := m.currentCards(); len(got) != 1 || got[0].Title != "Welcome to eeban" {
		t.Fatalf("Backlog visible cards under filter = %+v", got)
	}
	out := m.View()
	if !strings.Contains(out, "no matches") {
		t.Fatalf("columns without a match should say so:\n%s", out)
	}

	// esc on the board clears the filter.
	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.filterActive() {
		t.Fatalf("esc did not clear the filter")
	}
	if len(m.currentCards()) != 2 {
		t.Fatalf("filter still applied after clear: %+v", m.currentCards())
	}
}

func TestReorderBlockedWhileFiltering(t *testing.T) {
	m := loadModel(t)
	m = send(m, press("/"))
	m = typeRunes(m, "e")
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	m2, _ := m.Update(press("J"))
	if got := m2.(Model); got.toast == "" {
		t.Fatalf("expected a toast explaining reorder is blocked while filtering")
	}
}
