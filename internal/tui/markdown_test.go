package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"list bullet", "- one\n- two", "• one"},
		{"link expansion", "[docs](https://example.com)", "https://example.com"},
		{"plain text passes through", "just some text", "just some text"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := renderMarkdown(tc.body, 40)
			if !strings.Contains(out, tc.want) {
				t.Errorf("renderMarkdown(%q) = %q, want it to contain %q", tc.body, out, tc.want)
			}
		})
	}
}

func TestRenderMarkdownZeroWidth(t *testing.T) {
	// A non-positive width shouldn't panic or drop content.
	out := renderMarkdown("hello", 0)
	if !strings.Contains(out, "hello") {
		t.Errorf("renderMarkdown with width 0 = %q, want it to contain %q", out, "hello")
	}
}
