package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/domain"
)

func testBoard() domain.Board {
	return domain.Board{
		ID:   1,
		Name: "My Board",
		Columns: []domain.Column{
			{ID: 1, Name: "Backlog", Position: 1000, Cards: []domain.Card{
				{ID: 1, Title: "first"}, {ID: 2, Title: "second"},
			}},
			{ID: 2, Name: "In Progress", Position: 2000},
			{ID: 3, Name: "Done", Position: 3000, Cards: []domain.Card{
				{ID: 3, Title: "third"},
			}},
		},
	}
}

func send(m tea.Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func TestViewRendersColumns(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})

	out := m.View()
	for _, want := range []string{"My Board", "Backlog", "In Progress", "Done", "first", "third"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q\n%s", want, out)
		}
	}
}

func TestNavigationClampsAndFollowsColumns(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})

	// Left at the first column is a no-op.
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if m.colCursor != 0 {
		t.Fatalf("colCursor = %d, want 0", m.colCursor)
	}

	// Move down within Backlog, then right past the empty column to Done.
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cardCursor != 1 {
		t.Fatalf("cardCursor = %d, want 1", m.cardCursor)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.colCursor != 1 || m.cardCursor != 0 {
		t.Fatalf("after moving to empty column: col=%d card=%d, want 1/0", m.colCursor, m.cardCursor)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}) // clamp at last
	if m.colCursor != 2 {
		t.Fatalf("colCursor = %d, want 2 (clamped)", m.colCursor)
	}

	// Down is clamped to the single card in Done.
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cardCursor != 0 {
		t.Fatalf("cardCursor = %d, want 0 (clamped)", m.cardCursor)
	}
}
