package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMarkdownArticle(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		wantHTML    []string
		wantOutline []*OutlineNode
	}{
		{
			name:        "no headings: plain paragraph",
			src:         "hello world\n",
			wantHTML:    []string{"<p>hello world</p>"},
			wantOutline: nil,
		},
		{
			name:     "single h1",
			src:      "# Hello\n",
			wantHTML: []string{`id="hello"`, "Hello"},
			wantOutline: []*OutlineNode{
				{level: 1, Text: "Hello", Anchor: "hello"},
			},
		},
		{
			name: "nested h1 > h2 > h3",
			src: "# A\n" +
				"## B\n" +
				"### C\n",
			wantOutline: []*OutlineNode{
				{
					level: 1, Text: "A", Anchor: "a",
					Children: []*OutlineNode{
						{
							level: 2, Text: "B", Anchor: "b",
							Children: []*OutlineNode{
								{level: 3, Text: "C", Anchor: "c"},
							},
						},
					},
				},
			},
		},
		{
			name: "h4 is excluded from outline but still rendered",
			src: "# A\n" +
				"#### D\n" +
				"## B\n",
			wantHTML: []string{"<h4", "D"},
			wantOutline: []*OutlineNode{
				{
					level: 1, Text: "A", Anchor: "a",
					Children: []*OutlineNode{
						{level: 2, Text: "B", Anchor: "b"},
					},
				},
			},
		},
		{
			name: "multiple h1 siblings",
			src: "# A\n" +
				"# B\n",
			wantOutline: []*OutlineNode{
				{level: 1, Text: "A", Anchor: "a"},
				{level: 1, Text: "B", Anchor: "b"},
			},
		},
		{
			name: "h2 without preceding h1 becomes a root",
			src:  "## Only\n",
			wantOutline: []*OutlineNode{
				{level: 2, Text: "Only", Anchor: "only"},
			},
		},
		{
			name: "sibling h2 after popping deeper stack",
			src: "# A\n" +
				"## B\n" +
				"### C\n" +
				"## D\n",
			wantOutline: []*OutlineNode{
				{
					level: 1, Text: "A", Anchor: "a",
					Children: []*OutlineNode{
						{
							level: 2, Text: "B", Anchor: "b",
							Children: []*OutlineNode{
								{level: 3, Text: "C", Anchor: "c"},
							},
						},
						{level: 2, Text: "D", Anchor: "d"},
					},
				},
			},
		},
		{
			name: "japanese heading keeps unicode slug",
			src:  "# こんにちは\n",
			wantOutline: []*OutlineNode{
				{level: 1, Text: "こんにちは", Anchor: "こんにちは"},
			},
		},
		{
			name: "duplicate headings get suffixed anchors",
			src: "# hello\n" +
				"# hello\n" +
				"# hello\n",
			wantOutline: []*OutlineNode{
				{level: 1, Text: "hello", Anchor: "hello"},
				{level: 1, Text: "hello", Anchor: "hello-1"},
				{level: 1, Text: "hello", Anchor: "hello-2"},
			},
		},
		{
			name:        "fenced code block renders with chroma classes",
			src:         "```go\nfmt.Println()\n```\n",
			wantHTML:    []string{"chroma"},
			wantOutline: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, outline := ParseMarkdownArticle(tt.src)
			for _, s := range tt.wantHTML {
				assert.Contains(t, html, s)
			}
			assert.Equal(t, tt.wantOutline, outline)
		})
	}
}
