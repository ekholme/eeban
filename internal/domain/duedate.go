package domain

import (
	"math"
	"sort"
	"strings"
	"time"
)

// DueStatus classifies a card's due date relative to the current day.
type DueStatus int

const (
	DueNone    DueStatus = iota // no due date set
	DueFuture                   // comfortably in the future
	DueSoon                     // today or within SoonDays days
	DueOverdue                  // the due day has already passed
)

// SoonDays is how many days ahead (inclusive of today) still counts as "due
// soon" rather than merely "future".
const SoonDays = 2

// dateLayout is the storage format for due dates.
const dateLayout = "2006-01-02"

// UpcomingDays is how many days past today the cross-board due view reaches.
const UpcomingDays = 5

// DaysUntilDue returns the number of whole days from now's day to dueDate (a
// "YYYY-MM-DD" string): 0 for today, negative when overdue. ok is false when
// dueDate is nil, blank, or unparseable.
func DaysUntilDue(dueDate *string, now time.Time) (days int, ok bool) {
	if dueDate == nil || *dueDate == "" {
		return 0, false
	}
	due, err := time.ParseInLocation(dateLayout, *dueDate, now.Location())
	if err != nil {
		return 0, false
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())

	// Round rather than truncate so a DST transition's 23- or 25-hour day
	// still counts as one day.
	return int(math.Round(dueDay.Sub(today).Hours() / 24)), true
}

// DueStatusFor classifies dueDate (a "YYYY-MM-DD" string, possibly nil or
// blank) against now. An unparseable date is treated as unset.
func DueStatusFor(dueDate *string, now time.Time) DueStatus {
	days, ok := DaysUntilDue(dueDate, now)
	switch {
	case !ok:
		return DueNone
	case days < 0:
		return DueOverdue
	case days <= SoonDays:
		return DueSoon
	default:
		return DueFuture
	}
}

// DueCard is a card annotated with the board and column it lives on, for
// views that list cards from every board at once.
type DueCard struct {
	Card
	BoardID    int64
	BoardName  string
	ColumnName string
	DaysUntil  int // whole days until due; negative when overdue
}

// IsDoneColumn reports whether a column holds finished work, judged by its
// name ("Done", case-insensitive). There is no explicit done flag on columns.
func IsDoneColumn(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "done")
}

// UpcomingDue keeps the cards that are overdue, due today, or due within
// UpcomingDays days of now, skipping cards in a done column. It sets each
// kept card's DaysUntil and orders them soonest first; input order is
// preserved among cards due the same day.
func UpcomingDue(cards []DueCard, now time.Time) []DueCard {
	out := make([]DueCard, 0, len(cards))
	for _, c := range cards {
		if IsDoneColumn(c.ColumnName) {
			continue
		}
		days, ok := DaysUntilDue(c.DueDate, now)
		if !ok || days > UpcomingDays {
			continue
		}
		c.DaysUntil = days
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].DaysUntil < out[j].DaysUntil })
	return out
}
