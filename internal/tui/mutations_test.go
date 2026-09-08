package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/service"
	"github.com/ekholme/eeban/internal/store"
)

func newTestService(t *testing.T) *service.Service {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "eeban.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return service.New(db)
}

// runMutation drives msg through Update and, when it yields a tea.Cmd,
// executes that command synchronously and feeds the resulting message back
// in. That's enough to exercise the create/move/delete round trip without a
// real tea.Program driving the event loop.
func runMutation(m Model, msg tea.Msg) Model {
	next, cmd := m.Update(msg)
	nm := next.(Model)
	if cmd != nil {
		next2, _ := nm.Update(cmd())
		nm = next2.(Model)
	}
	return nm
}

// typeRunes feeds characters into the active textinput, discarding the
// cursor-blink commands each keystroke produces (they're irrelevant here and
// executing them would block on the blink interval).
func typeRunes(m Model, s string) Model {
	for _, r := range s {
		m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func TestCreateCardFlow(t *testing.T) {
	svc := newTestService(t)
	board, err := svc.DefaultBoard(context.Background())
	if err != nil {
		t.Fatalf("DefaultBoard: %v", err)
	}

	m := send(New(svc, board), tea.WindowSizeMsg{Width: 120, Height: 40})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if !m.adding {
		t.Fatalf("expected adding mode after 'n'")
	}

	m = typeRunes(m, "WriteTests")
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.adding {
		t.Fatalf("still in adding mode after enter")
	}
	if m.err != nil {
		t.Fatalf("unexpected error: %v", m.err)
	}
	if !strings.Contains(m.View(), "WriteTests") {
		t.Fatalf("new card not visible:\n%s", m.View())
	}
	if c, ok := m.selectedCard(); !ok || c.Title != "WriteTests" {
		t.Fatalf("cursor did not follow new card: %+v, %v", c, ok)
	}
}

func TestCreateCardEmptyTitleIsNoop(t *testing.T) {
	svc := newTestService(t)
	board, _ := svc.DefaultBoard(context.Background())
	m := send(New(svc, board), tea.WindowSizeMsg{Width: 120, Height: 40})

	before := len(m.currentCards())
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.adding {
		t.Fatalf("still in adding mode")
	}
	if len(m.currentCards()) != before {
		t.Fatalf("card count changed on empty title: got %d, want %d", len(m.currentCards()), before)
	}
}

func TestDeleteCardFlow(t *testing.T) {
	svc := newTestService(t)
	board, _ := svc.DefaultBoard(context.Background())
	m := send(New(svc, board), tea.WindowSizeMsg{Width: 120, Height: 40})

	before := len(m.currentCards())
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if m.err != nil {
		t.Fatalf("unexpected error: %v", m.err)
	}
	if len(m.currentCards()) != before-1 {
		t.Fatalf("card count = %d, want %d", len(m.currentCards()), before-1)
	}
	if strings.Contains(m.View(), "Welcome to eeban") {
		t.Fatalf("deleted card still visible:\n%s", m.View())
	}
}

func TestMoveCardAcrossColumnsFlow(t *testing.T) {
	svc := newTestService(t)
	board, _ := svc.DefaultBoard(context.Background())
	m := send(New(svc, board), tea.WindowSizeMsg{Width: 120, Height: 40})

	card, ok := m.selectedCard()
	if !ok {
		t.Fatalf("no card selected initially")
	}

	m = runMutation(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")})

	if m.err != nil {
		t.Fatalf("unexpected error: %v", m.err)
	}
	if m.colCursor != 1 {
		t.Fatalf("colCursor = %d, want 1 after moving right", m.colCursor)
	}
	got, ok := m.selectedCard()
	if !ok || got.ID != card.ID {
		t.Fatalf("cursor did not follow moved card: %+v, %v", got, ok)
	}
	if got.ColumnID != m.board.Columns[1].ID {
		t.Fatalf("card column_id = %d, want %d", got.ColumnID, m.board.Columns[1].ID)
	}
}

func TestReorderCardFlow(t *testing.T) {
	svc := newTestService(t)
	board, _ := svc.DefaultBoard(context.Background())
	m := send(New(svc, board), tea.WindowSizeMsg{Width: 120, Height: 40})

	first, ok := m.selectedCard()
	if !ok {
		t.Fatalf("no card selected initially")
	}

	m = runMutation(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("J")})

	if m.err != nil {
		t.Fatalf("unexpected error: %v", m.err)
	}
	if m.cardCursor != 1 {
		t.Fatalf("cardCursor = %d, want 1 after reordering down", m.cardCursor)
	}
	cards := m.currentCards()
	if len(cards) < 2 || cards[1].ID != first.ID {
		t.Fatalf("card not moved to index 1: %+v", cards)
	}
}

func TestMutationsAreNoopWithNilService(t *testing.T) {
	m := New(nil, testBoard())
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	before := m
	for _, key := range []string{"n", "d", "H", "L", "J", "K"} {
		next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		nm := next.(Model)
		if cmd != nil {
			t.Fatalf("key %q produced a cmd with nil service", key)
		}
		if nm.adding {
			t.Fatalf("key %q entered adding mode with nil service", key)
		}
		m = nm
	}
	if len(m.board.Columns) != len(before.board.Columns) {
		t.Fatalf("board mutated with nil service")
	}
}
