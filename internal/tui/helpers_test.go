package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// loadModel builds a Model wired to a fresh seeded database and sized.
func loadModel(t *testing.T) Model {
	t.Helper()
	svc := newTestService(t)
	board, err := svc.DefaultBoard(context.Background())
	if err != nil {
		t.Fatalf("DefaultBoard: %v", err)
	}
	return send(New(svc, board), tea.WindowSizeMsg{Width: 140, Height: 40})
}

// press builds a rune key message for the given characters.
func press(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}
