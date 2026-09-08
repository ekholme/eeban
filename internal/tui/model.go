// Package tui is the Bubble Tea front end. It calls the service for state and
// renders it; it never touches SQL.
package tui

import (
	"fmt"
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

	editing bool // edit form active for the selected card
	form    editForm

	addingColumn   bool // name input active for a new column
	renamingColumn bool // name input active for renaming the current column
	columnInput    textinput.Model

	err error // last mutation error, surfaced as a toast

	toast    string // transient notification text, "" when hidden
	toastSeq int    // bumped per toast so a stale timer can't clear a newer one

	showHelp bool          // full-screen keybinding overlay
	confirm  *confirmState // pending y/n confirmation, nil when none

	keys   KeyMap
	styles Styles
}

// confirmState is a pending yes/no confirmation. action is the command run
// when the user answers "y".
type confirmState struct {
	prompt string
	action tea.Cmd
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
			cmd := m.setToast("error: " + msg.err.Error())
			return m, cmd
		}
		m.err = nil
		m.toast = ""
		m.board = msg.board
		switch {
		case msg.selectCardID != nil:
			m.selectCardByID(*msg.selectCardID)
		case msg.selectColumnID != nil:
			m.selectColumnByID(*msg.selectColumnID)
		default:
			m.colCursor = clamp(m.colCursor, 0, m.lastColIndex())
			m.clampCardCursor()
		}
		return m, nil

	case toastExpiredMsg:
		if msg.seq == m.toastSeq {
			m.toast = ""
			m.err = nil
		}
		return m, nil

	case tea.KeyMsg:
		if m.confirm != nil {
			return m.updateConfirm(msg)
		}
		if m.showHelp {
			return m.updateHelp(msg)
		}
		if m.adding {
			return m.updateAdding(msg)
		}
		if m.editing {
			return m.updateEditForm(msg)
		}
		if m.addingColumn || m.renamingColumn {
			return m.updateColumnInput(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = true
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
		case key.Matches(msg, m.keys.Edit):
			return m.startEdit()
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
		case key.Matches(msg, m.keys.NewColumn):
			return m.startAddingColumn()
		case key.Matches(msg, m.keys.RenameColumn):
			return m.startRenamingColumn()
		case key.Matches(msg, m.keys.DeleteColumn):
			return m.startDeleteColumn()
		case key.Matches(msg, m.keys.ColumnLeft):
			return m.reorderColumn(-1)
		case key.Matches(msg, m.keys.ColumnRight):
			return m.reorderColumn(1)
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

// startEdit opens the full edit form (title, body, priority, due date) for
// the selected card.
func (m Model) startEdit() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	card, ok := m.selectedCard()
	if !ok {
		return m, nil
	}
	m.form = newEditForm(card)
	m.editing = true
	return m, textinput.Blink
}

// updateEditForm routes key input to the edit form while it's active,
// submitting or cancelling as directed.
func (m Model) updateEditForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form, cmd, action := m.form.update(msg)
	m.form = form

	switch action {
	case formCancel:
		m.editing = false
		return m, nil
	case formSubmit:
		m.editing = false
		title := strings.TrimSpace(m.form.title.Value())
		if title == "" {
			return m, nil
		}
		return m, m.updateCardCmd(m.form.cardID, title, m.form.body.Value(), m.form.priority, m.form.dueDatePtr())
	}
	return m, cmd
}

// startAddingColumn opens the name input for a new column at the end of the
// board.
func (m Model) startAddingColumn() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	ti := textinput.New()
	ti.Placeholder = "Column name"
	ti.CharLimit = 100
	ti.Focus()
	m.columnInput = ti
	m.addingColumn = true
	return m, textinput.Blink
}

// startRenamingColumn opens the name input pre-filled with the current
// column's name.
func (m Model) startRenamingColumn() (tea.Model, tea.Cmd) {
	if m.svc == nil || len(m.board.Columns) == 0 {
		return m, nil
	}
	ti := textinput.New()
	ti.Placeholder = "Column name"
	ti.CharLimit = 100
	ti.SetValue(m.board.Columns[m.colCursor].Name)
	ti.CursorEnd()
	ti.Focus()
	m.columnInput = ti
	m.renamingColumn = true
	return m, textinput.Blink
}

// updateColumnInput routes key input to the column name textinput while
// adding or renaming a column.
func (m Model) updateColumnInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.addingColumn = false
		m.renamingColumn = false
		return m, nil
	case tea.KeyEnter:
		name := strings.TrimSpace(m.columnInput.Value())
		adding := m.addingColumn
		m.addingColumn = false
		m.renamingColumn = false
		if name == "" {
			return m, nil
		}
		if adding {
			return m, m.createColumnCmd(m.board.ID, name)
		}
		columnID := m.board.Columns[m.colCursor].ID
		return m, m.renameColumnCmd(columnID, name)
	}

	var cmd tea.Cmd
	m.columnInput, cmd = m.columnInput.Update(msg)
	return m, cmd
}

// startDeleteColumn asks to confirm removing the current column and its cards.
func (m Model) startDeleteColumn() (tea.Model, tea.Cmd) {
	if m.svc == nil || len(m.board.Columns) == 0 {
		return m, nil
	}
	col := m.board.Columns[m.colCursor]
	m.confirm = &confirmState{
		prompt: fmt.Sprintf("Delete column %q and its cards?", col.Name),
		action: m.deleteColumnCmd(col.ID),
	}
	return m, nil
}

// reorderColumn swaps the current column with its neighbour dir steps away
// (-1 left, +1 right).
func (m Model) reorderColumn(dir int) (tea.Model, tea.Cmd) {
	if m.svc == nil || len(m.board.Columns) == 0 {
		return m, nil
	}
	col := m.board.Columns[m.colCursor]
	newIndex := m.colCursor + dir
	if newIndex < 0 || newIndex >= len(m.board.Columns) {
		return m, nil
	}
	return m, m.moveColumnCmd(m.board.ID, col.ID, newIndex)
}

// startDelete asks to confirm removing the selected card.
func (m Model) startDelete() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	card, ok := m.selectedCard()
	if !ok {
		return m, nil
	}
	m.confirm = &confirmState{
		prompt: fmt.Sprintf("Delete card %q?", truncateText(card.Title, 40)),
		action: m.deleteCardCmd(card.ID),
	}
	return m, nil
}

// updateConfirm handles keys while a yes/no confirmation is pending: "y" (or
// enter) runs the pending action, anything else dismisses it.
func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		action := m.confirm.action
		m.confirm = nil
		return m, action
	default:
		m.confirm = nil
		return m, nil
	}
}

// updateHelp handles keys while the help overlay is open. ctrl+c still quits;
// "?" or esc closes it; every other key is swallowed.
func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help), key.Matches(msg, m.keys.Back):
		m.showHelp = false
	}
	return m, nil
}

// setToast shows text as a transient notification and returns the command
// that clears it after toastDuration, unless a newer toast has replaced it.
func (m *Model) setToast(text string) tea.Cmd {
	m.toast = text
	m.toastSeq++
	return toastExpireCmd(m.toastSeq)
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

// selectColumnByID moves the cursor onto the column with the given id, if it
// is still present after a reload.
func (m *Model) selectColumnByID(id int64) {
	for i, col := range m.board.Columns {
		if col.ID == id {
			m.colCursor = i
			m.clampCardCursor()
			return
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
