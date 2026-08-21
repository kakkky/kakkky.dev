package view

import (
	"fmt"
	"html"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// github.com/yuin/goldmark の Extender インターフェースを満たす
type directiveExtension struct{}

func NewDirectiveExtension() goldmark.Extender { return directiveExtension{} }

func (directiveExtension) Extend(m goldmark.Markdown) {
	// parser : :::name args を検出して directive を生やす
	m.Parser().AddOptions(parser.WithBlockParsers(
		// paragraph parser (priority 1000) より前なら値は任意。
		// paragraph より後にすると、未知 name の :::hoge が directive として open されず生テキストで残る挙動が壊れる。
		// 700 (fenced code) と 800 (blockquote) の間 とかにしておく
		util.Prioritized(directiveParser{}, 750),
	))

	// renderer : directiveBlockNode を HTML に変換する
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(directiveRenderer{}, 500),
	))
}

type directiveSpec struct {
	open  func(w util.BufWriter, args string)
	close func(w util.BufWriter)
}

var directiveSpecs = map[string]directiveSpec{
	"message": {
		open: func(w util.BufWriter, args string) {
			// flex レイアウトで左にアイコン、右に本文。
			// アイコン色は text-* で親から継承 (SVG が fill="currentColor")。
			var bg, iconColor, icon string
			switch args {
			case "warn":
				bg, iconColor, icon = "bg-amber-100", "text-amber-600", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" fill-rule="evenodd" class="w-7 h-7 flex-shrink-0"><path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/></svg>`
			default:
				bg, iconColor, icon = "bg-blue-100", "text-blue-600", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" fill-rule="evenodd" class="w-7 h-7 flex-shrink-0"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/></svg>`
			}
			fmt.Fprintf(w,
				`<div class="my-4 rounded-lg p-4 %s flex gap-3 items-start">`+
					`<span class="%s">%s</span>`+
					`<div class="min-w-0 flex-1 [&>*:first-child]:mt-0 [&>*:last-child]:mb-0">`+"\n",
				bg, iconColor, icon)
		},
		close: func(w util.BufWriter) {
			w.WriteString("</div></div>\n")
		},
	},
	"toggle": {
		open: func(w util.BufWriter, args string) {
			fmt.Fprintf(w,
				`<details class="group my-4 rounded-md border border-gray-200 bg-gray-50 open:bg-white">`+
					`<summary class="cursor-pointer px-3 py-2 text-sm text-gray-600 [list-style:revert] group-open:border-b group-open:border-gray-200">%s</summary>`+
					`<div class="px-4 py-3 bg-gray-50 rounded-b-md [&>*:first-child]:mt-0 [&>*:last-child]:mb-0">`+"\n",
				html.EscapeString(args))
		},
		close: func(w util.BufWriter) {
			w.WriteString("</div></details>\n")
		},
	},
}

// github.com/yuin/goldmark/ast の Node インターフェースを満たす
type directiveNode struct {
	ast.BaseBlock
	Name, Arg string
}

var kindDirective = ast.NewNodeKind("Directive")

func (b *directiveNode) Kind() ast.NodeKind { return kindDirective }
func (b *directiveNode) Dump(src []byte, level int) {
	ast.DumpHelper(b, src, level, map[string]string{"Name": b.Name, "Args": b.Arg}, nil)
}

// github.com/yuin/goldmark/parser の BlockParser インターフェースを満たす
type directiveParser struct{}

func (directiveParser) Trigger() []byte {
	// 発火文字。':' で始まる行だけ Open が呼ばれる
	return []byte{':'}
}

func (directiveParser) CanInterruptParagraph() bool {
	// 段落継続中でも directive を割り込ませたい
	return true
}

func (directiveParser) CanAcceptIndentedLine() bool {
	// 4 スペース以上インデントされた行は indented code block に譲る
	return false
}

func (directiveParser) Close(ast.Node, text.Reader, parser.Context) {
	// 状態の変更等をしていないので何もしない
}

func (directiveParser) Open(_ ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || // 空文字
		pos+3 > len(line) || // ::: に満たない
		(line[pos] != ':' || line[pos+1] != ':' || line[pos+2] != ':') || // ::: が３つ続いていない
		(pos+3 < len(line) && line[pos+3] == ':') { // :::: のように4文字ある
		return nil, parser.NoChildren
	}
	name, arg := splitNameArg(string(line[pos+3:]))
	if _, ok := directiveSpecs[name]; !ok {
		return nil, parser.NoChildren
	}

	reader.AdvanceToEOL()
	return &directiveNode{Name: name, Arg: arg}, parser.HasChildren
}

func (directiveParser) Continue(_ ast.Node, reader text.Reader, _ parser.Context) parser.State {
	line, _ := reader.PeekLine()

	// 閉じ忘れとして 暗黙的にclose
	if line == nil {
		return parser.Close
	}
	w, pos := util.IndentWidth(line, reader.LineOffset())

	// 閉じ ::: 以外は継続
	if w >= 4 || // インデントが 4 スペース以上は閉じ ::: として認識しない
		pos+3 > len(line) ||
		(line[pos] != ':' || line[pos+1] != ':' || line[pos+2] != ':') ||
		!util.IsBlank(line[pos+3:]) { // ::: の後に空白以外の文字がある場合
		return parser.Continue | parser.HasChildren
	}

	// 閉じ ::: を消費して閉じる
	reader.AdvanceToEOL()
	return parser.Close
}

func splitNameArg(s string) (name string, arg string) {
	s = strings.TrimSpace(s)
	name, rest, _ := strings.Cut(s, " ")
	return name, strings.TrimSpace(rest)
}

// github.com/yuin/goldmark/renderer の NodeRenderer インターフェースを満たす
type directiveRenderer struct{}

func (directiveRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindDirective, renderDirective)
}

func renderDirective(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*directiveNode)
	d := directiveSpecs[n.Name]
	if entering {
		d.open(w, n.Arg)
	} else {
		d.close(w)
	}
	return ast.WalkContinue, nil
}
