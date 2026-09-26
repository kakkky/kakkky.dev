package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMarkdownArticle(t *testing.T) {
	t.Parallel()
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
		{
			name: "directive: message info renders flex layout with info icon + inner markdown",
			src: ":::message info\n" +
				"**bold** text\n" +
				":::\n",
			wantHTML: []string{
				`<div class="my-4 rounded-lg p-4 bg-blue-100 flex gap-3 items-start">`,
				`<span class="text-blue-600">`,
				`viewBox="0 0 24 24"`, // Material Icons Filled
				`<div class="min-w-0 flex-1 [&>*:first-child]:mt-0 [&>*:last-child]:mb-0">`,
				"<p><strong>bold</strong> text</p>",
				"</div>",
			},
			wantOutline: nil,
		},
		{
			name: "directive: message warn renders amber variant with warn icon",
			src: ":::message warn\n" +
				"careful\n" +
				":::\n",
			wantHTML: []string{
				`<div class="my-4 rounded-lg p-4 bg-amber-100 flex gap-3 items-start">`,
				`<span class="text-amber-600">`,
				`viewBox="0 0 24 24"`,
				"<p>careful</p>",
				"</div>",
			},
			wantOutline: nil,
		},
		{
			name: "directive: unknown message kind falls back to info variant",
			src: ":::message hoge\n" +
				"body\n" +
				":::\n",
			wantHTML: []string{
				`<div class="my-4 rounded-lg p-4 bg-blue-100 flex gap-3 items-start">`,
				`<span class="text-blue-600">`,
				"<p>body</p>",
			},
			wantOutline: nil,
		},
		{
			name: "directive: toggle renders details/summary with title and toggles bg on open",
			src: ":::toggle 補足: 詳細\n" +
				"body text\n" +
				":::\n",
			wantHTML: []string{
				`<details class="group my-4 rounded-md border border-gray-200 bg-gray-50 open:bg-white">`,
				`<summary class="cursor-pointer px-3 py-2 text-sm text-gray-600 [list-style:revert] group-open:border-b group-open:border-gray-200">補足: 詳細</summary>`,
				`<div class="px-4 py-3 bg-gray-50 rounded-b-md [&>*:first-child]:mt-0 [&>*:last-child]:mb-0">`,
				"<p>body text</p>",
				"</details>",
			},
			wantOutline: nil,
		},
		{
			name: "directive: toggle title is html-escaped",
			src: ":::toggle <script>\n" +
				"x\n" +
				":::\n",
			wantHTML:    []string{"&lt;script&gt;</summary>"},
			wantOutline: nil,
		},
		{
			name: "directive: unknown name falls back to raw text",
			src: ":::hoge\n" +
				"body\n" +
				":::\n",
			wantHTML:    []string{":::hoge"},
			wantOutline: nil,
		},
		{
			name: "directive: heading inside directive is picked up by outline",
			src: ":::message info\n" +
				"## sub heading\n" +
				":::\n",
			wantHTML: []string{"bg-blue-100", "<h2"},
			wantOutline: []*OutlineNode{
				{level: 2, Text: "sub heading", Anchor: "sub-heading"},
			},
		},
		{
			name: "directive: unclosed at EOF is implicitly closed",
			src: ":::message info\n" +
				"body without closing fence\n",
			wantHTML: []string{
				"bg-blue-100",
				"body without closing fence",
				"</div>",
			},
			wantOutline: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			html, outline := ParseMarkdownArticle(tt.src)
			for _, s := range tt.wantHTML {
				assert.Contains(t, html, s)
			}
			assert.Equal(t, tt.wantOutline, outline)
		})
	}
}

func TestParseMarkdownArticle_XSS(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		src         string
		contains    []string
		notContains []string
	}{
		{
			name:        "raw <script> block is stripped",
			src:         "<script>alert(1)</script>\n",
			notContains: []string{"<script", "alert(1)"},
		},
		{
			name:        "inline <script> is stripped, surrounding text remains",
			src:         "hello <script>alert(1)</script> world\n",
			contains:    []string{"<p>hello ", " world</p>"},
			notContains: []string{"<script"},
		},
		{
			name:        "raw <img onerror> is stripped",
			src:         "<img src=x onerror=alert(1)>\n",
			notContains: []string{"<img", "onerror"},
		},
		{
			name:        "raw <a href=data:...> is stripped",
			src:         `<a href="data:text/html,<script>alert(1)</script>">click</a>` + "\n",
			contains:    []string{"click"},
			notContains: []string{"<a href", "data:text/html", "<script"},
		},
		{
			name:        "markdown link with javascript: href is neutralized to empty href",
			src:         "[link](javascript:alert(1))\n",
			contains:    []string{`href=""`, ">link</a>"},
			notContains: []string{"javascript:"},
		},
		{
			name:        "markdown link with data:text/html href is neutralized to empty href",
			src:         "[link](data:text/html,<script>alert(1)</script>)\n",
			contains:    []string{`href=""`, ">link</a>"},
			notContains: []string{"data:text/html", "<script"},
		},
		{
			name:        "markdown link with vbscript: href is neutralized to empty href",
			src:         "[link](vbscript:msgbox(1))\n",
			contains:    []string{`href=""`},
			notContains: []string{"vbscript:"},
		},
		{
			name:        "autolink with javascript: is not promoted to linkPreview and href is dropped",
			src:         "<javascript:alert(1)>\n",
			notContains: []string{`href="javascript:`, "turbo-frame"},
		},
		{
			name:        "autolink with data:text/html is not promoted to linkPreview",
			src:         "<data:text/html,alert(1)>\n",
			notContains: []string{`href="data:`, "turbo-frame"},
		},
		{
			name:        "image with javascript: src is neutralized",
			src:         "![x](javascript:alert(1))\n",
			notContains: []string{`src="javascript:`, "javascript:alert"},
		},
		{
			name: "directive toggle title escapes attribute-breakout attempt",
			src: `:::toggle " onclick="alert(1)` + "\n" +
				"body\n" +
				":::\n",
			contains:    []string{"&#34;"}, // 記号 " が属性値内で無害化されている
			notContains: []string{`" onclick="`, `onclick="alert`},
		},
		{
			name: "directive toggle title escapes tag-breakout attempt",
			src: ":::toggle </summary><script>alert(1)</script>\n" +
				"body\n" +
				":::\n",
			contains:    []string{"&lt;/summary&gt;", "&lt;script&gt;"},
			notContains: []string{"</summary><script", "<script>alert"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			html, _ := ParseMarkdownArticle(tt.src)
			for _, s := range tt.contains {
				assert.Contains(t, html, s)
			}
			for _, s := range tt.notContains {
				assert.NotContains(t, html, s)
			}
		})
	}
}
