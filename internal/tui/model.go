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
	svc     *service.Service
	board   domain.Board
	boardID int64

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

	settingWIP bool // numeric input active for the current column's WIP limit
	wipInput   textinput.Model

	filtering   bool // search input active
	filterInput textinput.Model
	filter      string // committed fuzzy query over title + body
	filterLabel int64  // when non-zero, only cards carrying this label show

	labelPicker    bool // label overlay open for the selected card
	labelCursor    int
	labelNaming    bool // sub-input: name for a new label
	labelNameInput textinput.Model

	showArchive   bool // full-screen archived-card list
	archive       []domain.Card
	archiveCursor int

	boardSwitcher  bool // full-screen board list
	boards         []domain.Board
	boardCursor    int
	boardNaming    bool // sub-input: name for a new board
	boardRenaming  bool // sub-input: rename the highlighted board
	boardNameInput textinput.Model

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
		svc:     svc,
		board:   board,
		boardID: board.ID,
		keys:    DefaultKeyMap(),
		styles:  DefaultStyles(),
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
		return m.onBoardLoaded(msg)

	case archiveLoadedMsg:
		if msg.err != nil {
			return m, m.setToast("error: " + msg.err.Error())
		}
		m.archive = msg.cards
		m.archiveCursor = clamp(m.archiveCursor, 0, max(len(m.archive)-1, 0))
		return m, nil

	case boardsLoadedMsg:
		if msg.err != nil {
			return m, m.setToast("error: " + msg.err.Error())
		}
		m.boards = msg.boards
		m.syncBoardCursor()
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
		if m.settingWIP {
			return m.updateWIPInput(msg)
		}
		if m.filtering {
			return m.updateFiltering(msg)
		}
		if m.labelPicker {
			return m.updateLabelPicker(msg)
		}
		if m.boardSwitcher {
			return m.updateBoardSwitcher(msg)
		}
		if m.showArchive {
			return m.updateArchive(msg)
		}
		return m.updateBoard(msg)
	}
	return m, nil
}

// updateBoard handles keys on the main board view.
func (m Model) updateBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.filterActive() {
			m.filter = ""
			m.filterLabel = 0
			m.clampToFilter()
			return m, nil
		}
		m.showDetail = false
	case key.Matches(msg, m.keys.Labels):
		return m.openLabelPicker()
	case key.Matches(msg, m.keys.Search):
		return m.startFiltering()
	case key.Matches(msg, m.keys.New):
		return m.startAdding()
	case key.Matches(msg, m.keys.Edit):
		return m.startEdit()
	case key.Matches(msg, m.keys.Delete):
		return m.startDelete()
	case key.Matches(msg, m.keys.Archive):
		return m.archiveSelected()
	case key.Matches(msg, m.keys.ArchiveView):
		return m.openArchive()
	case key.Matches(msg, m.keys.Boards):
		return m.openBoardSwitcher()
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
	case key.Matches(msg, m.keys.WIPLimit):
		return m.startSettingWIP()
	case key.Matches(msg, m.keys.ColumnLeft):
		return m.reorderColumn(-1)
	case key.Matches(msg, m.keys.ColumnRight):
		return m.reorderColumn(1)
	}
	return m, nil
}

// onBoardLoaded folds a board reload into the model: it swaps in the fresh
// board and restores the cursor onto whichever item the mutation touched.
func (m Model) onBoardLoaded(msg boardLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, m.setToast("error: " + msg.err.Error())
	}

	m.err = nil
	m.toast = ""
	m.board = msg.board
	m.boardID = msg.board.ID

	switch {
	case msg.resetCursor:
		m.colCursor, m.cardCursor = 0, 0
		m.filter, m.filterLabel = "", 0
	case msg.selectCardID != nil:
		m.selectCardByID(*msg.selectCardID)
	case msg.selectColumnID != nil:
		m.selectColumnByID(*msg.selectColumnID)
	default:
		m.colCursor = clamp(m.colCursor, 0, m.lastColIndex())
		m.clampCardCursor()
	}

	// Keep the archive list fresh while it's on screen.
	if m.showArchive {
		return m, m.loadArchiveCmd()
	}
	return m, nil
}

// archiveSelected moves the selected card into the archive.
func (m Model) archiveSelected() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	card, ok := m.selectedCard()
	if !ok {
		return m, nil
	}
	return m, m.archiveCardCmd(card.ID)
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

// startSettingWIP opens the numeric input for the current column's WIP limit.
func (m Model) startSettingWIP() (tea.Model, tea.Cmd) {
	if m.svc == nil || len(m.board.Columns) == 0 {
		return m, nil
	}
	ti := textinput.New()
	ti.Placeholder = "WIP limit (blank clears)"
	ti.CharLimit = 3
	if lim := m.board.Columns[m.colCursor].WIPLimit; lim != nil {
		ti.SetValue(fmt.Sprintf("%d", *lim))
		ti.CursorEnd()
	}
	ti.Focus()
	m.wipInput = ti
	m.settingWIP = true
	return m, textinput.Blink
}

// updateWIPInput routes keys to the WIP limit input. Enter commits; a blank
// or zero value clears the limit; a non-number is rejected with a toast.
func (m Model) updateWIPInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.settingWIP = false
		return m, nil
	case tea.KeyEnter:
		raw := strings.TrimSpace(m.wipInput.Value())
		m.settingWIP = false
		columnID := m.board.Columns[m.colCursor].ID
		if raw == "" {
			return m, m.setColumnWIPCmd(columnID, nil)
		}
		n, err := parsePositiveInt(raw)
		if err != nil {
			return m, m.setToast("WIP limit must be a number")
		}
		if n == 0 {
			return m, m.setColumnWIPCmd(columnID, nil)
		}
		return m, m.setColumnWIPCmd(columnID, &n)
	}

	var cmd tea.Cmd
	m.wipInput, cmd = m.wipInput.Update(msg)
	return m, cmd
}

// startFiltering opens the fuzzy-search input, seeded with the active query.
func (m Model) startFiltering() (tea.Model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "fuzzy search title + body"
	ti.CharLimit = 100
	ti.SetValue(m.filter)
	ti.CursorEnd()
	ti.Focus()
	m.filterInput = ti
	m.filtering = true
	return m, textinput.Blink
}

// updateFiltering routes keys to the search input, committing the query live
// so the board narrows as you type.
func (m Model) updateFiltering(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filtering = false
		m.filter = ""
		m.clampToFilter()
		return m, nil
	case tea.KeyEnter:
		m.filtering = false
		m.filter = strings.TrimSpace(m.filterInput.Value())
		m.clampToFilter()
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.filter = strings.TrimSpace(m.filterInput.Value())
	m.clampToFilter()
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
	if m.filterActive() {
		return m, m.setToast("clear the filter to reorder cards")
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
// still present (and visible under the active filter) after a reload.
func (m *Model) selectCardByID(id int64) {
	for ci, col := range m.board.Columns {
		for cj, c := range m.visibleCards(col) {
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

// filterActive reports whether any filter is narrowing the board.
func (m Model) filterActive() bool {
	return m.filter != "" || m.filterLabel != 0
}

// visibleCards returns col's cards after applying the active text and label
// filters. With no filter it returns the column's cards unchanged.
func (m Model) visibleCards(col domain.Column) []domain.Card {
	if !m.filterActive() {
		return col.Cards
	}
	out := make([]domain.Card, 0, len(col.Cards))
	for _, c := range col.Cards {
		if m.filterLabel != 0 && !c.HasLabel(m.filterLabel) {
			continue
		}
		if !c.CardMatches(m.filter) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (m Model) currentCards() []domain.Card {
	if m.colCursor < 0 || m.colCursor >= len(m.board.Columns) {
		return nil
	}
	return m.visibleCards(m.board.Columns[m.colCursor])
}

func (m Model) lastCardIndex() int {
	return len(m.currentCards()) - 1
}

// selectedCard returns the card under the cursor, or false when the current
// column has no visible cards.
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

// clampToFilter re-homes both cursors after the visible set changes.
func (m *Model) clampToFilter() {
	m.colCursor = clamp(m.colCursor, 0, max(m.lastColIndex(), 0))
	m.clampCardCursor()
}
