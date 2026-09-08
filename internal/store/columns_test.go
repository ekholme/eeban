package store

import (
	"context"
	"testing"
)

func TestCreateColumnAppendsAtEnd(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	col, err := db.CreateColumn(ctx, 1, "Review")
	if err != nil {
		t.Fatalf("CreateColumn: %v", err)
	}
	if col.Name != "Review" || col.BoardID != 1 {
		t.Errorf("CreateColumn returned %+v", col)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	cols := board.Columns
	last := cols[len(cols)-1]
	if last.ID != col.ID || last.Name != "Review" {
		t.Errorf("new column not last on board: %+v", cols)
	}
}

func TestRenameColumn(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if err := db.RenameColumn(ctx, 1, "Todo"); err != nil {
		t.Fatalf("RenameColumn: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if board.Columns[0].Name != "Todo" {
		t.Errorf("column name = %q, want Todo", board.Columns[0].Name)
	}
}

func TestDeleteColumnCascadesCards(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if err := db.DeleteColumn(ctx, 1); err != nil {
		t.Fatalf("DeleteColumn: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	for _, c := range board.Columns {
		if c.ID == 1 {
			t.Fatalf("deleted column still present: %+v", c)
		}
	}
	for _, c := range board.Columns {
		for _, card := range c.Cards {
			if card.ColumnID == 1 {
				t.Fatalf("card from deleted column still present: %+v", card)
			}
		}
	}
}

func TestMoveColumnReorders(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// Seed order: 1 Backlog, 2 In Progress, 3 Done.
	if err := db.MoveColumn(ctx, 1, 3, 0); err != nil {
		t.Fatalf("MoveColumn: %v", err)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(board.Columns) != 3 || board.Columns[0].ID != 3 {
		ids := make([]int64, len(board.Columns))
		for i, c := range board.Columns {
			ids[i] = c.ID
		}
		t.Fatalf("columns after move = %v, want [3, 1, 2]", ids)
	}
}
