package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpcomingViewListsAndJumpsAcrossBoards(t *testing.T) {
	m := loadModel(t)
	ctx := context.Background()
	day := func(offset int) *string {
		s := time.Now().AddDate(0, 0, offset).Format("2006-01-02")
		return &s
	}

	// Card 1 overdue on the default board; a card due in 3 days on a second
	// board; card 3 due well outside the window.
	if err := m.svc.UpdateCard(ctx, 1, "Welcome to eeban", "", 0, day(-2)); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	if err := m.svc.UpdateCard(ctx, 3, "Build the card form", "", 0, day(30)); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	side, err := m.svc.CreateBoard(ctx, "Side")
	if err != nil {
		t.Fatalf("CreateBoard: %v", err)
	}
	sideBoard, _ := m.svc.Board(ctx, side.ID)
	sideCard, err := m.svc.CreateCard(ctx, sideBoard.Columns[1].ID, "side task")
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	if err := m.svc.UpdateCard(ctx, sideCard.ID, "side task", "", 0, day(3)); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}

	// Card 4 sits in the default board's Done column: hidden even though overdue.
	if err := m.svc.UpdateCard(ctx, 4, "Scaffold the repo", "", 0, day(-1)); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}

	m = runMutation(m, press("U"))
	if !m.showUpcoming {
		t.Fatalf("'U' did not open the upcoming view")
	}
	if len(m.upcoming) != 2 || m.upcoming[0].ID != 1 || m.upcoming[1].ID != sideCard.ID {
		t.Fatalf("upcoming = %+v", m.upcoming)
	}
	out := m.View()
	for _, want := range []string{"Overdue", "Next 5 days", "Welcome to eeban", "2d overdue", "side task", "Side › In Progress", "in 3d"} {
		if !strings.Contains(out, want) {
			t.Errorf("upcoming view missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Build the card form") {
		t.Errorf("card due in 30 days should not be listed:\n%s", out)
	}
	if strings.Contains(out, "Scaffold the repo") {
		t.Errorf("card in the Done column should not be listed:\n%s", out)
	}

	// Jump to the card on the other board.
	m = send(m, press("j"))
	m = runMutation(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.showUpcoming {
		t.Fatalf("enter should close the upcoming view")
	}
	if m.boardID != side.ID {
		t.Fatalf("boardID = %d, want %d", m.boardID, side.ID)
	}
	if c, ok := m.selectedCard(); !ok || c.ID != sideCard.ID {
		t.Fatalf("selected card = %+v, want side task", c)
	}
}

func TestUpcomingViewEmpty(t *testing.T) {
	m := runMutation(loadModel(t), press("U"))
	if !strings.Contains(m.View(), "Nothing overdue") {
		t.Fatalf("empty upcoming view missing message:\n%s", m.View())
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.showUpcoming {
		t.Fatalf("esc did not close the upcoming view")
	}
}
