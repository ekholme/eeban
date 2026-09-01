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
