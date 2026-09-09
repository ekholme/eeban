package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/domain"
)

func TestDueDateHighlightInView(t *testing.T) {
	overdue := "2000-01-01"
	soon := time.Now().Format("2006-01-02")
	b := domain.Board{
		ID:   1,
		Name: "B",
		Columns: []domain.Column{{
			ID: 1, Name: "Col", Position: 1000, Cards: []domain.Card{
				{ID: 1, Title: "late thing", DueDate: &overdue},
				{ID: 2, Title: "today thing", DueDate: &soon},
			},
		}},
	}
	m := send(New(nil, b), tea.WindowSizeMsg{Width: 100, Height: 40})
	out := m.View()
	if !strings.Contains(out, "overdue") {
		t.Fatalf("overdue card not flagged:\n%s", out)
	}
	if !strings.Contains(out, "due soon") {
		t.Fatalf("due-soon card not flagged:\n%s", out)
	}
}
