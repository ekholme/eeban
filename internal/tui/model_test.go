package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/domain"
)

func testBoard() domain.Board {
	due := "2026-09-15"
	return domain.Board{
		ID:   1,
		Name: "My Board",
		Columns: []domain.Column{
			{ID: 1, Name: "Backlog", Position: 1000, Cards: []domain.Card{
				{ID: 1, Title: "first", Body: "the first card body", Priority: domain.PriorityHigh, DueDate: &due},
				{ID: 2, Title: "second"},
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

func TestDetailPaneToggles(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})

	if strings.Contains(m.View(), "Priority") {
		t.Fatalf("detail pane visible before toggle:\n%s", m.View())
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	out := m.View()
	for _, want := range []string{"Priority", "High", "Due", "2026-09-15", "the first card body"} {
		if !strings.Contains(out, want) {
			t.Errorf("open detail pane missing %q\n%s", want, out)
		}
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if strings.Contains(m.View(), "Priority") {
		t.Fatalf("detail pane still visible after esc:\n%s", m.View())
	}

	// enter also closes it again.
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if strings.Contains(m.View(), "Priority") {
		t.Fatalf("detail pane still visible after second enter:\n%s", m.View())
	}
}

func TestDetailPaneTracksSelection(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	if c, ok := m.selectedCard(); !ok || c.Title != "first" {
		t.Fatalf("selectedCard = %+v, %v; want title \"first\"", c, ok)
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if c, _ := m.selectedCard(); c.Title != "second" {
		t.Fatalf("selectedCard title = %q, want \"second\"", c.Title)
	}
	if !strings.Contains(m.View(), "None") {
		t.Errorf("detail pane should show priority None for card \"second\"\n%s", m.View())
	}
}

func TestDetailPaneEmptyColumn(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}) // In Progress, no cards
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	if !strings.Contains(m.View(), "No card selected") {
		t.Fatalf("expected empty-state detail pane:\n%s", m.View())
	}
}

func TestPriorityLabel(t *testing.T) {
	cases := map[int]string{
		domain.PriorityNone:   "None",
		domain.PriorityLow:    "Low",
		domain.PriorityMedium: "Medium",
		domain.PriorityHigh:   "High",
		99:                    "None",
	}
	for p, want := range cases {
		if got := priorityLabel(p); got != want {
			t.Errorf("priorityLabel(%d) = %q, want %q", p, got, want)
		}
	}
}
