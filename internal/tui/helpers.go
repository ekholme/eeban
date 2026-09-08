package tui

import "github.com/ekholme/eeban/internal/domain"

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
