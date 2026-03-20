package text

import "testing"

func TestConvertMarkdownToSlackMrkdwn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty / whitespace
		{"empty", "", ""},
		{"whitespace only", "   \n\n  ", ""},

		// Paragraphs
		{"single paragraph", "Hello world", "Hello world"},
		{"two paragraphs", "First paragraph.\n\nSecond paragraph.", "First paragraph.\n\nSecond paragraph."},

		// Inline formatting
		{"bold", "This is **important** text", "This is *important* text"},
		{"italic", "This is *emphasized* text", "This is _emphasized_ text"},
		{"code span", "Use `fmt.Println` here", "Use `fmt.Println` here"},
		{"strikethrough", "This is ~~deleted~~ text", "This is ~deleted~ text"},
		{"bold and italic", "Both **bold** and *italic*", "Both *bold* and _italic_"},

		// Links
		{"link with text", "[Google](https://google.com)", "<https://google.com|Google>"},
		{"link in paragraph", "Visit [site](https://x.com) now", "Visit <https://x.com|site> now"},

		// Headings
		{"h1", "# Title", "*Title*"},
		{"h2", "## Subtitle", "*Subtitle*"},
		{"h3 with inline", "### A **bold** heading", "*A *bold* heading*"},

		// Code blocks
		{"fenced code", "```\nfoo\nbar\n```", "```\nfoo\nbar\n```"},
		{"fenced code with lang", "```python\nprint('hi')\n```", "```\nprint('hi')\n```"},

		// Lists
		{"bullet list", "- one\n- two\n- three", "\u2022 one\n\u2022 two\n\u2022 three"},
		{"ordered list", "1. one\n2. two\n3. three", "1. one\n2. two\n3. three"},
		{"bullet list with formatting", "- **bold** item\n- `code` item", "\u2022 *bold* item\n\u2022 `code` item"},

		// Blockquotes
		{"blockquote", "> quoted text", "> quoted text"},
		{"multi-line blockquote", "> line one\n> line two", "> line one\n> line two"},

		// Mixed content
		{"paragraph then list", "Intro:\n\n- a\n- b", "Intro:\n\n\u2022 a\n\u2022 b"},
		{"paragraph then code", "Here is code:\n\n```\nx = 1\n```", "Here is code:\n\n```\nx = 1\n```"},

		// Thematic break
		{"hr", "above\n\n---\n\nbelow", "above\n\n---\n\nbelow"},

		// Complex message (typical Murphy output)
		{
			"typical agent message",
			"The trained result was negative.\n\nIf you mean the **2026-03-13 UTC Round 7 objective sweep**, the old Round 6 anchor stayed best:\n\n- large-batch BCE: **0.6381 / 0.6496**\n- equal-bin BCE: **0.6189 / 0.6146**\n\nSo the new objective did **not** improve the detector.",
			"The trained result was negative.\n\nIf you mean the *2026-03-13 UTC Round 7 objective sweep*, the old Round 6 anchor stayed best:\n\n\u2022 large-batch BCE: *0.6381 / 0.6496*\n\u2022 equal-bin BCE: *0.6189 / 0.6146*\n\nSo the new objective did *not* improve the detector.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertMarkdownToSlackMrkdwn(tt.input)
			if got != tt.expected {
				t.Errorf("ConvertMarkdownToSlackMrkdwn(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.expected)
			}
		})
	}
}
