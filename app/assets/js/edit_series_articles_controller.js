import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["item", "list", "newItemTemplate", "deletions"]

  #dragging = null

  onDragStart(event) {
    const item = event.target.closest(this.#itemSelector())
    if (!item) return
    this.#dragging = item
    event.dataTransfer.effectAllowed = "move"
    // 一部ブラウザ で dataTransfer に何か入れないと drop が発火 しない
    event.dataTransfer.setData("text/plain", item.id || "")
    requestAnimationFrame(() => item.classList.add("opacity-50"))
  }

  onDragOver(event) {
    if (!this.#dragging) return
    const overItem = event.target.closest(this.#itemSelector())
    if (!overItem || overItem === this.#dragging) {
      event.preventDefault()
      return
    }
    event.preventDefault()
    event.dataTransfer.dropEffect = "move"

    const rect = overItem.getBoundingClientRect()
    const midY = rect.top + rect.height / 2
    if (event.clientY < midY) {
      overItem.parentNode.insertBefore(this.#dragging, overItem)
    } else {
      overItem.parentNode.insertBefore(this.#dragging, overItem.nextSibling)
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

  addNewItem() {
    const clone = this.newItemTemplateTarget.content.firstElementChild.cloneNode(true)
    this.listTarget.appendChild(clone)
    clone.querySelector("input[name='new_article_title']")?.focus()
  }

  removeItem(event) {
    const li = event.currentTarget.closest(this.#itemSelector())
    if (!li) return
    const articleIdInput = li.querySelector("input[name='article_id']")
    if (articleIdInput) {
      const title = event.params.articleTitle ?? ""
      if (!window.confirm(`記事「${title}」を 削除 しますか?\n(ページ右上の「更新」ボタンで確定されます)`)) return
      const deletion = document.createElement("input")
      deletion.type = "hidden"
      deletion.name = "delete_article_id"
      deletion.value = articleIdInput.value
      this.deletionsTarget.appendChild(deletion)
    }
    li.remove()
  }

  removeNewItem(event) {
    const li = event.currentTarget.closest(this.#itemSelector())
    li?.remove()
  }

  #itemSelector() {
    return `[data-${this.identifier}-target='item']`
  }
}
