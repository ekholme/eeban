package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenRunsMigrationsAndSeeds(t *testing.T) {
	ctx := context.Background()

	db, err := Open(ctx, filepath.Join(t.TempDir(), "eeban.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	board, err := db.LoadBoard(ctx, 1)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}

	if board.Name != "My Board" {
		t.Errorf("board name = %q, want %q", board.Name, "My Board")
	}
	if len(board.Columns) != 3 {
		t.Fatalf("got %d columns, want 3", len(board.Columns))
	}
	if got := board.Columns[0].Name; got != "Backlog" {
		t.Errorf("first column = %q, want Backlog", got)
	}
	if got := len(board.Columns[0].Cards); got != 2 {
		t.Errorf("Backlog has %d cards, want 2", got)
	}

	// Columns and cards must come back ordered by position.
	for i := 1; i < len(board.Columns); i++ {
		if board.Columns[i-1].Position > board.Columns[i].Position {
			t.Errorf("columns not ordered by position: %+v", board.Columns)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()

	db, err := Open(ctx, filepath.Join(t.TempDir(), "eeban.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.migrate(ctx); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM boards`).Scan(&n); err != nil {
		t.Fatalf("count boards: %v", err)
	}
	if n != 1 {
		t.Errorf("boards count = %d, want 1 (seed must not run twice)", n)
	}
}
