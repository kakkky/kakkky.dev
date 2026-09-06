// article-editor: edit page の <textarea name="body"> を CodeMirror 6 の
// markdown editor に差し替える Stimulus controller。
//
// form 経由の送信は textarea を軸に動いているので、textarea 自体は残したまま
// hidden にして裏に置き、CodeMirror の doc 変更を updateListener で
// textarea.value に同期している (mount 失敗時のフォールバックにもなる)。
//
// turbo-frame#article-editor が preview 状態に swap されると DOM から外れて
// disconnect() → destroy、editor 状態に戻ると再 mount される。
import { Controller } from "@hotwired/stimulus"
import { EditorView, basicSetup } from "codemirror"
import { markdown } from "@codemirror/lang-markdown"
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language"
import { tags as t } from "@lezer/highlight"

// heading の視覚階層は basicSetup の defaultHighlightStyle (色付け) では
// 表現されないので、上に size / weight を重ねる。Lezer tag → style の対応で、
// CodeMirror が syntax tree を歩いて該当 token に自動生成 class を当ててくれる。
const mdHighlightStyle = HighlightStyle.define([
  { tag: t.heading1, fontSize: "1.8em", fontWeight: "700" },
  { tag: t.heading2, fontSize: "1.5em", fontWeight: "700" },
  { tag: t.heading3, fontSize: "1.3em", fontWeight: "700" },
  { tag: t.heading4, fontSize: "1.15em", fontWeight: "700" },
  { tag: t.heading5, fontSize: "1.05em", fontWeight: "700" },
  { tag: t.heading6, fontWeight: "700" },
  { tag: t.strong, fontWeight: "700" },
  { tag: t.emphasis, fontStyle: "italic" },
])

export default class extends Controller {
  static targets = ["body"]

  connect() {
    const body = this.bodyTarget.value

    // CodeMirror を mount する div を textarea の直前に挿入する。
    // flex-1 min-h-0 で、親 (partial の外側 flex column) の残り高さを埋める。
    this.container = document.createElement("div")
    this.container.className = "cm-host flex-1 min-h-0"
    this.bodyTarget.parentNode.insertBefore(this.container, this.bodyTarget)

    this.view = new EditorView({
      doc: body,
      parent: this.container,
      extensions: [
        basicSetup,
        markdown(),
        // heading size を上乗せ (色は defaultHighlightStyle が担当)。
        syntaxHighlighting(mdHighlightStyle),
        EditorView.lineWrapping,
        EditorView.theme({
          // .cm-editor root。turbo-frame の flex 高さを継承したいので height:100%。
          "&": { height: "100%", fontSize: "17px" },
          // scroller はスクロール可能領域。等幅 & 行間広め。
          ".cm-scroller": { fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace", lineHeight: "1.7" },
          // .cm-content は contenteditable な本文。ここに padding を当てると
          // 「テキストと border 端との内側余白」になり、クリック位置も自然に扱える。
          ".cm-content": { padding: "1.5rem 1.75rem" },
          // default の focus outline は container 側の border と干渉するので消す。
          "&.cm-focused": { outline: "none" },
          // 行番号 gutter は不要なので隠す。
          ".cm-gutters": { display: "none" },
          // active line のハイライト背景色も邪魔なので透明化。
          ".cm-activeLine": { backgroundColor: "transparent" },
        }),
        // doc が変わるたびに textarea.value を同期。form submit で
        // <textarea name="body"> 経由で最新 body が POST される。
        EditorView.updateListener.of((v) => {
          if (v.docChanged) {
            this.bodyTarget.value = v.state.doc.toString()
          }
        }),
      ],
    })

    // editor の周囲 (toolbar と editor 本体の隙間、container の border 際など、
    // .cm-content の外側) を click しても focus が当たらないと使いづらいので、
    // wrapper 全体で click を拾い、button/link/フォーム要素以外なら
    // CodeMirror に focus + カーソルを文末に置く。
    this.wrapper = this.container.parentElement
    this.handleWrapperClick = (event) => {
      if (event.target.closest(".cm-editor")) return
      if (event.target.closest("button, a, input, select, textarea")) return
      this.#focusEnd()
    }
    this.wrapper.addEventListener("click", this.handleWrapperClick)
  }

  disconnect() {
    this.wrapper.removeEventListener("click", this.handleWrapperClick)
    this.view.destroy()
    this.container.remove()
  }

  // カーソルを doc の末尾に置いて focus する。
  // 続きから書きたいケースが多いので end に寄せている (先頭にしたければ anchor: 0)。
  #focusEnd() {
    if (!this.view) return
    this.view.focus()
    this.view.dispatch({
      selection: { anchor: this.view.state.doc.length },
    })
  }
}
