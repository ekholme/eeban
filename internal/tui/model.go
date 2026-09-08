// Package tui is the Bubble Tea front end. It calls the service for state and
// renders it; it never touches SQL.
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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

	showDetail bool // detail pane visible for the selected card

	adding     bool // title input active for a new card
	titleInput textinput.Model

	err error // last mutation error, cleared on the next keypress

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

	case boardLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.board = msg.board
		if msg.selectCardID != nil {
			m.selectCardByID(*msg.selectCardID)
		} else {
			m.colCursor = clamp(m.colCursor, 0, m.lastColIndex())
			m.clampCardCursor()
		}
		return m, nil

	case tea.KeyMsg:
		if m.adding {
			return m.updateAdding(msg)
		}

		m.err = nil
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
		case key.Matches(msg, m.keys.Detail):
			m.showDetail = !m.showDetail
		case key.Matches(msg, m.keys.Back):
			m.showDetail = false
		case key.Matches(msg, m.keys.New):
			return m.startAdding()
		case key.Matches(msg, m.keys.Delete):
			return m.startDelete()
		case key.Matches(msg, m.keys.MoveLeft):
			return m.moveCardToColumn(-1)
		case key.Matches(msg, m.keys.MoveRight):
			return m.moveCardToColumn(1)
		case key.Matches(msg, m.keys.ReorderUp):
			return m.reorderCard(-1)
		case key.Matches(msg, m.keys.ReorderDown):
			return m.reorderCard(1)
		}
	}
	return m, nil
}

// startAdding opens the title input for a new card in the current column.
func (m Model) startAdding() (tea.Model, tea.Cmd) {
	if m.svc == nil || len(m.board.Columns) == 0 {
		return m, nil
	}
	ti := textinput.New()
	ti.Placeholder = "Card title"
	ti.CharLimit = 200
	ti.Focus()
	m.titleInput = ti
	m.adding = true
	return m, textinput.Blink
}

// updateAdding routes key input to the title textinput while adding a card.
func (m Model) updateAdding(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.adding = false
		return m, nil
	case tea.KeyEnter:
		title := strings.TrimSpace(m.titleInput.Value())
		m.adding = false
		if title == "" {
			return m, nil
		}
		columnID := m.board.Columns[m.colCursor].ID
		return m, m.createCardCmd(columnID, title)
	}

	var cmd tea.Cmd
	m.titleInput, cmd = m.titleInput.Update(msg)
	return m, cmd
}

// startDelete removes the selected card.
func (m Model) startDelete() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	card, ok := m.selectedCard()
	if !ok {
		return m, nil
	}
	return m, m.deleteCardCmd(card.ID)
}

// moveCardToColumn moves the selected card to the column dir steps away
// (-1 left, +1 right), appending it at the end.
func (m Model) moveCardToColumn(dir int) (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	card, ok := m.selectedCard()
	if !ok {
		return m, nil
	}
	target := m.colCursor + dir
	if target < 0 || target > m.lastColIndex() {
		return m, nil
	}
	toColumn := m.board.Columns[target]
	return m, m.moveCardCmd(card.ID, toColumn.ID, len(toColumn.Cards))
}

// reorderCard swaps the selected card with its neighbour dir steps away
// (-1 up, +1 down) within the current column.
func (m Model) reorderCard(dir int) (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	card, ok := m.selectedCard()
	if !ok {
		return m, nil
	}
	newIndex := m.cardCursor + dir
	if newIndex < 0 || newIndex >= len(m.currentCards()) {
		return m, nil
	}
	columnID := m.board.Columns[m.colCursor].ID
	return m, m.moveCardCmd(card.ID, columnID, newIndex)
}

// selectCardByID moves the cursor onto the card with the given id, if it is
// still present after a reload.
func (m *Model) selectCardByID(id int64) {
	for ci, col := range m.board.Columns {
		for cj, c := range col.Cards {
			if c.ID == id {
				m.colCursor, m.cardCursor = ci, cj
				return
			}
		}
	}
	m.colCursor = clamp(m.colCursor, 0, m.lastColIndex())
	m.clampCardCursor()
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

// selectedCard returns the card under the cursor, or false when the current
// column is empty.
func (m Model) selectedCard() (domain.Card, bool) {
	cards := m.currentCards()
	if m.cardCursor < 0 || m.cardCursor >= len(cards) {
		return domain.Card{}, false
	}
	return cards[m.cardCursor], true
}

func (m *Model) clampCardCursor() {
	m.cardCursor = clamp(m.cardCursor, 0, max(m.lastCardIndex(), 0))
}
