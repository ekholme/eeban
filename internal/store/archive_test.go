package store

import (
	"context"
	"testing"
)

func TestArchiveAndUnarchiveCard(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// Card 1 lives in Backlog (column 1) alongside card 2.
	if err := db.ArchiveCard(ctx, 1); err != nil {
		t.Fatalf("ArchiveCard: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	for _, c := range board.Columns[0].Cards {
		if c.ID == 1 {
			t.Fatalf("archived card still on the board: %+v", c)
		}
	}

	archived, err := db.LoadArchived(ctx, 1)
	if err != nil {
		t.Fatalf("LoadArchived: %v", err)
	}
	if len(archived) != 1 || archived[0].ID != 1 {
		t.Fatalf("LoadArchived = %+v, want card 1", archived)
	}
	if archived[0].ArchivedAt == nil {
		t.Fatalf("archived card missing ArchivedAt")
	}

	// Restore it: back on the board, off the archive list, at the end of Backlog.
	if err := db.UnarchiveCard(ctx, 1); err != nil {
		t.Fatalf("UnarchiveCard: %v", err)
	}
	board, _ = db.LoadBoard(ctx, 1)
	last := board.Columns[0].Cards[len(board.Columns[0].Cards)-1]
	if last.ID != 1 {
		t.Fatalf("restored card not at end of column: %+v", board.Columns[0].Cards)
	}
	archived, _ = db.LoadArchived(ctx, 1)
	if len(archived) != 0 {
		t.Fatalf("archive list not empty after restore: %+v", archived)
	}
}
