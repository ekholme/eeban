package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLabelCreateAttachAndFilter(t *testing.T) {
	m := loadModel(t)

	m = send(m, press("t"))
	if !m.labelPicker {
		t.Fatalf("'t' did not open the label picker")
	}

	// Create a label.
	m = send(m, press("n"))
	if !m.labelNaming {
		t.Fatalf("'n' did not start naming a label")
	}
	m = typeRunes(m, "urgent")
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.board.Labels) != 1 || m.board.Labels[0].Name != "urgent" {
		t.Fatalf("board labels after create = %+v", m.board.Labels)
	}
	if m.board.Labels[0].Color == "" {
		t.Fatalf("new label was not assigned a colour")
	}

	// Attach it to the selected card.
	m = runMutation(m, tea.KeyMsg{Type: tea.KeySpace})
	card, ok := m.selectedCard()
	if !ok || !card.HasLabel(m.board.Labels[0].ID) {
		t.Fatalf("space did not attach the label: %+v", card)
	}

	// Filter the board by that label.
	m = send(m, press("f"))
	if m.labelPicker {
		t.Fatalf("'f' should close the picker")
	}
	if m.filterLabel != m.board.Labels[0].ID {
		t.Fatalf("filterLabel = %d, want %d", m.filterLabel, m.board.Labels[0].ID)
	}
	for _, c := range m.currentCards() {
		if !c.HasLabel(m.filterLabel) {
			t.Fatalf("label filter let through an unlabelled card: %+v", c)
		}
	}

	// Detach toggles back off.
	m = send(m, press("t"))
	m = runMutation(m, tea.KeyMsg{Type: tea.KeySpace})
	if card, _ := m.selectedCard(); card.HasLabel(m.board.Labels[0].ID) {
		t.Fatalf("second space did not detach the label")
	}
}
