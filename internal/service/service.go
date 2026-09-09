// Package service is the single API the TUI uses to read and mutate board
// state. It orchestrates domain rules and the store so the TUI never touches
// SQL directly.
package service

import (
	"context"

	"github.com/ekholme/eeban/internal/domain"
	"github.com/ekholme/eeban/internal/store"
)

// DefaultBoardID is the board shown until multi-board support lands.
const DefaultBoardID = 1

// Service wires the store to the callers (currently just the TUI).
type Service struct {
	db *store.DB
}

// New returns a Service backed by db.
func New(db *store.DB) *Service {
	return &Service{db: db}
}

// DefaultBoard loads the board the TUI opens on, including columns and cards.
func (s *Service) DefaultBoard(ctx context.Context) (domain.Board, error) {
	return s.db.LoadBoard(ctx, DefaultBoardID)
}

// Board loads a specific board by id, including its columns, cards, and labels.
func (s *Service) Board(ctx context.Context, id int64) (domain.Board, error) {
	return s.db.LoadBoard(ctx, id)
}

// Boards lists every board (id and name only) for the board switcher.
func (s *Service) Boards(ctx context.Context) ([]domain.Board, error) {
	return s.db.ListBoards(ctx)
}

// CreateBoard adds a new board seeded with default columns.
func (s *Service) CreateBoard(ctx context.Context, name string) (domain.Board, error) {
	return s.db.CreateBoard(ctx, name)
}

// RenameBoard overwrites a board's name.
func (s *Service) RenameBoard(ctx context.Context, id int64, name string) error {
	return s.db.RenameBoard(ctx, id, name)
}

// DeleteBoard permanently removes a board and everything on it.
func (s *Service) DeleteBoard(ctx context.Context, id int64) error {
	return s.db.DeleteBoard(ctx, id)
}

// ArchivedCards lists a board's archived cards, most recent first.
func (s *Service) ArchivedCards(ctx context.Context, boardID int64) ([]domain.Card, error) {
	return s.db.LoadArchived(ctx, boardID)
}

// ArchiveCard removes a card from the board view without deleting it.
func (s *Service) ArchiveCard(ctx context.Context, id int64) error {
	return s.db.ArchiveCard(ctx, id)
}

// UnarchiveCard restores an archived card to the end of its column.
func (s *Service) UnarchiveCard(ctx context.Context, id int64) error {
	return s.db.UnarchiveCard(ctx, id)
}

// SetColumnWIP sets or (with a nil limit) clears a column's WIP limit.
func (s *Service) SetColumnWIP(ctx context.Context, id int64, limit *int) error {
	return s.db.SetColumnWIP(ctx, id, limit)
}

// CreateLabel defines a new coloured label on a board.
func (s *Service) CreateLabel(ctx context.Context, boardID int64, name, color string) (domain.Label, error) {
	return s.db.CreateLabel(ctx, boardID, name, color)
}

// DeleteLabel removes a label from a board and every card that carried it.
func (s *Service) DeleteLabel(ctx context.Context, id int64) error {
	return s.db.DeleteLabel(ctx, id)
}

// SetCardLabel attaches (on) or detaches (off) a label from a card.
func (s *Service) SetCardLabel(ctx context.Context, cardID, labelID int64, on bool) error {
	return s.db.SetCardLabel(ctx, cardID, labelID, on)
}

// CreateCard adds a new card with just a title to the end of columnID.
func (s *Service) CreateCard(ctx context.Context, columnID int64, title string) (domain.Card, error) {
	return s.db.CreateCard(ctx, columnID, title)
}

// UpdateCard overwrites a card's editable fields.
func (s *Service) UpdateCard(ctx context.Context, id int64, title, body string, priority int, dueDate *string) error {
	return s.db.UpdateCard(ctx, id, title, body, priority, dueDate)
}

// DeleteCard permanently removes a card.
func (s *Service) DeleteCard(ctx context.Context, id int64) error {
	return s.db.DeleteCard(ctx, id)
}

// MoveCard relocates a card to toColumnID at position toIndex, reordering in
// place when toColumnID is the card's current column.
func (s *Service) MoveCard(ctx context.Context, cardID, toColumnID int64, toIndex int) error {
	return s.db.MoveCard(ctx, cardID, toColumnID, toIndex)
}

// CreateColumn adds a new column to the end of boardID.
func (s *Service) CreateColumn(ctx context.Context, boardID int64, name string) (domain.Column, error) {
	return s.db.CreateColumn(ctx, boardID, name)
}

// RenameColumn overwrites a column's name.
func (s *Service) RenameColumn(ctx context.Context, id int64, name string) error {
	return s.db.RenameColumn(ctx, id, name)
}

// DeleteColumn permanently removes a column and its cards.
func (s *Service) DeleteColumn(ctx context.Context, id int64) error {
	return s.db.DeleteColumn(ctx, id)
}

// MoveColumn relocates a column to position toIndex within boardID.
func (s *Service) MoveColumn(ctx context.Context, boardID, columnID int64, toIndex int) error {
	return s.db.MoveColumn(ctx, boardID, columnID, toIndex)
}
