// Package tui is the Bubble Tea front end. It calls the service for state and
// renders it; it never touches SQL.
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/domain"
	"github.com/ekholme/eeban/internal/service"
)

// Model is the root Bubble Tea model for eeban.
type Model struct {
	svc   *service.Service
	board domain.Board

	width  int
	height int

	colCursor  int // index into board.Columns
	cardCursor int // index into the selected column's cards

	keys   KeyMap
	styles Styles
}

// New builds the root model for an already-loaded board.
func New(svc *service.Service, board domain.Board) Model {
	return Model{
		svc:    svc,
		board:  board,
		keys:   DefaultKeyMap(),
		styles: DefaultStyles(),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			m.colCursor = clamp(m.colCursor-1, 0, m.lastColIndex())
			m.clampCardCursor()
		case key.Matches(msg, m.keys.Right):
			m.colCursor = clamp(m.colCursor+1, 0, m.lastColIndex())
			m.clampCardCursor()
		case key.Matches(msg, m.keys.Up):
			m.cardCursor = clamp(m.cardCursor-1, 0, m.lastCardIndex())
		case key.Matches(msg, m.keys.Down):
			m.cardCursor = clamp(m.cardCursor+1, 0, m.lastCardIndex())
		}
	}
	return m, nil
}

func (m Model) lastColIndex() int {
	return len(m.board.Columns) - 1
}

func (m Model) currentCards() []domain.Card {
	if m.colCursor < 0 || m.colCursor >= len(m.board.Columns) {
		return nil
	}
	return m.board.Columns[m.colCursor].Cards
}

func (m Model) lastCardIndex() int {
	return len(m.currentCards()) - 1
}

func (m *Model) clampCardCursor() {
	m.cardCursor = clamp(m.cardCursor, 0, max(m.lastCardIndex(), 0))
}
