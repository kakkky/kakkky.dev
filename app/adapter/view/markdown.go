package view

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var mdParser = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.CJK,
		extension.Footnote,
		NewDirectiveExtension(),
		NewlinkPreviewExtension(),
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
		),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

const maxOutlineLevel = 3

type OutlineNode struct {
	level    int
	Text     string
	Anchor   string
	Children []*OutlineNode
}

func ParseMarkdownArticle(src string) (html string, outline []*OutlineNode) {
	source := []byte(src)
	pCtx := parser.NewContext(
		parser.WithIDs(&unicodeIDs{values: map[string]struct{}{}}),
	)
	doc := mdParser.Parser().Parse(text.NewReader(source), parser.WithContext(pCtx))

	var buf bytes.Buffer
	_ = mdParser.Renderer().Render(&buf, source, doc)
	html = buf.String()

	outline = extractOutline(doc, source)

	return html, outline
}

func extractOutline(doc ast.Node, src []byte) []*OutlineNode {
	var roots []*OutlineNode
	var stack []*OutlineNode

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		// ノード要素が探索済みの場合は処理をskip
		if !entering {
			return ast.WalkContinue, nil
		}

		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		if heading.Level > maxOutlineLevel {
			return ast.WalkSkipChildren, nil
		}

		node := &OutlineNode{
			level:  heading.Level,
			Text:   headingText(heading, src),
			Anchor: headingAnchor(heading),
		}

		// stackには探索中ノード要素の祖先のノードのみを残しておく
		for len(stack) > 0 && stack[len(stack)-1].level >= heading.Level {
			stack = stack[:len(stack)-1]
		}

		switch {
		case len(stack) == 0:
			// stackが空の場合は親候補がいないということなのでトップレベル
			roots = append(roots, node)
		case len(stack) > 0:
			// stackのLastの要素を親として探索中ノード要素を子に追加
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, node)
		}

		stack = append(stack, node)
		return ast.WalkContinue, nil
	})

	return roots
}

func headingText(heading *ast.Heading, src []byte) string {
	var buf bytes.Buffer
	for c := heading.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(src))
		}
	}
	return buf.String()
}

func headingAnchor(heading *ast.Heading) string {
	id, ok := heading.AttributeString("id")
	if !ok {
		return ""
	}

	// goldmark は id属性 を []byte で扱っているため、確実に []byteで返る。
	// ref: https://github.com/yuin/goldmark/blob/master/parser/atx_heading.go#L168
	switch v := id.(type) {
	case []byte:
		return string(v)
	}
	return ""
}

// unicodeIDs は goldmark の parser.IDs interface 実装。
// 標準実装 (parser.NewContext のデフォルト) は非 ASCII を捨てるため、
// 日本語見出しが全て空文字になり "heading-1", "heading-2", ... というフォールバック id が付与されてしまう。
// これを避けるため Unicode letter/digit を残すカスタム実装を用意している
type unicodeIDs struct{ values map[string]struct{} }

func (s *unicodeIDs) Generate(value []byte, kind ast.NodeKind) []byte {

	// anchorにつけるためのslug生成
	var buf bytes.Buffer
	for _, r := range strings.ToLower(string(value)) {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			buf.WriteRune(r)
		case unicode.IsSpace(r), r == '-', r == '_':
			buf.WriteRune('-')
		}
	}
	result := buf.Bytes()

	// fallback
	if len(result) == 0 {
		result = []byte("heading")
	}

	// 重複の回避
	// slugに重複があった場合は index　をつけて　ユニークにする
	base := string(result)
	if _, exists := s.values[base]; !exists {
		s.values[base] = struct{}{}
		return result
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, exists := s.values[candidate]; !exists {
			s.values[candidate] = struct{}{}
			return []byte(candidate)
		}
	}
}

func (s *unicodeIDs) Put(value []byte) {
	s.values[string(value)] = struct{}{}
}
