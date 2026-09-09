package domain

import "strings"

// FuzzyMatch reports whether every rune of pattern appears in text in order
// (a case-insensitive subsequence match). An empty pattern matches anything.
func FuzzyMatch(pattern, text string) bool {
	if pattern == "" {
		return true
	}
	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)

	pr := []rune(pattern)
	i := 0
	for _, tr := range text {
		if tr == pr[i] {
			i++
			if i == len(pr) {
				return true
			}
		}
	}
	return false
}

// CardMatches reports whether the card satisfies a fuzzy text query against
// its title or body. An empty query always matches.
func (c Card) CardMatches(query string) bool {
	if strings.TrimSpace(query) == "" {
		return true
	}
	return FuzzyMatch(query, c.Title) || FuzzyMatch(query, c.Body)
}
