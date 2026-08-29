import { Controller } from "@hotwired/stimulus"

// tag-combobox: 選択済み tag を chip 表示。input に入力すると
// 既存 tag 候補を絞り込んで表示する。tab で highlight 中の候補を選択、
// enter で新規 tag を作成 (input 値をそのまま chip 化)、backspace で末尾削除。
export default class extends Controller {
  static targets = ["chips", "input", "suggestions", "template"]
  static values = {
    max: { type: Number, default: 5 },
    tags: { type: Array, default: [] },
  }

  connect() {
    this.selected = new Map()
    this.highlightIndex = -1
    this.blurTimer = null

    for (const chip of this.chipsTarget.querySelectorAll("[data-name]")) {
      this.selected.set(chip.dataset.name, chip)
    }

    this.#updateInputState()
    this.#hideSuggestions()
  }

  disconnect() {
    if (this.blurTimer) clearTimeout(this.blurTimer)
  }

  onFocus() {
    if (this.blurTimer) clearTimeout(this.blurTimer)
    this.#renderSuggestions()
  }

  onBlur() {
    // mousedown が suggestions で発火するのを待ってから hide
    if (this.blurTimer) clearTimeout(this.blurTimer)
    this.blurTimer = setTimeout(() => this.#hideSuggestions(), 150)
  }

  onInput() {
    this.highlightIndex = -1
    this.#renderSuggestions()
  }

  onKeydown(event) {
    const key = event.key

    if (key === "ArrowDown") {
      event.preventDefault()
      this.#moveHighlight(1)
      return
    }
    if (key === "ArrowUp") {
      event.preventDefault()
      this.#moveHighlight(-1)
      return
    }
    if (key === "Escape") {
      this.inputTarget.value = ""
      this.highlightIndex = -1
      this.#hideSuggestions()
      return
    }
    if (key === "Backspace" && this.inputTarget.value === "") {
      event.preventDefault()
      this.#removeLast()
      return
    }
    if (key === "Tab") {
      const highlighted = this.#currentHighlighted()
      if (highlighted) {
        event.preventDefault()
        this.#addTag(highlighted.dataset.name)
      }
      return
    }
    if (key === "Enter") {
      event.preventDefault()
      const highlighted = this.#currentHighlighted()
      if (highlighted) {
        this.#addTag(highlighted.dataset.name)
        return
      }
      const raw = this.inputTarget.value.trim()
      if (raw !== "") this.#addTag(raw)
      return
    }
  }

  removeChip(event) {
    const { name } = event.params
    this.#removeTag(name)
  }

  onSuggestionMousedown(event) {
    event.preventDefault()
    const name = event.currentTarget.dataset.name
    this.#addTag(name)
  }

  #currentHighlighted() {
    const options = this.suggestionsTarget.querySelectorAll("[data-name]")
    if (this.highlightIndex < 0 || this.highlightIndex >= options.length) return null
    return options[this.highlightIndex]
  }

  #moveHighlight(delta) {
    const options = this.suggestionsTarget.querySelectorAll("[data-name]")
    if (options.length === 0) return
    let next = this.highlightIndex + delta
    if (next < 0) next = options.length - 1
    if (next >= options.length) next = 0
    this.highlightIndex = next
    this.#applyHighlight()
  }

  #applyHighlight() {
    const options = this.suggestionsTarget.querySelectorAll("[data-name]")
    options.forEach((el, i) => {
      el.classList.toggle("bg-gray-100", i === this.highlightIndex)
    })
  }

  #renderSuggestions() {
    const q = this.inputTarget.value.trim().toLowerCase()
    const available = this.tagsValue
      .filter((name) => !this.selected.has(name))
      .filter((name) => q === "" || name.toLowerCase().includes(q))
      .slice(0, 8)

    const items = []
    const exact = this.tagsValue.some((n) => n.toLowerCase() === q)
    if (q !== "" && !exact && !this.selected.has(q)) {
      items.push({ name: this.inputTarget.value.trim(), isNew: true })
    }
    for (const name of available) items.push({ name, isNew: false })

    this.suggestionsTarget.innerHTML = ""
    if (items.length === 0) {
      this.#hideSuggestions()
      return
    }
    this.suggestionsTarget.removeAttribute("hidden")
    for (const it of items) {
      const li = document.createElement("li")
      li.dataset.name = it.name
      li.className = "cursor-pointer px-3 py-1.5 text-sm hover:bg-gray-100"
      li.setAttribute("role", "option")
      li.setAttribute("data-action", "mousedown->tag-combobox#onSuggestionMousedown")
      if (it.isNew) {
        const badge = document.createElement("span")
        badge.className = "ml-2 rounded bg-blue-100 px-1.5 py-0.5 text-[10px] font-medium text-blue-800"
        badge.textContent = "新規"
        li.textContent = `#${it.name} を作成`
        li.appendChild(badge)
      } else {
        li.textContent = `#${it.name}`
      }
      this.suggestionsTarget.appendChild(li)
    }
    if (this.highlightIndex >= items.length) this.highlightIndex = items.length - 1
    if (this.highlightIndex < 0) this.highlightIndex = 0
    this.#applyHighlight()
  }

  #addTag(name) {
    const trimmed = name.trim()
    if (trimmed === "") return
    if (this.selected.has(trimmed)) return
    if (this.selected.size >= this.maxValue) return

    const chip = this.#buildChip(trimmed)
    this.chipsTarget.appendChild(chip)
    this.selected.set(trimmed, chip)

    this.inputTarget.value = ""
    this.highlightIndex = -1
    this.#renderSuggestions()
    this.#updateInputState()
  }

  #removeTag(name) {
    const chip = this.selected.get(name)
    if (!chip) return
    chip.remove()
    this.selected.delete(name)
    this.#renderSuggestions()
    this.#updateInputState()
  }

  #removeLast() {
    const last = this.chipsTarget.querySelector("[data-name]:last-child")
    if (!last) return
    this.#removeTag(last.dataset.name)
  }

  #buildChip(name) {
    const tpl = this.templateTarget.content.cloneNode(true)
    const chip = tpl.querySelector("[data-name]")
    chip.dataset.name = name
    chip.querySelector("[data-chip-label]").textContent = `#${name}`
    const input = chip.querySelector("input[type='hidden']")
    input.value = name
    const removeBtn = chip.querySelector("[data-remove]")
    removeBtn.setAttribute("data-tag-combobox-name-param", name)
    return chip
  }

  #updateInputState() {
    const reached = this.selected.size >= this.maxValue
    this.inputTarget.disabled = reached
    this.inputTarget.placeholder = reached
      ? `tag は最大 ${this.maxValue} 個まで`
      : "tag 名を入力"
  }

  #hideSuggestions() {
    this.suggestionsTarget.setAttribute("hidden", "")
    this.highlightIndex = -1
  }
}
