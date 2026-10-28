package domain

import (
	"testing"
	"time"
)

func TestDueStatusFor(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	ptr := func(s string) *string { return &s }

	cases := []struct {
		name string
		due  *string
		want DueStatus
	}{
		{"nil", nil, DueNone},
		{"blank", ptr(""), DueNone},
		{"garbage", ptr("not-a-date"), DueNone},
		{"yesterday", ptr("2026-09-07"), DueOverdue},
		{"today", ptr("2026-09-08"), DueSoon},
		{"tomorrow", ptr("2026-09-09"), DueSoon},
		{"in two days", ptr("2026-09-10"), DueSoon},
		{"in three days", ptr("2026-09-11"), DueFuture},
		{"next month", ptr("2026-10-08"), DueFuture},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DueStatusFor(tc.due, now); got != tc.want {
				t.Errorf("DueStatusFor(%v) = %v, want %v", tc.due, got, tc.want)
			}
		})
	}
}

func TestUpcomingDue(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	ptr := func(s string) *string { return &s }
	card := func(id int64, due string) DueCard {
		return DueCard{Card: Card{ID: id, DueDate: &due}}
	}

	got := UpcomingDue([]DueCard{
		card(1, "2026-09-13"), // in five days: kept
		card(2, "2026-09-14"), // in six days: dropped
		card(3, "2026-09-08"), // today
		card(4, "2026-08-01"), // overdue
		card(5, "garbage"),    // unparseable: dropped
		card(6, "2026-09-08"), // today, after card 3
		{Card: Card{ID: 7, DueDate: ptr("2026-09-08")}, ColumnName: " done "}, // done column: dropped
	}, now)

	want := []struct {
		id   int64
		days int
	}{{4, -38}, {3, 0}, {6, 0}, {1, 5}}
	if len(got) != len(want) {
		t.Fatalf("UpcomingDue = %+v, want ids %v", got, want)
	}
	for i, w := range want {
		if got[i].ID != w.id || got[i].DaysUntil != w.days {
			t.Errorf("got[%d] = id %d days %d, want id %d days %d",
				i, got[i].ID, got[i].DaysUntil, w.id, w.days)
		}
	}
}
