import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["tagChips", "input", "tagSuggestions", "tagChipTemplate", "tagChip"]
  static values = {
    maxSelectedTagCount: { type: Number, default: 5 },
    existingTagNames: { type: Array, default: [] },
  }

  #chosenSuggestionIndex = -1

  connect() {
    this.#chosenSuggestionIndex = -1
    this.#updatePlaceholder()
    this.#syncInputDisabled()
    this.#hideSuggestions()
  }

  onFocus() {
    this.#renderSuggestionsList()
  }

  onBlur() {
    this.#hideSuggestions()
  }

  onInput() {
    this.#chosenSuggestionIndex = -1
    this.#renderSuggestionsList()
  }

  onKeydown(event) {
    switch (event.key) {
      case "ArrowDown":
        event.preventDefault()
        this.#moveChosenSuggestion(1)
        break
      case "ArrowUp":
        event.preventDefault()
        this.#moveChosenSuggestion(-1)
        break
      case "Escape":
        this.inputTarget.value = ""
        this.#chosenSuggestionIndex = -1
        this.#hideSuggestions()
        break
      case "Backspace":
        if (this.inputTarget.value === "") {
          event.preventDefault()
          const chips = this.tagChipTargets
          if (chips.length > 0) {
            this.#removeTag(chips[chips.length - 1].querySelector("input[type='hidden']").value)
          }
        }
        break
      case "Tab": {
        const hit = this.tagSuggestionsTarget.querySelectorAll("[data-name]")[this.#chosenSuggestionIndex]
        if (hit) {
          event.preventDefault()
          this.#addTag(hit.dataset.name)
        }
        break
      }
      case "Enter": {
        event.preventDefault()
        const hit = this.tagSuggestionsTarget.querySelectorAll("[data-name]")[this.#chosenSuggestionIndex]
        this.#addTag(hit ? hit.dataset.name : this.inputTarget.value.trim())
        break
      }
    }
  }

  removeChip({ params }) {
    this.#removeTag(params.name)
  }

  onSuggestionMousedown(event) {
    event.preventDefault()
    this.#addTag(event.currentTarget.dataset.name)
  }

  #renderSuggestionsList() {
    const q = this.inputTarget.value.trim().toLowerCase()
    const selected = new Set(
      this.tagChipTargets.map((c) => c.querySelector("input[type='hidden']").value),
    )
    const items = this.existingTagNamesValue
      .filter((name) => !selected.has(name))
      .filter((name) => q === "" || name.toLowerCase().includes(q))
      .slice(0, 8)
      .map((name) => ({ name, isNew: false }))
    const isExact = this.existingTagNamesValue.some((n) => n.toLowerCase() === q)

    // 入力文字列 が 既存タグ名と一致しておらず、 tagChip化 されていない場合は新規追加タグとみなせる
    if (q !== "" && !isExact && !selected.has(q)) {
      items.unshift({ name: this.inputTarget.value.trim(), isNew: true })
    }

    this.tagSuggestionsTarget.replaceChildren(
      ...items.map(({ name, isNew }) => {
        const li = document.createElement("li")
        li.dataset.name = name
        li.className = "cursor-pointer px-3 py-1.5 text-sm hover:bg-gray-100"
        li.setAttribute("role", "option")
        li.setAttribute("data-action", "mousedown->tag-input#onSuggestionMousedown")
        if (isNew) {
          li.textContent = `#${name} を作成`
          const badge = document.createElement("span")
          badge.className = "ml-2 rounded bg-blue-100 px-1.5 py-0.5 text-[10px] font-medium text-blue-800"
          badge.textContent = "新規"
          li.appendChild(badge)
        } else {
          li.textContent = `#${name}`
        }
        return li
      }),
    )

    if (items.length === 0) {
      this.#hideSuggestions()
      return
    }
    this.tagSuggestionsTarget.removeAttribute("hidden")
    this.#chosenSuggestionIndex = Math.max(0, Math.min(this.#chosenSuggestionIndex, items.length - 1))
    this.#applyChosenStyle()
  }

  #moveChosenSuggestion(delta) {
    const total = this.tagSuggestionsTarget.querySelectorAll("[data-name]").length
    if (total === 0) return
    this.#chosenSuggestionIndex = ((this.#chosenSuggestionIndex + delta) % total + total) % total
    this.#applyChosenStyle()
  }

  #applyChosenStyle() {
    this.tagSuggestionsTarget.querySelectorAll("[data-name]").forEach((el, i) => {
      el.classList.toggle("bg-gray-100", i === this.#chosenSuggestionIndex)
    })
  }

  #addTag(name) {
    const trimmed = name.trim()
    const alreadyAdded = this.tagChipTargets.some(
      (c) => c.querySelector("input[type='hidden']").value === trimmed,
    )
    if (trimmed === "" || alreadyAdded || this.tagChipTargets.length >= this.maxSelectedTagCountValue) return

    const chip = this.tagChipTemplateTarget.content.firstElementChild.cloneNode(true)
    chip.querySelector("[data-chip-label]").textContent = `#${trimmed}`
    chip.querySelector("input[type='hidden']").value = trimmed
    chip.querySelector("[data-remove]").setAttribute("data-tag-input-name-param", trimmed)
    this.tagChipsTarget.appendChild(chip)

    this.inputTarget.value = ""
    this.#chosenSuggestionIndex = -1
    this.#renderSuggestionsList()
    this.#updatePlaceholder()
    this.#syncInputDisabled()
  }

  #removeTag(name) {
    const chip = this.tagChipTargets.find(
      (c) => c.querySelector("input[type='hidden']").value === name,
    )
    if (!chip) return
    chip.remove()
    this.#renderSuggestionsList()
    this.#updatePlaceholder()
    this.#syncInputDisabled()
  }

  #updatePlaceholder() {
    const total = this.tagChipTargets.length
    switch (true) {
      case total >= this.maxSelectedTagCountValue:
        this.inputTarget.placeholder = `最大 ${this.maxSelectedTagCountValue} 個まで`
        break
      case total === 0:
        this.inputTarget.placeholder = "タグ名を入力"
        break
      default:
        this.inputTarget.placeholder = ""
    }
  }

  #syncInputDisabled() {
    this.inputTarget.disabled = this.tagChipTargets.length >= this.maxSelectedTagCountValue
  }

  #hideSuggestions() {
    this.tagSuggestionsTarget.setAttribute("hidden", "")
    this.#chosenSuggestionIndex = -1
  }
}
