# Database & Migration

PostgreSQL 17 前提。マイグレーション運用、スキーマ snapshot、型選択。

## ツール

- **golang-migrate** — マイグレーション up / down 実行
- **atlas** — 現在のスキーマ全体像を snapshot として dump

両者は役割が異なる。migrate が状態を進め、atlas が結果をファイル化する。

## ディレクトリ

- `app/driver/db/migrations/` — `NNNNNN_<name>.up.sql` と `.down.sql` のペア
- `app/driver/db/schema/schema.sql` — atlas による snapshot (現在のスキーマ全体)

## snapshot は必ずコミット

`schema.sql` は生成物だがコミット対象。理由は 2 つ:

- テストで `testhelper.SetupDB` が testcontainers 上に schema を apply するとき、この snapshot を embed して使う (マイグレーションを 1 個ずつ流し直さないため)
- スキーマの現況を PR / git blame で追いたい

マイグレーション追加時に `schema.sql` を commit し忘れると **テストが古いスキーマで走ってしまう** ので必ずセットで push する。

## `migrate.up` / `migrate.down` は schema.dump を自動 kick

Makefile 側で up / down 後に `schema.dump` を呼ぶように組んである。手動で `schema.dump` を叩く必要は通常ない。**`migrate.down` は破壊的操作** なので事前確認する。

## 型選択パターン

| 用途 | 型 |
|---|---|
| 短い項目 (title / slug / status など) | `VARCHAR(n)` — 長さ上限を明示 |
| 可変長のフリーテキスト (本文 / description) | `TEXT` |
| プライマリキー | `uuid` (`DEFAULT gen_random_uuid()`) |
| 一意な人間可読識別子 (slug 等) | `UNIQUE` 制約 |
| 時刻 | `timestamptz` (`NOT NULL DEFAULT now()`、公開時刻など任意値なら nullable) |
| enum っぽい列 (status など) | `CHECK` 制約で許容値を絞る |

## テストとの結線

- testcontainers 起動時に `schema.SQL` (Go embed) を丸ごと apply
- fixture は `testhelper.Insert(t, ctx, db, testhelper.Fixtures{...})` で直接 SQL insert
- 各テストで `t.Cleanup(func() { testhelper.TruncateAll(...) })` を必ず入れる (`docs/testing.md` の "TestMain / Cleanup パターン" 参照)

## マイグレーション追加のフロー

1. `make migrate.create NAME=<snake_case>` で up / down のペアを生成
2. up / down を両方書く。**down は必ず書く** (前回の migrate.up をロールバックできる状態を維持する)
3. `make migrate.up` を叩く (schema.dump も自動で走る)
4. 生成された `schema.sql` の差分を確認
5. commit には migrations 追加分と `schema.sql` の両方を含める
