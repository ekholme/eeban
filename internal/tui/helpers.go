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
