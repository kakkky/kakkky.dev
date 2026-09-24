# Testing

層別のテスト方針と共通スタイル。

## 層別方針

| 層 | 方針 |
|---|---|
| domain | 単体テスト (constructor バリデーション / 状態遷移) |
| usecase | 境界テスト。repository / query / client / cache **すべてモック** |
| repository | testcontainers で実 PostgreSQL、fixture insert |
| query | 同上 |
| client | 相手方式に応じて選択 (in-process gRPC / testcontainers / httptest) |
| cache | 単体 |
| view | 変換関数 (Markdown / チャート等) のみ単体。テンプレ本体はテストしない |
| driver | **書かない**。「動けばいい」層、テストのために本体を歪めない |
| handler | **書かない**。HTTP 経由の統合的な確認で十分 |

## 共通スタイル

- **常にテーブルテスト**。単発ケースでも次の形にする:

  ```go
  tests := []struct {
      name string
      // input / want / wantErr / setup / mock などケースに応じて
  }{
      { ... },
  }
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) { ... })
  }
  ```

- ケース名は **挙動を平叙で** (`"returns empty slice when no tags exist"`) か、`"success: ..."` / `"error: ..."` prefix。プロジェクト内で混在するが 1 テスト関数内では揃える
- **testify** の `require` (setup / 前提の失敗で即中断) と `assert` (継続検証) を使い分ける
- エラー比較は `errors.Is(err, <domain-error-var>)`。`.With` / `.Wrap` されていても種別判定は残る

## testhelper (`app/testhelper/`)

以下のユーティリティが用意されている:

- PostgreSQL を testcontainers で起動し `schema.SQL` を apply、`(*sqlx.DB, cleanup)` を返す helper
- LocalStack を testcontainers で起動し S3 bucket を作成する helper
- `Fixtures` struct を受け取って直接 SQL で挿入する insert helper
- 全テーブルを `TRUNCATE ... RESTART IDENTITY CASCADE` する truncate helper

fixture は factory / builder ではなく **struct を直接組んで helper に渡す** 方式。単純ケースはテスト内で `&domain.Xxx{...}` を直接 init する。

## TestMain / Cleanup パターン (repository / query)

DB を触るパッケージは以下:

- **package 全体で 1 度だけ** container を起動 (`TestMain` で `SetupDB` → グローバル `testDB` に格納)
- **各テストの末尾で** `t.Cleanup(func() { testhelper.TruncateAll(t, ctx, testDB) })` を必ず呼ぶ

これでテスト間の DB 状態リークを防ぐ。container 起動コストは package 単位で吸収する。

## Mock 生成

- `go.uber.org/mock` (旧 golang/mock を uber がフォークしたもの)
- domain の interface ファイル冒頭に `//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_<name>.go -package=mock` を置く
- 生成先は `app/testhelper/mock/` にまとめてコミット
- 再生成は `app/` ディレクトリで `go generate ./...`

## Usecase テストの mock 引数パターン

Usecase テストはケースごとに mock の期待値を差し替えるため、テーブルの各ケースに `mock func(...)` を持たせる:

```go
tests := []struct {
    name    string
    input   <Name>UsecaseInput
    mock    func(repo, txRepo *mock.MockRepository, ...) // 必要な mock を引数で受ける
    wantErr error
}{ ... }
```

トランザクションを使うケースは `WithTx` の callback を実行するように仕込む:

```go
repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
    func(ctx context.Context, fn func(domain.Repository) error) error {
        return fn(txRepo)
    },
)
```

これで tx 内の sub-repository 呼び出しも検証できる。

## 実行

- `make test` を **host で** 走らせる (compose exec ではない)。testcontainers が host の docker socket を必要とするため
- `-count=1` を付けて cache を無効化する。DB 触るテスト結果の使い回しを防ぐ目的
