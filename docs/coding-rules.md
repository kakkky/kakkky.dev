# Coding Rules

各層の書き方の規約。ここに載っていないケースは既存コードの近い層を読んで踏襲する。

## Handler (`app/adapter/handler/`)

- **1 handler = 1 usecase** が原則。ダッシュボードのように複数リソースをまとめて描画する場合のみ複数許容
- 命名: `<HTTPMethod><Resource>Handler` (URL の context prefix は型名に重複させない)
- 構造は 3 要素の固定形:
  - `struct` に `usecase` 依存 と必要な config 値 (公開 URL / GA measurement ID など)
  - `New<Xxx>Handler(...)` constructor
  - `ServeHTTP(rw http.ResponseWriter, r *http.Request)` を実装
- 入力の取り出し方:
  - path param → Go 1.22+ mux syntax の `r.PathValue(...)`
  - query → `r.URL.Query()`
  - form → `r.ParseForm()` してから `r.FormValue` / `r.Form[key]`
  - multipart → `r.ParseMultipartForm(maxMemory)` してから `r.FormFile`
- **エラーは context holder に載せて即 return**。`http.Error` や自前 status write は書かない。error render middleware が defer で拾う
- レスポンスは以下 5 種のうち 1 つに寄せる。同一 handler で混ぜない:
  - full page render (`pages.X(vm).Render(ctx, rw)`)
  - Turbo Frame partial (`if turbo.IsFrameRequest(r) { partials.X(vm).Render; return }` を handler 冒頭で分岐)
  - Turbo Stream (`turbo.StreamHeader(rw)` を書き込み前に → partial を render)
  - Redirect (`http.Redirect(rw, r, url, http.StatusSeeOther)`)
  - JSON (`json.Marshal` + `w.Write`)
- ViewModel の組み立ては handler 内の非公開 helper 関数に切り出す。テンプレに usecase の Output をそのまま渡さない

## Usecase (`app/usecase/`)

- シグネチャは固定:

  ```go
  func (u *<Name>Usecase) Exec(ctx context.Context, in <Name>Input) (<Name>Output, error)
  ```

- `Input` / `Output` は **値型**。pointer にしない
- エラー時は **`Output{}` (zero value) + err** を返す
- トランザクションは `repo.WithTx(ctx, func(tx domain.Repository) error { ... })` で 1 つに束ねる。tx の中では `tx.NewXxxRepository()` で sub-repository を factory から取る
- 呼び出し先が domain エラーを返してきた場合、そのまま伝播させる。usecase 層で HTTP status を意識しない

## Domain (`app/domain/`)

- **constructor でバリデーション**。不正な状態を持つエンティティのインスタンス化を禁じる
- 値オブジェクトは `type X <primitive>` の型定義で表現する。バリデーションが必要なら `NewX(...)` を用意
- エラーは 4 種類の事前定義変数 (invalid argument / not found / already exists / internal) に、`.With(msg)` (ユーザー向けメッセージのみ) か `.Wrap(cause, msg)` (原因付き) を被せて生成する。新しい種別を無闇に足さない
- Repository / QueryService / Client / Cache は **factory pattern の interface だけ** を domain に置く。実装は adapter 側
- interface ファイルの冒頭に `//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_<name>.go -package=mock` を置く

## Repository (`app/adapter/repository/`)

- struct は `sqlx.ExtContext` を受け取る。tx / 非 tx 両対応にするため
- row struct と `func (r <xxxRow>) to<Xxx>() *domain.<Xxx>` helper で domain 変換
- SQL は関数内に raw string で書く。定数化しない — クエリごとに影響範囲を局所化する
- **書き込み系のみ**。read-only の複雑クエリは `query/` へ

## Query (`app/adapter/query/`)

- **read-only の複雑クエリ、集計、複数 join、カーソルページネーション** を持つ
- 書き込みは絶対に持たせない。書き込みは repository へ
- Repository と同様に `sqlx.ExtContext` を受け取り、row → domain / view struct 変換 helper を持つ

## Client (`app/adapter/client/`)

- 外部システム連携 (オブジェクトストレージ / OGP / 外部 API など)
- domain 側は `Client` interface の factory pattern。個別クライアントは `NewXxxClient()` で取り出す
- 設定 (資格情報 / エンドポイント) は Wire で config から流し込む

## Cache (`app/adapter/cache/`)

- プロセスローカル cache
- domain 側は factory interface。TTL / eviction は実装側の関心

## Middleware (`app/adapter/middleware/`)

- HTTP 横断関心事のみ。ビジネスロジックを載せない
- 適用順序は mux 側で組み立てる (`docs/architecture.md` の "リクエストフロー" 参照)
- エラー系 middleware は `defer` + `errors.FromContext` パターンで実装する

## Logging (`app/logging/`)

- `log/slog` を使う。キー=値形式:

  ```go
  slog.Info("...", "key", value)
  slog.Error("...", "err", err)
  ```

- アクセスログのように構造化が必要なものは `slog.Group(...)` でまとめる
- ログ本文に生の HTTP body や multipart を丸ごと入れない。上限バイト数で切るか、種別で omit する

## 命名

- URL の context prefix (`/admin/...` など) を型名に重複させない
- Usecase の Input / Output は usecase 名 + `Input` / `Output`
- 値オブジェクトの型名は "型 + 意味" で単数形 (`Slug`, `TagID` など)

## Comments

- **原則書かない**
- 書くのは「なぜ」が非自明なとき (隠れた制約 / ワークアラウンド / 読み手が驚く挙動)
- 「何をしているか」は命名で伝える
- テストケースの `name` に挙動の要約を書く。テスト内コメントで補足しない
