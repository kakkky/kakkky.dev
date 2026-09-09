# kakkky.dev — AGENTS.md

個人ブログ (Go + Templ + Hotwire)。AI エージェント向けの実装指針。人間向けの説明は README.md を参照。

## Project Overview

- Go 1.25 / PostgreSQL 17
- View: [Templ](https://github.com/a-h/templ) + [hotwire-go](https://github.com/kakkky/hotwire-go) (Turbo + Stimulus, 自作ライブラリ)
- DI: [Google Wire](https://github.com/google/wire) (コード生成)
- DB: [pgx](https://github.com/jackc/pgx) + [sqlx](https://github.com/jmoiron/sqlx)
- Markdown → HTML: [goldmark](https://github.com/yuin/goldmark)
- Test: [testify](https://github.com/stretchr/testify) + [uber/mock](https://github.com/uber-go/mock) + [testcontainers-go](https://github.com/testcontainers/testcontainers-go)

## Commands

Makefile 経由。コンテナ操作 (dev.up 等) は原則ユーザーが手で叩く。Claude はコマンドを提示するに留める。

- `make dev.up / dev.down / dev.logs / dev.sh` — docker compose 操作
- `make wire.gen` — Wire で DI コード再生成 (interface / provider を追加したら必ず)
- `make templ.gen` — .templ → _templ.go を再生成
- `make chroma.gen` — シンタックスハイライトの CSS を再生成
- `make migrate.up / migrate.down / migrate.create NAME=xxx` — マイグレーション。up/down は自動で `schema.dump` を呼ぶ
- `make schema.dump` — atlas で `app/driver/db/schema/schema.sql` を再生成
- `make test` — host で `go test -count=1 ./...` (testcontainers を使うため compose exec ではなく host)

## Architecture

レイヤー (外側 → 内側):

```
cmd/server
  ↓
driver          httpserver / db  (実インフラ)
  ↓
adapter         handler / repository / view / middleware / client
  ↓
usecase         Input/Output 型 + Exec(ctx, in) (out, err)
  ↓
domain          エンティティ / 値オブジェクト / repository interface / エラー
```

- 依存方向は 外側 → 内側 のみ。domain は他の層に依存しない
- domain は repository の **interface だけ** を定義。実装は adapter/repository
- DI は Wire (コード生成)。`cmd/server/wire.go` の `InitServer()` → `config.Set` → `driver.Set` → `adapter.Set` → `NewServer`
- リクエストフロー: mux → middleware → `handler.ServeHTTP` → `usecase.Exec` → repository → view rendering

## Coding Rules

### Handler (`app/adapter/handler/`)

- **1 handler = 1 usecase**。fat handler にしない
- `http.Handler` を実装 (`ServeHTTP(rw, r)`)
- エラーは `RenderError(rw, r, err)` に流す。ドメインエラーは自動で適切な HTTP status にマップされる
- URL パラメータは `r.PathValue("...")`
- 例: `app/adapter/handler/get_article_handler.go`, `app/adapter/handler/error.go`

### Usecase (`app/usecase/`)

- シグネチャは `Exec(ctx context.Context, in XxxUsecaseInput) (XxxUsecaseOutput, error)`
- **Input/Output は値型** (pointer にしない)。エラー時は `Output{}` を返す
- トランザクションは `repo.WithTx(ctx, func(tx Repository) error { ... })`
- 例: `app/usecase/create_article.go`

### Naming

- **URL の context prefix を型名に重複させない**。`/admin/dashboard` なら `GetDashboardHandler` で十分 (`GetAdminDashboardHandler` は冗長)
- Input/Output 型は `<UsecaseName>Input` / `<UsecaseName>Output`

### Domain (`app/domain/`)

- constructor でバリデーション (`NewArticle` 等)。不正な状態を持つエンティティを作らせない
- 値オブジェクトは型定義で表現: `type Slug string` / `type TagID string`
- エラーは事前定義した `domain.ErrInvalidArgument` 等に `.With(msg)` でユーザー向けメッセージを付ける
- 例: `app/domain/article.go`, `app/domain/error.go`

### Repository (`app/adapter/repository/`)

- struct は `sqlx.ExtContext` を受け取る (tx / 非 tx 両対応)
- SQL の row 型 → domain entity への変換は `func (r xxxRow) toXxx() *domain.Xxx` の helper で

### Logging

- `log/slog` (キー=値形式)。`slog.Info("...", "key", val)` / `slog.Error(...)`

### Comments

- 原則書かない。書くのは「なぜ」が非自明な時 (隠れた制約・ワークアラウンド・驚く挙動) のみ
- 「何をしているか」は命名で伝える

## Testing

### スタイル

- **testify** — `require` (失敗で即終了) と `assert` (継続) を使い分け
- **常にテーブルテスト**。単発ケースでも `tests := []struct{name string, ...}` + `t.Run(tt.name, ...)` の形にする
- ケース名は `"success: ..."` / `"error: ..."` プレフィックス

### Layer 別方針

- **domain**: 単体テスト (バリデーション・状態遷移)
- **usecase**: gomock で境界テスト。モックは `testhelper/mock/` に配置、`//go:generate mockgen` で生成
- **adapter/repository**: testcontainers で実 PostgreSQL を立てる。`testhelper.SetupDB` / `testhelper.Insert` / `t.Cleanup(func(){ testhelper.TruncateAll(...) })`
- **driver**: **テストを書かない**。「動けばいい」層で、テストのために本体を歪めない

### 実行

- `make test` は host 実行 (testcontainers のため)。`-count=1` でキャッシュ無効化 (DB 触るテストの結果使い回しを防ぐ)

## DB & Migration

- ツール: [golang-migrate](https://github.com/golang-migrate/migrate) の CLI
- ディレクトリ: `app/driver/db/migrations/` (`NNNNNN_xxx.up.sql` / `.down.sql`)
- `make migrate.up` / `migrate.down` は **自動で `make schema.dump`** (atlas) を叩き `app/driver/db/schema/schema.sql` を更新
- **schema.sql snapshot は必ずコミット** (sqlc 未使用でも、現在のスキーマ全体像を残すため)
- 型選び:
  - 短い項目 (title / slug / status 等) → `VARCHAR(n)`
  - 可変長フリーテキスト (本文 / description) → `TEXT`
- ID は `uuid` (`gen_random_uuid()`)、slug は `UNIQUE`

## Frontend (Templ + Hotwire)

- view 配置: `app/adapter/view/{layout,components,pages}`
- Hotwire は自作の `github.com/kakkky/hotwire-go` を使用
  - Turbo Frame 判定: `turbo.IsFrameRequest(r)`
  - Turbo Stream レスポンス: `turbo.StreamHeader(w)`
- CSS は Tailwind (`app/assets/main.css`)。Chroma の CSS は `make chroma.gen`

## Ways of Working

- **設計を先に出す** — 新機能 / 拡張はまず spec, データ構造, 拡張ポイントを提示して合意してから実装。実装からいきなり出さない
- **コンテナはユーザーが起動** — Claude は `make dev.up` 等を勝手に走らせず、コマンド提示のみ
- **破壊的操作は事前確認** — `migrate.down`, DB 直操作, `git push --force`, ブランチ削除 等
