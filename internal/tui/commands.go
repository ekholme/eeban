package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/domain"
)

// boardLoadedMsg carries the result of a board reload, optionally after a
// mutation. When err is non-nil, board is meaningless and the current model
// state is left untouched. selectCardID, when set, tells the model which
// card the cursor should follow after the reload.
type boardLoadedMsg struct {
	board          domain.Board
	selectCardID   *int64
	selectColumnID *int64
	err            error
}

func (m Model) createCardCmd(columnID int64, title string) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		card, err := svc.CreateCard(ctx, columnID, title)
		if err != nil {
			return boardLoadedMsg{err: err}
		}
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, selectCardID: &card.ID, err: err}
	}
}

func (m Model) updateCardCmd(cardID int64, title, body string, priority int, dueDate *string) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.UpdateCard(ctx, cardID, title, body, priority, dueDate); err != nil {
			return boardLoadedMsg{err: err}
		}
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, selectCardID: &cardID, err: err}
	}
}

func (m Model) deleteCardCmd(cardID int64) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.DeleteCard(ctx, cardID); err != nil {
			return boardLoadedMsg{err: err}
		}
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, err: err}
	}
}

func (m Model) moveCardCmd(cardID, toColumnID int64, toIndex int) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.MoveCard(ctx, cardID, toColumnID, toIndex); err != nil {
			return boardLoadedMsg{err: err}
		}
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, selectCardID: &cardID, err: err}
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
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, selectColumnID: &col.ID, err: err}
	}
}

func (m Model) renameColumnCmd(columnID int64, name string) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.RenameColumn(ctx, columnID, name); err != nil {
			return boardLoadedMsg{err: err}
		}
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, selectColumnID: &columnID, err: err}
	}
}

func (m Model) deleteColumnCmd(columnID int64) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.DeleteColumn(ctx, columnID); err != nil {
			return boardLoadedMsg{err: err}
		}
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, err: err}
	}
}

func (m Model) moveColumnCmd(boardID, columnID int64, toIndex int) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		ctx := context.Background()
		if err := svc.MoveColumn(ctx, boardID, columnID, toIndex); err != nil {
			return boardLoadedMsg{err: err}
		}
		board, err := svc.DefaultBoard(ctx)
		return boardLoadedMsg{board: board, selectColumnID: &columnID, err: err}
	}
}
