# kakkky.dev — AGENTS.md

AI エージェント向けの実装指針。人間向けの説明は [README.md](README.md) を参照。

## Project Overview

- 個人ブログ / Go 1.25 / PostgreSQL 17
- View: [Templ](https://github.com/a-h/templ) + 自作の [hotwire-go](https://github.com/kakkky/hotwire-go) (Turbo + Stimulus) + Tailwind
- DI: [Google Wire](https://github.com/google/wire) (コード生成)
- DB: [pgx](https://github.com/jackc/pgx) + [sqlx](https://github.com/jmoiron/sqlx) + [golang-migrate](https://github.com/golang-migrate/migrate) + [atlas](https://atlasgo.io/)
- Markdown → HTML: [goldmark](https://github.com/yuin/goldmark)
- Test: [testify](https://github.com/stretchr/testify) + [uber-go/mock](https://github.com/uber-go/mock) + [testcontainers-go](https://github.com/testcontainers/testcontainers-go)
- コードはすべて `app/` 配下。エントリポイントは `app/cmd/server/`

## Quick Commands

Makefile 経由。詳細と再生成トリガは [docs/workflow.md](docs/workflow.md)。

- `make dev.up / dev.down / dev.logs / dev.sh` — docker compose 操作
- `make wire.gen` — Wire で DI コード再生成
- `make templ.gen` — `.templ` → `_templ.go` 再生成
- `make chroma.gen` — Chroma のハイライト CSS 再生成
- `make migrate.up / migrate.down / migrate.create NAME=xxx` — マイグレーション (`schema.dump` 自動 kick)
- `make schema.dump` — `schema.sql` を再生成
- `make test` — host で `go test -count=1 ./...`
- `cd app && go generate ./...` — mockgen 系の再生成

## Ways of Working

### Confirmation Gates (破壊的操作は事前確認)

以下は一度許可されても、別の文脈で再度確認する:

- 破壊的 git 操作: `git push --force`, `git reset --hard`, `git clean -fd`, ブランチ削除, タグ削除
- ファイルシステム破壊: `rm -rf`, ディレクトリ削除, 未コミット変更の上書き
- データベース書き込み / DDL / `make migrate.down`
- 本番デプロイ, 外部 API 送信, パッケージ公開
- 依存関係の追加 / 削除 / 大幅バージョン変更
- 設定ファイル (`.gitignore` / CI / secrets) の変更

### 進め方

- **設計を先に出す** — 新機能・拡張はまず spec / データ構造 / 拡張ポイントを提示し合意してから実装
- **コンテナはユーザーが叩く** — `make dev.up` 等を Claude が勝手に走らせない。コマンド提示のみ
- **同じ修正を 2 回以上指摘されたら根本原因を見直す**

### コード方針の要点

詳細は [docs/coding-rules.md](docs/coding-rules.md)。抜粋:

- 既存パターンを踏襲する。ただしアンチパターンには追従せず指摘する
- コメントは「なぜ」が非自明なときだけ
- タスクが要求していない機能 / 抽象化 / リファクタを足さない
- 不要コード / 後方互換シム / `// removed` コメントは残さない

## Docs Index

- [docs/architecture.md](docs/architecture.md) — レイヤー構造、DI、リクエストフロー、エラーフロー
- [docs/coding-rules.md](docs/coding-rules.md) — 各層のコーディング規約
- [docs/testing.md](docs/testing.md) — 層別テスト方針とスタイル
- [docs/frontend.md](docs/frontend.md) — Templ、hotwire-go、静的アセット、チャート
- [docs/db.md](docs/db.md) — マイグレーション運用、型選択
- [docs/workflow.md](docs/workflow.md) — Makefile 全ターゲット、再生成トリガ、開発フロー
