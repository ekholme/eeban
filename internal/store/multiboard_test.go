package store

import (
	"context"
	"testing"
)

func TestBoardLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	boards, err := db.ListBoards(ctx)
	if err != nil {
		t.Fatalf("ListBoards: %v", err)
	}
	if len(boards) != 1 || boards[0].Name != "My Board" {
		t.Fatalf("ListBoards = %+v", boards)
	}

	created, err := db.CreateBoard(ctx, "Side Project")
	if err != nil {
		t.Fatalf("CreateBoard: %v", err)
	}
	if created.ID == 0 || created.Name != "Side Project" {
		t.Fatalf("CreateBoard returned %+v", created)
	}

	// A new board comes seeded with the three default columns and no cards.
	loaded, err := db.LoadBoard(ctx, created.ID)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(loaded.Columns) != 3 {
		t.Fatalf("new board columns = %d, want 3", len(loaded.Columns))
	}
	for _, c := range loaded.Columns {
		if len(c.Cards) != 0 {
			t.Fatalf("new board column %q has cards: %+v", c.Name, c.Cards)
		}
	}

	if err := db.RenameBoard(ctx, created.ID, "Renamed"); err != nil {
		t.Fatalf("RenameBoard: %v", err)
	}
	boards, _ = db.ListBoards(ctx)
	if len(boards) != 2 || boards[1].Name != "Renamed" {
		t.Fatalf("ListBoards after rename = %+v", boards)
	}

	if err := db.DeleteBoard(ctx, created.ID); err != nil {
		t.Fatalf("DeleteBoard: %v", err)
	}
	boards, _ = db.ListBoards(ctx)
	if len(boards) != 1 {
		t.Fatalf("ListBoards after delete = %+v", boards)
	}
}
