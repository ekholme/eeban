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
	board        domain.Board
	selectCardID *int64
	err          error
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
