import { Controller } from "@hotwired/stimulus"

// edit page 全体の UX を束ねる controller。<form> に attach する。form は
// turbo-frame の swap を跨いで生き続けるので、以下の両方をここで扱える:
//
//   - キーボードショートカット
//       * Cmd/Ctrl + S : 更新 (form submit)
//       * Cmd/Ctrl + E : Preview / Edit トグル (editor toolbar のボタンを click)
//   - editor ↔ preview swap 時の scroll 位置同期
//       * 割合ベース (scrollTop / (scrollHeight - clientHeight)) で復元
//       * source では CodeMirror の .cm-scroller、preview では data-preview-scroller を対象
//       * turbo:frame-render / turbo:before-frame-render は bubble するので form で拾える
export default class extends Controller {
  static targets = ["toggle", "frame"]

  connect() {
    this.handleKeydown = (e) => this.#onKeydown(e)
    this.handleBeforeRender = (e) => this.#onBeforeRender(e)
    this.handleAfterRender = (e) => this.#onAfterRender(e)
    window.addEventListener("keydown", this.handleKeydown)
    this.element.addEventListener("turbo:before-frame-render", this.handleBeforeRender)
    this.element.addEventListener("turbo:frame-render", this.handleAfterRender)
  }

  disconnect() {
    window.removeEventListener("keydown", this.handleKeydown)
    this.element.removeEventListener("turbo:before-frame-render", this.handleBeforeRender)
    this.element.removeEventListener("turbo:frame-render", this.handleAfterRender)
  }

  #onKeydown(e) {
    if (!(e.metaKey || e.ctrlKey)) return
    const key = e.key.toLowerCase()

    if (key === "s") {
      e.preventDefault()
      this.element.requestSubmit()
      return
    }

    if (key === "e") {
      e.preventDefault()
      if (this.hasToggleTarget) this.toggleTarget.click()
    }
  }

  #onBeforeRender(e) {
    if (!this.hasFrameTarget || e.target !== this.frameTarget) return
    const s = this.#scroller()
    if (!s) return
    const denom = s.scrollHeight - s.clientHeight
    this.scrollRatio = denom > 0 ? s.scrollTop / denom : 0
  }

  #onAfterRender(e) {
    if (!this.hasFrameTarget || e.target !== this.frameTarget) return
    if (this.scrollRatio == null) return
    const target = this.scrollRatio
    // swap 直後は CodeMirror の scroller がまだ生えていない可能性があるので、
    // scroller が用意され scrollable な高さになるまで数フレーム待ってから適用する。
    let tries = 0
    const tryScroll = () => {
      const s = this.#scroller()
      const denom = s ? s.scrollHeight - s.clientHeight : 0
      if (!s || denom <= 0) {
        if (tries++ < 20) requestAnimationFrame(tryScroll)
        return
      }
      s.scrollTop = target * denom
    }
    requestAnimationFrame(tryScroll)
  }

  #scroller() {
    if (!this.hasFrameTarget) return null
    return this.frameTarget.querySelector(".cm-scroller, [data-preview-scroller]")
  }
}
