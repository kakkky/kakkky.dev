# Frontend

Templ + hotwire-go + Tailwind + Chroma + ECharts の構成と使い分け。

## View ディレクトリ (`app/adapter/view/`)

4 サブディレクトリに責務を分ける:

- `layout/` — 全ページを包む外枠 (`<html>` / `<head>` / `<body>` / ヘッダ)
- `pages/` — 各画面 1:1
- `components/` — 再利用可能な UI 部品 (バッジ / カード / ボタン / 入力など)
- `partials/` — **Turbo Frame / Turbo Stream 応答用の断片**、および Skeleton 等の状態表示。ページ全体を返さず一部だけ差し込むテンプレートはここ

`.templ` を編集したら `make templ.gen` で `_templ.go` を再生成する。

## ViewModel パターン

各テンプレは **typed struct を 1 つだけ** 受け取る。struct 名は `<TemplateName>ViewModel`。

- Handler が usecase の Output から ViewModel を組み立てる。組み立ては handler 内の非公開 helper に切り出す
- テンプレ内でビジネスロジックを持たない。表示上の分岐 (三項 / 条件クラス) までに留める
- 表示で使う config 値 (公開 URL / GA measurement ID など) も ViewModel に載せて渡す。テンプレから global を参照しない

## Base layout の役割

Base layout は以下を受け取る:

- サブタイトル (`<title>` に付ける)
- 本文の最大幅 (Tailwind クラス)
- 公開ベース URL (ヘッダのリンク生成用)
- Stimulus controllers のパス配列
- `extraHead ...templ.Component` — head に流し込む追加コンポーネント (importmap / echarts CDN script など、ページ限定で必要なもの)

`extraHead` を variadic にしておくことで、ページ側は必要なぶんだけ渡せる (0 個〜複数個)。

## hotwire-go (`github.com/kakkky/hotwire-go`)

自作の Hotwire ラッパー。以下のパターンで使う。

### Turbo

- **Frame リクエスト判定**: handler 冒頭で `if turbo.IsFrameRequest(r) { partials.X(...).Render(ctx, rw); return }` を分岐。以降は full page 用の処理
- **Stream レスポンス**: **書き込み前に** `turbo.StreamHeader(rw)` を呼び、その後に stream 用 partial を render
- **Templ 側の helper**: `turbo.Frame(id)` で Frame を宣言、`turbo.StreamAppend(target)` / `turbo.StreamUpdate(target)` / `turbo.StreamRemove(target)` で DOM 操作を表現、`turbo.AttrSrc(url)` / `turbo.AttrLoadingLazy()` で lazy frame を作る
- Frame ID は `const XxxID = "..."` としてテンプレファイル内に定義し、両側 (作る側と参照する側) で共有

### Stimulus

- **controllers の登録**: base layout の `stimulus.ScriptLoad(controllers...)` に path 配列で渡す
- **属性の埋め込み**: テンプレ内で `stimulus.AttrControllers(name)`, `AttrTargets(controller, name)`, `AttrValue(controller, key, value)`, `AttrActions(...)`, `AttrActionParam(...)` を `...` で spread する

## Stimulus controller registry (`app/assets/js/`)

JS 側の Stimulus controller は 2 箇所で管理する:

- 実体は `app/assets/js/*_controller.js` (静的アセットとして serve)
- Go 側の registry (`app/assets/js/controllers.go`) に **typed な `Controller{Name, Path}` 変数として宣言**

これによりページ側は `[]string{js.<XxxController>.Path, ...}` で参照でき、path の typo を型で防げる。controller を追加するときは JS ファイル追加 + registry 追加をセットで行う。

## Import Map (`app/assets/js/importmap.go`)

ES modules を CDN 経由で使うページ用。以下の pattern:

- 用途別 (CodeMirror / その他) の entry set を `map[string]string` で用意
- `js.ImportMap(entrySets ...map[string]string)` に **variadic で複数渡し、内部で merge** して 1 つの `<script type="importmap">` を返す
- ページ側は `layout.Base(..., js.ImportMap(js.<XxxImports>, ...))` の形で extraHead に流す

## ECharts (`app/adapter/view/echarts.go` + `app/assets/js/echarts_script.go`)

- **サーバサイドで** `<div>+<script>` の snippet を組み立てて `template.HTML` として ViewModel に埋める
- テンプレは `@templ.Raw(vm.LineChartHTML)` で raw HTML として render
- ECharts 本体は CDN。**チャートを使うページだけ** extraHead で CDN script (defer) を差し込む

## Tailwind + 独自 CSS (`app/assets/main.css`)

- Tailwind を import しつつ、記事本文用の typography (`.article-body :is(h1, h2, ...)`) やアニメーション keyframes は手書き
- Chroma のシンタックスハイライト CSS は `@import "./chroma.css" layer(base);` として base layer に流す
- `@source` で Tailwind のスキャン対象に `../adapter/view/**/*.templ` と `.go` を追加

## Chroma CSS (`app/script/gen_chroma_css/`)

- Chroma の style を指定した generator が stdout に CSS を吐く
- `make chroma.gen` が `app/assets/chroma.css` にリダイレクトして生成物を書き戻す
- 生成物はコミット済み。テーマ変更時だけ再生成する

## Handler ↔ view の橋渡し

前述の handler ルールを繰り返す (`docs/coding-rules.md` の Handler 節と重複するが、view 側からも参照される):

- Turbo Frame 分岐は handler の冒頭に置く
- Turbo Stream は `turbo.StreamHeader(rw)` を必ず書き込み前に
- Redirect は `http.StatusSeeOther` (303)
- エラーは handler 内で自分で render せず context holder に載せる。中で error page を出すのは error render middleware
