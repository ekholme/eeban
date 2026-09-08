package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ekholme/eeban/internal/domain"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "eeban.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateCardAppendsAtEnd(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	card, err := db.CreateCard(ctx, 1, "New task")
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	if card.Title != "New task" || card.ColumnID != 1 {
		t.Errorf("CreateCard returned %+v", card)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	cards := board.Columns[0].Cards
	last := cards[len(cards)-1]
	if last.ID != card.ID || last.Title != "New task" {
		t.Errorf("new card not last in column: %+v", cards)
	}
}

func TestUpdateCard(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	due := "2026-12-25"
	if err := db.UpdateCard(ctx, 1, "Renamed", "new body", 2, &due); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	card := board.Columns[0].Cards[0]
	if card.Title != "Renamed" || card.Body != "new body" || card.Priority != 2 {
		t.Errorf("card after update = %+v", card)
	}
	if card.DueDate == nil || *card.DueDate != due {
		t.Errorf("card.DueDate = %v, want %q", card.DueDate, due)
	}
}

func TestDeleteCard(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if err := db.DeleteCard(ctx, 1); err != nil {
		t.Fatalf("DeleteCard: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	for _, c := range board.Columns[0].Cards {
		if c.ID == 1 {
			t.Fatalf("deleted card still present: %+v", c)
		}
	}
}

func TestMoveCardWithinColumnReorders(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// Backlog seed order: card 1 "Welcome to eeban", card 2 "Press ? for help".
	if err := db.MoveCard(ctx, 2, 1, 0); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	cards := board.Columns[0].Cards
	if len(cards) != 2 || cards[0].ID != 2 || cards[1].ID != 1 {
		t.Errorf("cards after reorder = %+v, want [2, 1]", cards)
	}
}

func TestMoveCardAcrossColumns(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// Move card 1 (Backlog) to the end of column 2 (In Progress).
	if err := db.MoveCard(ctx, 1, 2, 1); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(board.Columns[0].Cards) != 1 || board.Columns[0].Cards[0].ID != 2 {
		t.Errorf("Backlog after move = %+v", board.Columns[0].Cards)
	}
	inProgress := board.Columns[1].Cards
	if len(inProgress) != 2 || inProgress[1].ID != 1 || inProgress[1].ColumnID != 2 {
		t.Errorf("In Progress after move = %+v", inProgress)
	}
}

func TestMoveCardRenormalizesCollapsedGap(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// Force a 1-unit gap in column 1 between the two seeded cards.
	if _, err := db.ExecContext(ctx, `UPDATE cards SET position = 1000 WHERE id = 1`); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE cards SET position = 1001 WHERE id = 2`); err != nil {
		t.Fatalf("setup: %v", err)
	}

	card, err := db.CreateCard(ctx, 1, "Squeeze me")
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	if err := db.MoveCard(ctx, card.ID, 1, 1); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	cards := board.Columns[0].Cards
	if len(cards) != 3 || cards[1].ID != card.ID {
		t.Fatalf("cards after squeeze move = %+v", cards)
	}
	for i := 1; i < len(cards); i++ {
		if cards[i].Position-cards[i-1].Position < domain.PositionGap-1 {
			t.Errorf("positions not renormalized: %+v", cards)
		}
	}
}
