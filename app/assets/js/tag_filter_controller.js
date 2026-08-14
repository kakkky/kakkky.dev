import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["options", "option"]
  static values = { max: { type: Number, default: 10 } }

  connect() {
    this.#restoreFromURL()
    this.#updateState()
  }

  toggle(event) {
    const { slug } = event.params
    const el = event.currentTarget
    if (this.#isSelected(el)) {
      this.#deselect(el, slug)
    } else {
      if (this.#selectedCount() >= this.maxValue) return
      this.#select(el, slug)
    }
    this.#updateState()
    this.element.requestSubmit()
  }

  #restoreFromURL() {
    const slugs = new Set(new URLSearchParams(window.location.search).getAll("tag"))
    for (const slug of slugs) {
      if (this.#selectedCount() >= this.maxValue) break
      const el = this.#optionBySlug(slug)
      if (el) this.#select(el, slug)
    }
  }

  #select(el, slug) {
    el.setAttribute("aria-pressed", "true")
    this.element.appendChild(this.#createTagInput(slug))
  }

  #deselect(el, slug) {
    el.setAttribute("aria-pressed", "false")
    const input = this.element.querySelector(`input[name="tag"][value="${CSS.escape(slug)}"]`)
    if (input) input.remove()
  }

  #updateState() {
    const reached = this.#selectedCount() >= this.maxValue
    for (const el of this.optionTargets) {
      const disable = !this.#isSelected(el) && reached
      el.classList.toggle("pointer-events-none", disable)
      el.classList.toggle("opacity-60", disable)
    }
  }

  #selectedCount() {
    return this.element.querySelectorAll('input[name="tag"]').length
  }

  #isSelected(el) {
    return el.getAttribute("aria-pressed") === "true"
  }

  #optionBySlug(slug) {
    return this.optionTargets.find((el) => el.dataset.tagSlug === slug)
  }

  #createTagInput(slug) {
    const input = document.createElement("input")
    input.type = "hidden"
    input.name = "tag"
    input.value = slug
    return input
  }
}
