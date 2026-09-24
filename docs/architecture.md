# Architecture

このプロジェクトの全体像。**どこに何があり、依存がどう流れ、リクエストがどう処理されるか** を扱う。個別の handler / usecase / エンティティ名は列挙しない — それらはコードを直接読めばよい。

## レイヤー

すべてのコードは `app/` 配下に置く。レイヤーは外側から内側へ以下:

```
app/cmd/server            エントリポイント
  ↓
app/driver/               実インフラ (db, httpserver)
  ↓
app/adapter/              HTTP に露出する境界と外部連携
  ↓
app/usecase/              ビジネスロジック
  ↓
app/domain/               エンティティ / 値オブジェクト / interface / エラー
```

依存は **外側 → 内側の一方向** のみ。`domain` は他のレイヤーに依存しない。

`app/adapter/` は責務ごとにサブディレクトリに分かれる:

- `handler/` — HTTP ハンドラ
- `middleware/` — HTTP 横断関心事
- `repository/` — RDB への書き込み・単純 CRUD
- `query/` — RDB からの read-only 複雑クエリ / 集計 / join
- `client/` — 外部システムとの通信 (S3 / OGP / GA など)
- `cache/` — プロセスローカル cache
- `view/` — Templ テンプレート

## 横断ユーティリティ

レイヤーに属さないパッケージ:

- `app/errors` — HTTP コンテキストにエラーを載せる holder
- `app/logging` — `log/slog` の初期化
- `app/config` — 環境変数からの設定ロード
- `app/script/` — ビルド時に走らせる一次スクリプト (CSS 生成など)
- `app/testhelper/` — testcontainers 起動、fixture、mock (`testhelper/mock/`)
- `app/assets/` — CSS / JS / 静的アセット、Stimulus controllers の Go 側 registry

## DI (Wire)

各層に `Set` を定義し、エントリポイントで束ねる。`domain` の interface は `wire.Bind` で `adapter` 側の実装型に結ぶ。DI 変更後は再生成が必須 (`docs/workflow.md` 参照)。

サブリポジトリ (Article / Tag / Series など) を持つ interface は **factory pattern** で表現する。呼び出し側は `repo.NewXxxRepository()` で取得する — Wire の provider をエンティティ数だけ増やさないため。同様の pattern を `QueryService` / `Client` / `Cache` にも適用する。

## リクエストフロー

```
mux (public / admin を別 wrap)
  → middleware chain
  → handler.ServeHTTP
  → usecase.Exec
  → repository / query / client / cache
  → view.Render (or Redirect / JSON)
```

Middleware chain は mux 側で reverse-loop を組む。**リスト先頭が outermost** (最初にリクエストに触れる)。順序は原則:

1. リクエストスコープの初期化 (エラー holder / トレースコンテキスト)
2. アクセスログ
3. Content-Type 既定値
4. エラー render (defer で書き戻す)
5. (admin のみ) 認証
6. NotFound (mux 直下、パターン未マッチを domain の NotFound に変換)

## エラーフロー

各層は生の Go エラーを返さない。**domain のエラー種別に `.With(msg)` か `.Wrap(cause, msg)` を被せて返す**。種別は "invalid argument / not found / already exists / internal" の 4 つに集約する。

handler は自分でレスポンスを書かず、`errors.Set(r.Context(), err)` で **コンテキストに載せて即 return**。outer の error render middleware が defer で:

1. `errors.FromContext(ctx)` で取り出す
2. 種別を HTTP status に map (invalid → 400, not found → 404, already exists → 409, internal → 500)
3. Turbo Frame / Stream リクエストなら flash 用の partial を stream で返す (400/409 は 422 に変換 — Turbo が消費できるのは 200/422 のみ)
4. それ以外は error ページを full HTML で返す
5. 5xx かつ非 panic は外部エラートラッキングに通知

panic 復旧も同じ middleware の defer で行い、internal 種別として扱う。

## 公開と管理の分離

ルーティングは 2 系統に分けて別々に wrap する:

- **public**: 誰でも触れる URL。標準の middleware だけを差す
- **admin**: 管理用 URL。public と同じ middleware に加えて認証 middleware を差す

認証は Cloudflare Access の JWT を検証し、許可メールと一致するかで判定する。不一致時は **not found 種別のエラーを set** して 404 として render し、admin URL の存在自体を隠す。開発環境ではバイパスする分岐を持つ。

## エントリポイントと config

`app/cmd/server` は main / wire / server 起動をこの順で行う。config は環境変数から `app/config` が組み立て、Wire の入力になる。`ENV` 変数で dev / prod を判別し、認証や外部連携の分岐に使う。
