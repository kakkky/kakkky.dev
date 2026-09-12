import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["item"]

  #dragging = null

  onDragStart(event) {
    const item = event.target.closest("[data-series-articles-target='item']")
    if (!item) return
    this.#dragging = item
    event.dataTransfer.effectAllowed = "move"
    // 一部ブラウザ で dataTransfer に何か入れないと drop が発火 しない
    event.dataTransfer.setData("text/plain", item.id)
    requestAnimationFrame(() => item.classList.add("opacity-50"))
  }

  onDragOver(event) {
    if (!this.#dragging) return
    const target = event.target.closest("[data-series-articles-target='item']")
    if (!target || target === this.#dragging) {
      event.preventDefault()
      return
    }
    event.preventDefault()
    event.dataTransfer.dropEffect = "move"

    const rect = target.getBoundingClientRect()
    const midY = rect.top + rect.height / 2
    if (event.clientY < midY) {
      target.parentNode.insertBefore(this.#dragging, target)
    } else {
      target.parentNode.insertBefore(this.#dragging, target.nextSibling)
    }
  }

  onDrop(event) {
    if (!this.#dragging) return
    event.preventDefault()
    this.#dragging.classList.remove("opacity-50")
    this.#dragging = null
  }

  onDragEnd() {
    if (!this.#dragging) return
    this.#dragging.classList.remove("opacity-50")
    this.#dragging = null
  }
}
