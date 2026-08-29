// Turbo Streams の custom actions を登録する module。
// layout の <script type="module" src="..."> から一度だけ読み込まれる。
//
// import 元は turbo.ScriptImport() と同じ jsDelivr CDN。 ESM は URL 単位で
// singleton なので、Turbo Drive が使う StreamActions と同じ instance を触る。
import { StreamActions } from "https://cdn.jsdelivr.net/npm/@hotwired/turbo@8.0.23/dist/turbo.es2017-esm.js"

// replace-url: history.replaceState で URL bar を差し替える。
// 使い方: <turbo-stream action="replace-url" data-src="/new/path"></turbo-stream>
StreamActions["replace-url"] = function () {
  const src = this.getAttribute("data-src")
  if (src) history.replaceState({}, "", src)
}
