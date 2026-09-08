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
