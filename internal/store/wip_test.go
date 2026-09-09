package store

import (
	"context"
	"testing"
)

func TestSetColumnWIP(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	limit := 3
	if err := db.SetColumnWIP(ctx, 1, &limit); err != nil {
		t.Fatalf("SetColumnWIP: %v", err)
	}
	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if board.Columns[0].WIPLimit == nil || *board.Columns[0].WIPLimit != 3 {
		t.Fatalf("WIPLimit = %v, want 3", board.Columns[0].WIPLimit)
	}

	if err := db.SetColumnWIP(ctx, 1, nil); err != nil {
		t.Fatalf("SetColumnWIP clear: %v", err)
	}
	board, _ = db.LoadBoard(ctx, 1)
	if board.Columns[0].WIPLimit != nil {
		t.Fatalf("WIPLimit = %v, want nil after clear", board.Columns[0].WIPLimit)
	}
}
