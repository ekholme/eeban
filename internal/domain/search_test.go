package domain

import "testing"

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		pattern, text string
		want          bool
	}{
		{"", "anything", true},
		{"abc", "abc", true},
		{"abc", "a-b-c", true},
		{"ABC", "aXbXc", true},
		{"abc", "acb", false},
		{"abcd", "abc", false},
		{"card", "Build the card form", true},
	}
	for _, tc := range cases {
		if got := FuzzyMatch(tc.pattern, tc.text); got != tc.want {
			t.Errorf("FuzzyMatch(%q, %q) = %v, want %v", tc.pattern, tc.text, got, tc.want)
		}
	}
}

func TestCardMatches(t *testing.T) {
	c := Card{Title: "Ship the release", Body: "cut a tag and push"}
	if !c.CardMatches("") {
		t.Error("empty query should match")
	}
	if !c.CardMatches("ship") {
		t.Error("title query should match")
	}
	if !c.CardMatches("tag") {
		t.Error("body query should match")
	}
	if c.CardMatches("zzz") {
		t.Error("unrelated query should not match")
	}
}
