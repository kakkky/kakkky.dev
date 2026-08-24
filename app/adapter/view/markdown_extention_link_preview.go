package view

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"html"
	"net/url"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type linkPreviewExtension struct{}

func NewlinkPreviewExtension() goldmark.Extender { return linkPreviewExtension{} }

func (linkPreviewExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		// GFM autolink (priority 999) より後に走らせる
		util.Prioritized(linkPreviewTransformer{}, 1000),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(linkPreviewRenderer{}, 500),
	))
}

var kindlinkPreview = ast.NewNodeKind("linkPreview")

type linkPreviewNode struct {
	ast.BaseBlock
	URL string
}

func (n *linkPreviewNode) Kind() ast.NodeKind { return kindlinkPreview }
func (n *linkPreviewNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"URL": n.URL}, nil)
}

type linkPreviewTransformer struct{}

func (linkPreviewTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	var autos []*ast.AutoLink

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		paragraph, ok := n.(*ast.Paragraph)
		if !ok || paragraph.ChildCount() != 1 {
			return ast.WalkContinue, nil
		}

		// extension.GFM 有効の場合に、生 URL は *ast.AutoLinkとなっている
		auto, ok := paragraph.FirstChild().(*ast.AutoLink)
		if !ok {
			return ast.WalkContinue, nil
		}
		autos = append(autos, auto)
		return ast.WalkSkipChildren, nil
	})

	// ast.AutoLink の 親要素 Paragraph を linkPreviewNode に置き換える
	for _, auto := range autos {
		paragraph := auto.Parent()                                                           // Paragraph
		parent := paragraph.Parent()                                                         // Document (or list item 等)
		parent.ReplaceChild(parent, paragraph, &linkPreviewNode{URL: string(auto.URL(src))}) // paragraph ごと差し替え
	}
}

type linkPreviewRenderer struct{}

func (linkPreviewRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindlinkPreview, renderlinkPreview)
}

func renderlinkPreview(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// leaf node なので入る時だけ処理
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*linkPreviewNode)

	escURL := html.EscapeString(n.URL)
	fmt.Fprintf(w,
		`<turbo-frame id="%s" src="/link-preview?url=%s" loading="lazy" class="block my-4">`+
			`<a href="%s" target="_blank" rel="noopener noreferrer"`+
			` class="block h-24 rounded-lg border border-gray-200 bg-gray-50 animate-pulse">`+
			`<span class="sr-only">%s</span>`+
			`</a>`+
			`</turbo-frame>`+"\n",
		previewFrameID(n.URL),
		url.QueryEscape(n.URL),
		escURL,
		escURL,
	)

	// linkPreviewNode は子を持たないので探索打ち切り
	return ast.WalkSkipChildren, nil
}

func previewFrameID(rawURL string) string {
	sum := sha1.Sum([]byte(rawURL))
	return "link-preview-" + hex.EncodeToString(sum[:])[:12]
}
