# Workflow

Makefile ターゲット、再生成トリガ、日々の開発フロー。

## コンテナ操作はユーザーが叩く

`make dev.up` 等のコンテナ操作は Claude が勝手に走らせない。**必要ならコマンドを提示するだけ** に留める。ユーザーが自分で叩く。

## Makefile 全ターゲット

### Docker (開発コンテナ)

| ターゲット | 目的 |
|---|---|
| `make dev.up` | compose up -d |
| `make dev.build` | app image のビルド |
| `make dev.down` | compose down |
| `make dev.restart` | app サービスの再起動 |
| `make dev.logs` | app のログを follow |
| `make dev.sh` | app コンテナに bash で入る |

### Code generation

| ターゲット | 目的 |
|---|---|
| `make wire.gen` | google/wire で `wire_gen.go` を再生成 |
| `make templ.gen` | `.templ` → `_templ.go` を再生成 |
| `make chroma.gen` | Chroma のコードハイライト CSS を `app/assets/chroma.css` に生成 |

Mock (`app/testhelper/mock/mock_*.go`) は Makefile ターゲットにはない。domain interface の `//go:generate` を使うので **`app/` で `go generate ./...`** を叩く。

### Database

| ターゲット | 目的 |
|---|---|
| `make migrate.create NAME=<snake_case>` | up / down のペアを新規作成 |
| `make migrate.up` | 未適用のマイグレーションを全部 up (schema.dump 自動 kick) |
| `make migrate.down` | 1 段戻す (schema.dump 自動 kick、**破壊的**) |
| `make migrate.version` | 現在のバージョン確認 |
| `make schema.dump` | atlas で `schema.sql` を再生成 |
| `make seed` | seed データ投入 (開発環境専用) |

### Test

| ターゲット | 目的 |
|---|---|
| `make test` | host で `go test -count=1 ./...`。testcontainers 用に host 実行、`-count=1` で cache 無効 |

## 再生成トリガ表

| 何を編集したら | 走らせるコマンド |
|---|---|
| `.templ` または ViewModel struct | `make templ.gen` |
| domain の interface / adapter の provider を追加削除 | `make wire.gen` |
| `//go:generate mockgen` のある interface | `cd app && go generate ./...` |
| migration の up/down 新規追加 | `make migrate.up` (schema.dump 自動) |
| Chroma のテーマを変更したいとき | `make chroma.gen` |

**コミット前に `git status` で生成物 (`_templ.go` / `wire_gen.go` / `mock_*.go` / `schema.sql` / `chroma.css`) が入っているか確認する** — 再生成し忘れは CI 通過を難しくする。

## 開発フロー

1. 要件を整理 → 実装プランを提示 → ユーザーと合意
2. **内側から外側の順で build up**: domain → usecase → repository / query / client → handler / view
3. 各層でテストを書きながら進める (`docs/testing.md` 参照)
4. 生成物の再生成コマンドを流す
5. `git status` で意図した差分だけになっているか確認 → commit

## 破壊的操作は事前確認

- `git push --force`, `git reset --hard`, ブランチ削除
- `make migrate.down`, DB への直接書き込み
- 依存の追加 / 削除 / 大幅バージョン変更
- CI / secrets / `.gitignore` 変更

上記は一度許可されても別の文脈で再確認する。
