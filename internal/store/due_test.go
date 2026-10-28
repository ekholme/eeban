package store

import (
	"context"
	"testing"
)

func TestLoadDueCardsSpansBoards(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	ptr := func(s string) *string { return &s }

	other, err := db.CreateBoard(ctx, "Side")
	if err != nil {
		t.Fatalf("CreateBoard: %v", err)
	}
	otherBoard, _ := db.LoadBoard(ctx, other.ID)
	sideCard, err := db.CreateCard(ctx, otherBoard.Columns[1].ID, "side task")
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	if err := db.UpdateCard(ctx, sideCard.ID, "side task", "", 0, ptr("2026-01-01")); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	if err := db.UpdateCard(ctx, 1, "Welcome to eeban", "", 0, ptr("2026-02-01")); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	// Archived and blank-dated cards are excluded.
	if err := db.UpdateCard(ctx, 2, "Press ? for help", "", 0, ptr("2026-01-15")); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	if err := db.ArchiveCard(ctx, 2); err != nil {
		t.Fatalf("ArchiveCard: %v", err)
	}
	if err := db.UpdateCard(ctx, 3, "Build the card form", "", 0, ptr("")); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}

	label, err := db.CreateLabel(ctx, other.ID, "urgent", "1")
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	if err := db.SetCardLabel(ctx, sideCard.ID, label.ID, true); err != nil {
		t.Fatalf("SetCardLabel: %v", err)
	}

	cards, err := db.LoadDueCards(ctx)
	if err != nil {
		t.Fatalf("LoadDueCards: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("LoadDueCards = %+v, want 2 cards", cards)
	}
	first, second := cards[0], cards[1]
	if first.ID != sideCard.ID || first.BoardName != "Side" || first.ColumnName != "In Progress" {
		t.Errorf("first = %+v, want side task on Side/In Progress", first)
	}
	if len(first.Labels) != 1 || first.Labels[0].Name != "urgent" {
		t.Errorf("first labels = %+v", first.Labels)
	}
	if second.ID != 1 || second.BoardID != 1 || second.BoardName != "My Board" || second.ColumnName != "Backlog" {
		t.Errorf("second = %+v, want card 1 on My Board/Backlog", second)
	}
}
