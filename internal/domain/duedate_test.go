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
