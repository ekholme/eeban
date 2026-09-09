package domain

import "time"

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

// DueStatusFor classifies dueDate (a "YYYY-MM-DD" string, possibly nil or
// blank) against now. An unparseable date is treated as unset.
func DueStatusFor(dueDate *string, now time.Time) DueStatus {
	if dueDate == nil || *dueDate == "" {
		return DueNone
	}
	due, err := time.ParseInLocation(dateLayout, *dueDate, now.Location())
	if err != nil {
		return DueNone
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())

	days := int(dueDay.Sub(today).Hours() / 24)
	switch {
	case days < 0:
		return DueOverdue
	case days <= SoonDays:
		return DueSoon
	default:
		return DueFuture
	}
}
