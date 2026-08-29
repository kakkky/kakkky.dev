import { Controller } from "@hotwired/stimulus"

// article-form: cmd/ctrl+s で form を送信し、送信中は save button の
// loading state をトグルする。
export default class extends Controller {
  static targets = ["submit", "submitLabel", "submitSpinner"]

  connect() {
    this.onKeydown = this.onKeydown.bind(this)
    window.addEventListener("keydown", this.onKeydown)
  }

  disconnect() {
    window.removeEventListener("keydown", this.onKeydown)
  }

  onKeydown(event) {
    if ((event.metaKey || event.ctrlKey) && event.key === "s") {
      event.preventDefault()
      this.element.requestSubmit()
    }
  }

  onSubmitStart() {
    if (!this.hasSubmitTarget) return
    this.submitTarget.disabled = true
    this.submitTarget.setAttribute("aria-busy", "true")
    if (this.hasSubmitLabelTarget) this.submitLabelTarget.classList.add("opacity-0")
    if (this.hasSubmitSpinnerTarget) this.submitSpinnerTarget.classList.remove("hidden")
  }

  onSubmitEnd() {
    if (!this.hasSubmitTarget) return
    this.submitTarget.disabled = false
    this.submitTarget.removeAttribute("aria-busy")
    if (this.hasSubmitLabelTarget) this.submitLabelTarget.classList.remove("opacity-0")
    if (this.hasSubmitSpinnerTarget) this.submitSpinnerTarget.classList.add("hidden")
  }
}
