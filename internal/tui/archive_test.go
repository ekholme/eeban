package tui

import (
	"strings"
	"testing"
)

func TestArchiveRestoreRoundTrip(t *testing.T) {
	m := loadModel(t)
	first, _ := m.selectedCard()

	m = runMutation(m, press("a"))
	for _, c := range m.board.Columns[0].Cards {
		if c.ID == first.ID {
			t.Fatalf("archived card still on the board")
		}
	}

	// Open the archive view; it should list the card.
	m = runMutation(m, press("A"))
	if !m.showArchive {
		t.Fatalf("'A' did not open the archive view")
	}
	if len(m.archive) != 1 || m.archive[0].ID != first.ID {
		t.Fatalf("archive list = %+v", m.archive)
	}
	if !strings.Contains(m.View(), first.Title) {
		t.Fatalf("archive view missing the card:\n%s", m.View())
	}

	// Restore it and refresh the list.
	m = runMutation(m, press("r"))
	m = send(m, m.loadArchiveCmd()())
	if len(m.archive) != 0 {
		t.Fatalf("archive list not empty after restore: %+v", m.archive)
	}

	m.showArchive = false
	found := false
	for _, c := range m.board.Columns[0].Cards {
		if c.ID == first.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("restored card not back in its column")
	}
}
