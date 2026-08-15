-- 開発環境用の seed データ
-- 冪等性のため関連テーブルごと TRUNCATE してから固定 UUID で INSERT する

BEGIN;

TRUNCATE TABLE
    article_tags,
    series_articles,
    series_tags,
    articles,
    series,
    tags
RESTART IDENTITY CASCADE;

-- ---- tags (10) ----
INSERT INTO tags (id, slug, name) VALUES
    ('11111111-1111-1111-1111-000000000001', 'go',           'Go'),
    ('11111111-1111-1111-1111-000000000002', 'typescript',   'TypeScript'),
    ('11111111-1111-1111-1111-000000000003', 'react',        'React'),
    ('11111111-1111-1111-1111-000000000004', 'docker',       'Docker'),
    ('11111111-1111-1111-1111-000000000005', 'postgres',     'PostgreSQL'),
    ('11111111-1111-1111-1111-000000000006', 'nextjs',       'Next.js'),
    ('11111111-1111-1111-1111-000000000007', 'aws',          'AWS'),
    ('11111111-1111-1111-1111-000000000008', 'testing',      'Testing'),
    ('11111111-1111-1111-1111-000000000009', 'ddd',          'DDD'),
    ('11111111-1111-1111-1111-000000000010', 'architecture', 'Architecture');

-- ---- articles (100) ----
-- カテゴリごとに generate_series でまとめて生成する。
-- UUID は 22222222-2222-2222-2222-000000000001 〜 000000000100 の連番で固定する。
-- n が特定値の記事だけ draft (published_at は NULL)、それ以外は published とする。
-- published_at は 2024-01-01 10:00:00+09 を起点に (n-1) * 8 日 ずつずらして時系列に散らす。

-- Go: 001-025
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'go-tips-' || lpad(n::text, 3, '0'),
    'Go Tips #' || lpad(n::text, 3, '0'),
    E'# Go Tips #' || lpad(n::text, 3, '0') || E'\n\nGo に関する Tips 記事 (No.' || n || E')。\n\n```go\nfunc main() { fmt.Println("hello, go") }\n```\n',
    CASE WHEN n = ANY(ARRAY[5, 17]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[5, 17]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(1, 25) n;

-- TypeScript: 026-045
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'ts-tips-' || lpad(n::text, 3, '0'),
    'TypeScript Tips #' || lpad(n::text, 3, '0'),
    E'# TypeScript Tips #' || lpad(n::text, 3, '0') || E'\n\nTypeScript に関する Tips 記事 (No.' || n || E')。\n\n```ts\nfunction pick<T, K extends keyof T>(o: T, k: K[]): Pick<T, K> { return {} as any }\n```\n',
    CASE WHEN n = ANY(ARRAY[33, 42]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[33, 42]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(26, 45) n;

-- React: 046-060
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'react-tips-' || lpad(n::text, 3, '0'),
    'React Tips #' || lpad(n::text, 3, '0'),
    E'# React Tips #' || lpad(n::text, 3, '0') || E'\n\nReact に関する Tips 記事 (No.' || n || E')。\n\n```tsx\nexport function Hello() { return <p>hi</p> }\n```\n',
    CASE WHEN n = ANY(ARRAY[55]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[55]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(46, 60) n;

-- PostgreSQL: 061-072
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'pg-tips-' || lpad(n::text, 3, '0'),
    'PostgreSQL Tips #' || lpad(n::text, 3, '0'),
    E'# PostgreSQL Tips #' || lpad(n::text, 3, '0') || E'\n\nPostgreSQL に関する Tips 記事 (No.' || n || E')。\n\n```sql\nSELECT count(*) FROM articles WHERE status = ''published'';\n```\n',
    CASE WHEN n = ANY(ARRAY[66]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[66]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(61, 72) n;

-- Docker: 073-082
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'docker-tips-' || lpad(n::text, 3, '0'),
    'Docker Tips #' || lpad(n::text, 3, '0'),
    E'# Docker Tips #' || lpad(n::text, 3, '0') || E'\n\nDocker に関する Tips 記事 (No.' || n || E')。\n\n```yaml\nservices:\n  app:\n    build: .\n```\n',
    CASE WHEN n = ANY(ARRAY[78]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[78]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(73, 82) n;

-- Next.js: 083-092
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'nextjs-tips-' || lpad(n::text, 3, '0'),
    'Next.js Tips #' || lpad(n::text, 3, '0'),
    E'# Next.js Tips #' || lpad(n::text, 3, '0') || E'\n\nNext.js に関する Tips 記事 (No.' || n || E')。\n\n```tsx\nexport default function Page() { return <main>hi</main> }\n```\n',
    CASE WHEN n = ANY(ARRAY[89]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[89]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(83, 92) n;

-- AWS: 093-097
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'aws-tips-' || lpad(n::text, 3, '0'),
    'AWS Tips #' || lpad(n::text, 3, '0'),
    E'# AWS Tips #' || lpad(n::text, 3, '0') || E'\n\nAWS に関する Tips 記事 (No.' || n || E')。\n',
    'published',
    ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval)
FROM generate_series(93, 97) n;

-- Testing: 098-100
INSERT INTO articles (id, slug, title, body, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'testing-tips-' || lpad(n::text, 3, '0'),
    'Testing Tips #' || lpad(n::text, 3, '0'),
    E'# Testing Tips #' || lpad(n::text, 3, '0') || E'\n\nテストに関する Tips 記事 (No.' || n || E')。\n',
    'published',
    ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval)
FROM generate_series(98, 100) n;

-- 長文サンプル (レイアウト / TOC / コード / callout / toggle 検証用)
INSERT INTO articles (id, slug, title, body, status, published_at) VALUES (
    '22222222-2222-2222-2222-000000000999'::uuid,
    'sample-long-article',
    'Go + DDD で個人ブログを設計する',
    $body$
# Go + DDD で個人ブログを設計する

この記事は、個人ブログ (kakkky.dev) を Go + DDD の考え方で設計・実装している過程で得られた知見をまとめたものです。フレームワークを極力使わず、標準ライブラリと薄いユーティリティで組み立てる方針で進めています。読み手として想定しているのは、業務で Go を書いた経験があり、DDD の基礎用語 (エンティティ、値オブジェクト、リポジトリ) が分かる方です。全体像 → 各レイヤの設計 → テスト戦略 → UI 実装 → まとめ、の順で進みます。

:::info
本記事はライブな設計メモです。実装が進むにつれて記述内容が古くなる可能性があります。最新の設計判断はソースコード (`app/` 以下) を優先してください。
:::

## 全体アーキテクチャ

まずは全体構成から。ここが揃っていないと個別の議論がブレるので、最初にコンテキストを共有しておきます。

### レイヤ構成

大枠は「オニオンアーキテクチャの薄い版」で、次の 4 レイヤに分けています。

- `domain`: エンティティ、値オブジェクト、リポジトリ interface、ドメインエラー
- `usecase`: ユースケース (アプリケーションサービス)。Input/Output は DTO
- `adapter`: HTTP handler、view (templ)、リポジトリ実装、middleware
- `driver`: 実行時のインフラ (HTTP サーバ、DB 接続、schema、seed)

依存の向きは `driver → adapter → usecase → domain` のみ許可します。domain は他のどのレイヤも参照しません。これは Go の import グラフを見ればすぐ確認できるので、レビュー時にも判定しやすいです。

### 依存の向きを守るための仕組み

Go には Java の `module-info` のような強制手段が無いため、規約と CI で守ります。具体的には以下 3 つを組み合わせています。

1. **パッケージ命名の一貫性**: `domain/xxx.go` は必ず `package domain`
2. **CI での import チェック**: `go list -f '{{.ImportPath}} {{.Imports}}'` を grep して禁止依存を検出
3. **PR レビューでの目視**: 数が少ないうちは人力で十分

:::info
最初から重厚な依存管理ツール (例えば `go-arch-lint`) を入れるより、まずは grep レベルで運用してから必要に応じて足すのが個人開発ではちょうどよいバランスです。
:::

## ドメイン層のモデリング

ドメイン層はビジネス上の「不変条件」を表現する場所です。ここが痩せているとロジックが usecase に漏れ、テストがしにくくなります。

### 値オブジェクトとエンティティ

値オブジェクトは「等価性が値の等価性で決まる」もの、エンティティは「同一性が ID で決まる」ものです。ブログドメインでは次のように分けました。

- 値オブジェクト: `Slug`, `Title`, `Body`, `ArticleStatus`
- エンティティ: `Article`, `Series`, `Tag`

値オブジェクトはコンストラクタでバリデーションを行い、生成後は不変です。Go は immutable な struct を厳密に強制できませんが、フィールドを非公開にしてコンストラクタからのみ生成させる、というパターンで運用しています。

```go
package domain

import (
	"fmt"
	"unicode/utf8"
)

const (
	ArticleTitleMaxLength = 100
	ArticleBodyMaxLength  = 50000
)

type Article struct {
	ID          ArticleID
	Slug        Slug
	Title       string
	Body        string
	Status      ArticleStatus
	PublishedAt time.Time
	TagIDs      []TagID
}

func NewArticle(slug Slug, title, body string, status ArticleStatus) (*Article, error) {
	if title == "" {
		return nil, ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(title) > ArticleTitleMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("タイトル は %d 文字以内 です", ArticleTitleMaxLength))
	}
	return &Article{Slug: slug, Title: title, Body: body, Status: status}, nil
}
```

:::info
「Body が空でない」といった不変条件は、DB のスキーマ制約 (NOT NULL) と二重にチェックしています。DB 制約は多層防御として重要ですが、ドメイン層のエラーメッセージのほうがユーザーに親しみやすいので、まずアプリで弾く方針です。
:::

### リポジトリの契約

リポジトリは domain 層で interface のみ定義し、実装は `adapter/repository` に置きます。この構造は使い古された定石ですが、テストのしやすさに直結するので必須です。

```go
package domain

type ArticleRepository interface {
	FindBySlug(ctx context.Context, slug Slug) (*Article, error)
	FindByIDs(ctx context.Context, ids ...ArticleID) ([]*Article, error)
}
```

`FindBySlug` は「見つからなかった場合」を `domain.ErrNotFound` で表現します。sentinel error パターンです。呼び出し側は `errors.Is(err, domain.ErrNotFound)` で判定します。

:::warn
`sql.ErrNoRows` をそのまま返してしまうと、ドメイン層が database/sql に依存してしまいます。必ず repository 実装内で `domain.ErrNotFound` に変換してから返してください。
:::

## ユースケース層の設計

usecase 層は「1 リクエスト = 1 ユースケース」の粒度で切ります。Handler は薄く、パースと呼び出しだけを担当し、実処理は usecase に閉じ込めます。

### Input/Output DTO

usecase の Input/Output は **値型の DTO** として定義します。ポインタで持たない理由は次の通りです。

- 不変データなのでポインタで共有する意味がない
- nil ゼロ値 の区別を呼び出し側に強いてしまう
- テストの `want` にゼロ値を書くのが自然になる

```go
type GetArticleUsecaseInput struct {
	Slug domain.Slug
}

type GetArticleUsecaseOutput struct {
	Article domain.Article
	Tags    map[domain.TagID]domain.Tag
}

func (us *GetArticleUsecase) Exec(ctx context.Context, in GetArticleUsecaseInput) (GetArticleUsecaseOutput, error) {
	article, err := us.articleRepo.FindBySlug(ctx, in.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return GetArticleUsecaseOutput{}, nil
		}
		return GetArticleUsecaseOutput{}, err
	}
	tags, err := us.tagRepo.FindByIDs(ctx, article.TagIDs...)
	if err != nil {
		return GetArticleUsecaseOutput{}, err
	}
	tagsByID := make(map[domain.TagID]domain.Tag, len(tags))
	for _, t := range tags {
		tagsByID[t.ID] = *t
	}
	return GetArticleUsecaseOutput{Article: *article, Tags: tagsByID}, nil
}
```

Tags を map で持つのは、Article 側の `TagIDs` 順に UI で並べたいからです。呼び出し側は `for _, id := range article.TagIDs { tag := out.Tags[id] }` と書けます。

:::toggle Article.TagIDs の順序保証について
`TagIDs` の順序は repository の SQL で `ORDER BY tag_id` を付けて安定化させています。順序が UI に影響する箇所では、SQL レベルで順序を確定させておく方が、Go 側でソートし直すより意図が明確です。
:::

### エラー方針

usecase から返るエラーは 3 種類に集約しています。

- `domain.ErrInvalidArgument`: 入力バリデーション違反
- `domain.ErrNotFound`: 対象が存在しない
- `domain.ErrInternal`: それ以外の予期しないエラー

handler 側で `errors.Is` で分岐し、それぞれ 400 / 404 / 500 にマップします。この設計だと handler が薄く保てるうえ、新しいユースケースを足したときに handler 側の追随が最小限で済みます。

## インフラ層

DB 選定、SQL 実行、マイグレーションなど、外部リソースに触れる部分をここに集約します。

### PostgreSQL の選定理由

個人ブログには過剰ですが、PostgreSQL を採用しています。理由は次の通り。

- 型が豊富 (`uuid`, `jsonb`, 配列型など)
- Full-text search が組み込みで使える (将来の検索機能への布石)
- 業務で使っているので運用ノウハウが流用できる
- ローカル docker で完結する

:::info
Full-text search は現時点で未実装ですが、`tsvector` カラムを生成列として持たせて index を張る形を想定しています。移行時には既存記事に対する backfill が必要なので、マイグレーションを分割して段階的にリリースします。
:::

### sqlx との付き合い方

`database/sql` の生 API はボイラプレートが多いので、`sqlx` を薄く被せています。ORM を使わない方針の理由:

- SQL は書けたほうがデバッグが早い
- ORM の DSL を学ぶコストを避けたい
- N+1 を意識しやすいコードにしたい

以下は Article を Slug で取ってくる例です。struct タグベースのマッピングで済むので、コード量が最小限で済みます。

```go
func (ar *ArticleRepository) FindBySlug(ctx context.Context, slug domain.Slug) (*domain.Article, error) {
	var row articleRow
	if err := sqlx.GetContext(ctx, ar.db, &row, `
SELECT a.id::text     AS id,
       a.slug         AS slug,
       a.title        AS title,
       a.body         AS body,
       a.status       AS status,
       a.published_at AS published_at,
       ARRAY(SELECT tag_id::text FROM article_tags WHERE article_id = a.id ORDER BY tag_id) AS tag_ids
FROM articles a
WHERE a.slug = $1
`, string(slug)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound.With("article not found")
		}
		return nil, domain.ErrInternal.Wrap(err, "find article by slug")
	}
	return row.toArticle(), nil
}
```

:::info
`ARRAY(SELECT ... ORDER BY ...)` で関連テーブルの ID を配列として一発で取ってくると、N+1 を回避しつつコードもシンプルに保てます。ただし件数が数百を超えると PostgreSQL 側の配列生成コストが無視できなくなるので、その場合は素直に JOIN + アプリ側で分解するほうが良いです。
:::

## テスト戦略

テストは「ユニット」「統合」「E2E」の 3 層で構成しますが、個人開発では E2E は薄く、ユニットと統合を厚くします。

### ユニットテスト

domain と usecase はモック (`gomock`) を使ってユニットテストします。テーブルテスト + `testify` で統一しています。

```go
func TestGetArticleUsecase_Exec(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		input   GetArticleUsecaseInput
		mock    func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		want    GetArticleUsecaseOutput
		wantErr error
	}{
		{
			name:  "success",
			input: GetArticleUsecaseInput{Slug: "s"},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				ar.EXPECT().FindBySlug(ctx, domain.Slug("s")).Return(&domain.Article{ID: "a1"}, nil)
				tr.EXPECT().FindByIDs(ctx).Return([]*domain.Tag{}, nil)
			},
			want: GetArticleUsecaseOutput{Article: domain.Article{ID: "a1"}, Tags: map[domain.TagID]domain.Tag{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { /* ... */ })
	}
}
```

単発ケースでも `tests` スライスに入れる、というルールを敷いています。ケース追加時に diff が小さくなり、レビューがしやすくなります。

### 統合テスト

repository は `testcontainers-go` で本物の PostgreSQL を立てて統合テストします。SQL をモックしないのは、DB 側のスキーマ変更や PostgreSQL 固有の挙動 (配列型、`ORDER BY` の順序など) に依存する箇所が多く、モックだと安心感が得られないためです。

:::warn
統合テストは並列実行時に schema/tenant がぶつからないよう、テストごとにトランザクション or 一時 schema を切ります。CI では `-parallel 1` で愚直に直列実行するのが最も安定します。
:::

### driver 層のテストは書かない

`driver/*` は「動けばよい」薄い層に留めます。ここに単体テストを書き始めると本体を歪めてしまう (テストのために interface を切る、依存を注入可能にする、など) ので、統合テストで間接的にカバーします。

## UI レイヤ

Go の web UI 開発体験は年々改善しており、`templ` + `Tailwind` + `Hotwire` の組み合わせがちょうどよく感じます。

### templ + Tailwind

`templ` は Go 製のテンプレートエンジンで、`.templ` ファイルから型安全な Go コードを生成してくれます。React 的に「コンポーネント」を関数として書けるので、Go の関数分割の感覚がそのまま UI に持ち込めます。

```templ
templ Article(vm ArticleViewModel) {
	@layout.Base(vm.Title) {
		<article class="min-w-0">
			<h1 class="text-3xl font-bold">{ vm.Title }</h1>
			<div class="article-body">
				@templ.Raw(renderMarkdown(vm.BodyMD))
			</div>
		</article>
	}
}
```

Tailwind は standalone binary をコンテナに焼き込む形で使用しています。npm/Node を挟まないので、Go だけで完結する開発環境を維持できます。

### Hotwire (Turbo/Stimulus)

`hotwire-go` を使うと Rails の Hotwire がそのまま Go でも使えます。Turbo Drive によるページ遷移の高速化、Turbo Frames による部分更新、Stimulus による軽い JS 制御を組み合わせて、SPA レスに近い体験を最小コストで実現できます。

:::info
`<turbo-frame>` の中のリンクを踏むと、Turbo は同じ frame ID を持つ要素をレスポンス内で探します。詳細ページなど「別コンテキスト」に飛ばしたいときは `data-turbo-frame="_top"` を付けて frame を抜けさせるのを忘れずに。
:::

## 参考リンク

外部の参考資料をいくつか。単独行の URL は link preview カードに展開されます。

https://go.dev

https://templ.guide

https://turbo.hotwired.dev

## まとめ

Go + DDD で個人ブログを設計するにあたり、以下の指針で進めています。

- **layered architecture** を軽量に運用: 依存の向きだけ守り、細かい規約は追加しない
- **DTO は値型**: usecase の Input/Output は迷わず値渡し
- **standard library ファースト**: `sqlx` や `templ` のような薄いユーティリティは許容、重厚な ORM/フレームワークは避ける
- **テストはユニットと統合を厚く**: driver 層は「動けばいい」で割り切る
- **UI は Hotwire で SPA レス**: React を持ち込まず、Go の型システムで完結させる

ここまでで約 1 万字ほどの分量になりました。目次 (右サイドバー) の展開、コードブロックのハイライト、コールアウト (`:::info` など)、トグル (`:::toggle`) の描画などが正しく反映されているか、この記事一つで一通り確認できます。今後、実装が進むにつれて設計方針をアップデートしていくので、続報は Feed からどうぞ。
$body$,
    'published',
    '2026-08-15 10:00:00+09'
);

INSERT INTO article_tags (article_id, tag_id) VALUES
    ('22222222-2222-2222-2222-000000000999', '11111111-1111-1111-1111-000000000001'), -- go
    ('22222222-2222-2222-2222-000000000999', '11111111-1111-1111-1111-000000000009'), -- ddd
    ('22222222-2222-2222-2222-000000000999', '11111111-1111-1111-1111-000000000010'); -- architecture

-- ---- article_tags ----
-- カテゴリごとに primary tag を付ける。generate_series でまとめて INSERT する。
INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000001'::uuid -- go
FROM generate_series(1, 25) n;

INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000002'::uuid -- typescript
FROM generate_series(26, 45) n;

INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000003'::uuid -- react
FROM generate_series(46, 60) n;

INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000005'::uuid -- postgres
FROM generate_series(61, 72) n;

INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000004'::uuid -- docker
FROM generate_series(73, 82) n;

INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000006'::uuid -- nextjs
FROM generate_series(83, 92) n;

INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000007'::uuid -- aws
FROM generate_series(93, 97) n;

INSERT INTO article_tags (article_id, tag_id)
SELECT ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
       '11111111-1111-1111-1111-000000000008'::uuid -- testing
FROM generate_series(98, 100) n;

-- 追加のタグ付け (複数タグを持つ記事を混ぜて多様性を持たせる)
INSERT INTO article_tags (article_id, tag_id) VALUES
    ('22222222-2222-2222-2222-000000000020', '11111111-1111-1111-1111-000000000009'), -- Go x DDD
    ('22222222-2222-2222-2222-000000000021', '11111111-1111-1111-1111-000000000009'), -- Go x DDD
    ('22222222-2222-2222-2222-000000000022', '11111111-1111-1111-1111-000000000009'), -- Go x DDD
    ('22222222-2222-2222-2222-000000000010', '11111111-1111-1111-1111-000000000010'), -- Go x Architecture
    ('22222222-2222-2222-2222-000000000083', '11111111-1111-1111-1111-000000000003'), -- Next.js x React
    ('22222222-2222-2222-2222-000000000084', '11111111-1111-1111-1111-000000000003'), -- Next.js x React
    ('22222222-2222-2222-2222-000000000085', '11111111-1111-1111-1111-000000000003'), -- Next.js x React
    ('22222222-2222-2222-2222-000000000093', '11111111-1111-1111-1111-000000000004'), -- AWS x Docker
    ('22222222-2222-2222-2222-000000000094', '11111111-1111-1111-1111-000000000004'); -- AWS x Docker

-- ---- series (10) ----
INSERT INTO series (id, slug, title, description, status, published_at) VALUES
    ('33333333-3333-3333-3333-000000000001', 'go-fundamentals',   'Go 基礎講座',              'Go を書く上で押さえておきたい基本トピック連載。',           'published_completed', '2024-01-15 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000002', 'go-advanced',       'Go 応用パターン',          'ゴルーチン設計・エラー戦略など応用トピック連載。',         'published_ongoing',   '2024-03-01 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000003', 'ddd-in-go',         'Go で学ぶ DDD',            'ドメインモデルを Go でどう表現するかを扱う連載。',         'published_ongoing',   '2024-05-01 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000004', 'ts-fundamentals',   'TypeScript 基礎講座',      '型の基本から実務パターンまでの TypeScript 連載。',         'published_completed', '2024-06-15 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000005', 'react-basics',      'React 入門',               'React の基本 hook と設計パターン連載。',                   'published_ongoing',   '2024-08-01 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000006', 'pg-mastery',        'PostgreSQL 実践',          'Index・トランザクション・パフォーマンス連載。',            'published_completed', '2024-10-01 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000007', 'docker-basics',     'Docker 入門',              'コンテナビルドと開発環境構築の連載。',                     'published_completed', '2024-11-15 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000008', 'nextjs-app-router', 'Next.js App Router 実践',  'App Router の設計と Server Components の実践連載。',       'published_ongoing',   '2025-01-01 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000009', 'aws-basics',        'AWS 入門',                 'IAM / VPC / ECS など AWS の基礎連載。',                    'published_ongoing',   '2025-03-01 10:00:00+09'),
    ('33333333-3333-3333-3333-000000000010', 'testing-strategy',  'テスト戦略',               'ユニット / 統合 / E2E をどう組み合わせるかの連載。',       'published_completed', '2025-05-01 10:00:00+09');

-- ---- series_articles ----
-- series_articles.PRIMARY KEY は article_id なので 1 記事は 1 series
INSERT INTO series_articles (series_id, article_id, position) VALUES
    -- go-fundamentals
    ('33333333-3333-3333-3333-000000000001', '22222222-2222-2222-2222-000000000001', 1),
    ('33333333-3333-3333-3333-000000000001', '22222222-2222-2222-2222-000000000002', 2),
    ('33333333-3333-3333-3333-000000000001', '22222222-2222-2222-2222-000000000003', 3),
    -- go-advanced
    ('33333333-3333-3333-3333-000000000002', '22222222-2222-2222-2222-000000000010', 1),
    ('33333333-3333-3333-3333-000000000002', '22222222-2222-2222-2222-000000000011', 2),
    ('33333333-3333-3333-3333-000000000002', '22222222-2222-2222-2222-000000000012', 3),
    -- ddd-in-go
    ('33333333-3333-3333-3333-000000000003', '22222222-2222-2222-2222-000000000020', 1),
    ('33333333-3333-3333-3333-000000000003', '22222222-2222-2222-2222-000000000021', 2),
    ('33333333-3333-3333-3333-000000000003', '22222222-2222-2222-2222-000000000022', 3),
    -- ts-fundamentals
    ('33333333-3333-3333-3333-000000000004', '22222222-2222-2222-2222-000000000026', 1),
    ('33333333-3333-3333-3333-000000000004', '22222222-2222-2222-2222-000000000027', 2),
    ('33333333-3333-3333-3333-000000000004', '22222222-2222-2222-2222-000000000028', 3),
    -- react-basics
    ('33333333-3333-3333-3333-000000000005', '22222222-2222-2222-2222-000000000046', 1),
    ('33333333-3333-3333-3333-000000000005', '22222222-2222-2222-2222-000000000047', 2),
    ('33333333-3333-3333-3333-000000000005', '22222222-2222-2222-2222-000000000048', 3),
    -- pg-mastery
    ('33333333-3333-3333-3333-000000000006', '22222222-2222-2222-2222-000000000061', 1),
    ('33333333-3333-3333-3333-000000000006', '22222222-2222-2222-2222-000000000062', 2),
    ('33333333-3333-3333-3333-000000000006', '22222222-2222-2222-2222-000000000063', 3),
    -- docker-basics
    ('33333333-3333-3333-3333-000000000007', '22222222-2222-2222-2222-000000000073', 1),
    ('33333333-3333-3333-3333-000000000007', '22222222-2222-2222-2222-000000000074', 2),
    ('33333333-3333-3333-3333-000000000007', '22222222-2222-2222-2222-000000000075', 3),
    -- nextjs-app-router
    ('33333333-3333-3333-3333-000000000008', '22222222-2222-2222-2222-000000000083', 1),
    ('33333333-3333-3333-3333-000000000008', '22222222-2222-2222-2222-000000000084', 2),
    ('33333333-3333-3333-3333-000000000008', '22222222-2222-2222-2222-000000000085', 3),
    -- aws-basics
    ('33333333-3333-3333-3333-000000000009', '22222222-2222-2222-2222-000000000093', 1),
    ('33333333-3333-3333-3333-000000000009', '22222222-2222-2222-2222-000000000094', 2),
    ('33333333-3333-3333-3333-000000000009', '22222222-2222-2222-2222-000000000095', 3),
    -- testing-strategy
    ('33333333-3333-3333-3333-000000000010', '22222222-2222-2222-2222-000000000098', 1),
    ('33333333-3333-3333-3333-000000000010', '22222222-2222-2222-2222-000000000099', 2),
    ('33333333-3333-3333-3333-000000000010', '22222222-2222-2222-2222-000000000100', 3);

-- ---- series_tags ----
INSERT INTO series_tags (series_id, tag_id) VALUES
    ('33333333-3333-3333-3333-000000000001', '11111111-1111-1111-1111-000000000001'), -- go-fundamentals x go
    ('33333333-3333-3333-3333-000000000002', '11111111-1111-1111-1111-000000000001'), -- go-advanced x go
    ('33333333-3333-3333-3333-000000000003', '11111111-1111-1111-1111-000000000001'), -- ddd-in-go x go
    ('33333333-3333-3333-3333-000000000003', '11111111-1111-1111-1111-000000000009'), -- ddd-in-go x ddd
    ('33333333-3333-3333-3333-000000000004', '11111111-1111-1111-1111-000000000002'), -- ts-fundamentals x typescript
    ('33333333-3333-3333-3333-000000000005', '11111111-1111-1111-1111-000000000003'), -- react-basics x react
    ('33333333-3333-3333-3333-000000000006', '11111111-1111-1111-1111-000000000005'), -- pg-mastery x postgres
    ('33333333-3333-3333-3333-000000000007', '11111111-1111-1111-1111-000000000004'), -- docker-basics x docker
    ('33333333-3333-3333-3333-000000000008', '11111111-1111-1111-1111-000000000006'), -- nextjs-app-router x nextjs
    ('33333333-3333-3333-3333-000000000008', '11111111-1111-1111-1111-000000000003'), -- nextjs-app-router x react
    ('33333333-3333-3333-3333-000000000009', '11111111-1111-1111-1111-000000000007'), -- aws-basics x aws
    ('33333333-3333-3333-3333-000000000010', '11111111-1111-1111-1111-000000000008'); -- testing-strategy x testing

COMMIT;
