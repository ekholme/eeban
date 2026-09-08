package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpOverlayToggles(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})

	if strings.Contains(m.View(), "keybindings") {
		t.Fatalf("help overlay visible before '?'")
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if !m.showHelp {
		t.Fatalf("'?' did not open the help overlay")
	}
	out := m.View()
	for _, want := range []string{"keybindings", "new card", "rename column", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("help overlay missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "In Progress") {
		t.Errorf("board still rendered behind the help overlay\n%s", out)
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.showHelp {
		t.Fatalf("esc did not close the help overlay")
	}

	// '?' also closes it again.
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if m.showHelp {
		t.Fatalf("second '?' did not close the help overlay")
	}
}

func TestHelpOverlaySwallowsOtherKeys(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.colCursor != 0 {
		t.Fatalf("navigation ran while help was open: colCursor = %d", m.colCursor)
	}
	if !m.showHelp {
		t.Fatalf("an unrelated key closed the help overlay")
	}
}

func TestDeleteCardAsksForConfirmation(t *testing.T) {
	svc := newTestService(t)
	board, _ := svc.DefaultBoard(context.Background())
	m := send(New(svc, board), tea.WindowSizeMsg{Width: 120, Height: 40})

	before := len(m.currentCards())

	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	if m.confirm == nil {
		t.Fatalf("'d' did not open a confirm dialog")
	}
	if !strings.Contains(m.View(), "Delete card") {
		t.Fatalf("confirm prompt not shown in the footer:\n%s", m.View())
	}

	// Answering "n" dismisses without deleting.
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if m.confirm != nil {
		t.Fatalf("'n' did not dismiss the confirm dialog")
	}
	if len(m.currentCards()) != before {
		t.Fatalf("card deleted despite cancelling: %d, want %d", len(m.currentCards()), before)
	}

	// Answering "y" runs the delete.
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if m.confirm != nil {
		t.Fatalf("confirm dialog still open after 'y'")
	}
	if m.err != nil {
		t.Fatalf("unexpected error: %v", m.err)
	}
	if len(m.currentCards()) != before-1 {
		t.Fatalf("card count = %d, want %d", len(m.currentCards()), before-1)
	}
}

func TestDeleteColumnConfirmCancelIsNoop(t *testing.T) {
	svc := newTestService(t)
	board, _ := svc.DefaultBoard(context.Background())
	m := send(New(svc, board), tea.WindowSizeMsg{Width: 120, Height: 40})

	before := len(m.board.Columns)
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	if m.confirm == nil {
		t.Fatalf("'D' did not open a confirm dialog")
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.confirm != nil {
		t.Fatalf("esc did not dismiss the confirm dialog")
	}
	if len(m.board.Columns) != before {
		t.Fatalf("column deleted despite cancelling: %d, want %d", len(m.board.Columns), before)
	}
}

func TestErrorToastShowsAndExpires(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})

	next, cmd := m.Update(boardLoadedMsg{err: errors.New("boom")})
	m = next.(Model)
	if m.toast == "" || !strings.Contains(m.View(), "boom") {
		t.Fatalf("error toast not shown:\n%s", m.View())
	}
	if cmd == nil {
		t.Fatalf("no expiry command returned for the toast")
	}

	// A stale timer must not clear a toast that has since been replaced.
	m = send(m, boardLoadedMsg{err: errors.New("second")})
	m = send(m, toastExpiredMsg{seq: m.toastSeq - 1})
	if m.toast == "" {
		t.Fatalf("a stale timer cleared the current toast")
	}

	// The matching timer clears it.
	m = send(m, toastExpiredMsg{seq: m.toastSeq})
	if m.toast != "" {
		t.Fatalf("toast not cleared by its own timer: %q", m.toast)
	}
}

func TestSuccessfulReloadClearsToast(t *testing.T) {
	m := send(New(nil, testBoard()), tea.WindowSizeMsg{Width: 120, Height: 40})
	m = send(m, boardLoadedMsg{err: errors.New("boom")})
	if m.toast == "" {
		t.Fatalf("precondition: toast should be set")
	}

	m = send(m, boardLoadedMsg{board: testBoard()})
	if m.toast != "" {
		t.Fatalf("successful reload left the toast up: %q", m.toast)
	}
}
