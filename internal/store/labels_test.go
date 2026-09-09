package store

import (
	"context"
	"testing"
)

func TestLabelLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	label, err := db.CreateLabel(ctx, 1, "bug", "1")
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	if label.ID == 0 || label.Name != "bug" || label.Color != "1" {
		t.Fatalf("CreateLabel returned %+v", label)
	}

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(board.Labels) != 1 || board.Labels[0].Name != "bug" {
		t.Fatalf("board.Labels = %+v", board.Labels)
	}

	// Attach to card 1, confirm it comes back on the card.
	if err := db.SetCardLabel(ctx, 1, label.ID, true); err != nil {
		t.Fatalf("SetCardLabel on: %v", err)
	}
	// A second attach is a no-op, not an error.
	if err := db.SetCardLabel(ctx, 1, label.ID, true); err != nil {
		t.Fatalf("SetCardLabel on (repeat): %v", err)
	}

	board, _ = db.LoadBoard(ctx, 1)
	if got := board.Columns[0].Cards[0].Labels; len(got) != 1 || got[0].ID != label.ID {
		t.Fatalf("card labels = %+v", got)
	}

	// Detach.
	if err := db.SetCardLabel(ctx, 1, label.ID, false); err != nil {
		t.Fatalf("SetCardLabel off: %v", err)
	}
	board, _ = db.LoadBoard(ctx, 1)
	if got := board.Columns[0].Cards[0].Labels; len(got) != 0 {
		t.Fatalf("card labels after detach = %+v", got)
	}

	// Deleting a label removes it from the board and cascades card_labels.
	if err := db.SetCardLabel(ctx, 2, label.ID, true); err != nil {
		t.Fatalf("SetCardLabel: %v", err)
	}
	if err := db.DeleteLabel(ctx, label.ID); err != nil {
		t.Fatalf("DeleteLabel: %v", err)
	}
	board, _ = db.LoadBoard(ctx, 1)
	if len(board.Labels) != 0 {
		t.Fatalf("board.Labels after delete = %+v", board.Labels)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM card_labels`).Scan(&n); err != nil {
		t.Fatalf("count card_labels: %v", err)
	}
	if n != 0 {
		t.Fatalf("card_labels rows after label delete = %d, want 0", n)
	}
}
