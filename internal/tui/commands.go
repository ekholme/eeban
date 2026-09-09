package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/domain"
	"github.com/ekholme/eeban/internal/service"
)

// toastDuration is how long a transient toast stays on screen.
const toastDuration = 4 * time.Second

// toastExpiredMsg fires toastDuration after a toast is shown. seq guards
// against an older timer clearing a toast that has since been replaced.
type toastExpiredMsg struct{ seq int }

func toastExpireCmd(seq int) tea.Cmd {
	return tea.Tick(toastDuration, func(time.Time) tea.Msg {
		return toastExpiredMsg{seq: seq}
	})
}

// boardLoadedMsg carries the result of a board reload, optionally after a
// mutation. When err is non-nil, board is meaningless and the current model
// state is left untouched. selectCardID / selectColumnID, when set, tell the
// model which item the cursor should follow after the reload; resetCursor
// sends the cursor home instead (used when switching boards).
type boardLoadedMsg struct {
	board          domain.Board
	selectCardID   *int64
	selectColumnID *int64
	resetCursor    bool
	err            error
}

// archiveLoadedMsg carries the archived-card list for the archive view.
type archiveLoadedMsg struct {
	cards []domain.Card
	err   error
}

// boardsLoadedMsg carries the board list for the board switcher.
type boardsLoadedMsg struct {
	boards []domain.Board
	err    error
}

// loadBoardMsg reloads m.boardID and packages it as a boardLoadedMsg.
func loadBoardMsg(svc *service.Service, boardID int64, selCard, selCol *int64) tea.Msg {
	board, err := svc.Board(context.Background(), boardID)
	return boardLoadedMsg{board: board, selectCardID: selCard, selectColumnID: selCol, err: err}
}

func (m Model) reloadCmd(selCard, selCol *int64) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg { return loadBoardMsg(svc, boardID, selCard, selCol) }
}

func (m Model) createCardCmd(columnID int64, title string) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		card, err := svc.CreateCard(ctx, columnID, title)
		if err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, &card.ID, nil)
	}
}

func (m Model) updateCardCmd(cardID int64, title, body string, priority int, dueDate *string) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.UpdateCard(ctx, cardID, title, body, priority, dueDate); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, &cardID, nil)
	}
}

func (m Model) deleteCardCmd(cardID int64) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.DeleteCard(ctx, cardID); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, nil)
	}
}

func (m Model) moveCardCmd(cardID, toColumnID int64, toIndex int) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.MoveCard(ctx, cardID, toColumnID, toIndex); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, &cardID, nil)
	}
}

func (m Model) archiveCardCmd(cardID int64) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.ArchiveCard(ctx, cardID); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, nil)
	}
}

func (m Model) unarchiveCardCmd(cardID int64) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.UnarchiveCard(ctx, cardID); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, &cardID, nil)
	}
}

func (m Model) loadArchiveCmd() tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		cards, err := svc.ArchivedCards(context.Background(), boardID)
		return archiveLoadedMsg{cards: cards, err: err}
	}
}

func (m Model) setColumnWIPCmd(columnID int64, limit *int) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.SetColumnWIP(ctx, columnID, limit); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, &columnID)
	}
}

func (m Model) createLabelCmd(name, color string) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if _, err := svc.CreateLabel(ctx, boardID, name, color); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, nil)
	}
}

func (m Model) deleteLabelCmd(labelID int64) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.DeleteLabel(ctx, labelID); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, nil)
	}
}

func (m Model) setCardLabelCmd(cardID, labelID int64, on bool) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.SetCardLabel(ctx, cardID, labelID, on); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, &cardID, nil)
	}
}

func (m Model) loadBoardsCmd() tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		boards, err := svc.Boards(context.Background())
		return boardsLoadedMsg{boards: boards, err: err}
	}
}

// switchBoardCmd loads a different board and homes the cursor.
func (m Model) switchBoardCmd(boardID int64) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		board, err := svc.Board(context.Background(), boardID)
		return boardLoadedMsg{board: board, resetCursor: true, err: err}
	}
}

func (m Model) createBoardCmd(name string) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		board, err := svc.CreateBoard(context.Background(), name)
		if err != nil {
			return boardsLoadedMsg{err: err}
		}
		loaded, err := svc.Board(context.Background(), board.ID)
		return boardLoadedMsg{board: loaded, resetCursor: true, err: err}
	}
}

func (m Model) renameBoardCmd(boardID int64, name string) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		if err := svc.RenameBoard(context.Background(), boardID, name); err != nil {
			return boardsLoadedMsg{err: err}
		}
		board, err := svc.Board(context.Background(), boardID)
		return boardLoadedMsg{board: board, err: err}
	}
}

func (m Model) deleteBoardCmd(boardID, switchTo int64) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		if err := svc.DeleteBoard(context.Background(), boardID); err != nil {
			return boardsLoadedMsg{err: err}
		}
		board, err := svc.Board(context.Background(), switchTo)
		return boardLoadedMsg{board: board, resetCursor: true, err: err}
	}
}

func (m Model) createColumnCmd(boardID int64, name string) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		col, err := svc.CreateColumn(ctx, boardID, name)
		if err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, &col.ID)
	}
}

func (m Model) renameColumnCmd(columnID int64, name string) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.RenameColumn(ctx, columnID, name); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, &columnID)
	}
}

func (m Model) deleteColumnCmd(columnID int64) tea.Cmd {
	svc, boardID := m.svc, m.boardID
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.DeleteColumn(ctx, columnID); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, nil)
	}
}

func (m Model) moveColumnCmd(boardID, columnID int64, toIndex int) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.MoveColumn(ctx, boardID, columnID, toIndex); err != nil {
			return boardLoadedMsg{err: err}
		}
		return loadBoardMsg(svc, boardID, nil, &columnID)
	}
}
