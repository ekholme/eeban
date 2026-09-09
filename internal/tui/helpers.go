package tui

import (
	"strconv"

	"github.com/ekholme/eeban/internal/domain"
)

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// truncateText shortens s to at most n runes, appending an ellipsis when it
// had to cut.
func truncateText(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

// priorityLabel maps a domain priority constant to its display string.
func priorityLabel(p int) string {
	switch p {
	case domain.PriorityLow:
		return "Low"
	case domain.PriorityMedium:
		return "Medium"
	case domain.PriorityHigh:
		return "High"
	default:
		return "None"
	}
}

// parsePositiveInt parses s as a non-negative base-10 integer.
func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, strconv.ErrRange
	}
	return n, nil
}
