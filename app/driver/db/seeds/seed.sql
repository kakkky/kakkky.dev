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

-- ---- tags ----
INSERT INTO tags (id, slug, name) VALUES
    ('11111111-1111-1111-1111-000000000001', 'go',         'Go'),
    ('11111111-1111-1111-1111-000000000002', 'typescript', 'TypeScript'),
    ('11111111-1111-1111-1111-000000000003', 'react',      'React'),
    ('11111111-1111-1111-1111-000000000004', 'docker',     'Docker'),
    ('11111111-1111-1111-1111-000000000005', 'postgres',   'PostgreSQL'),
    ('11111111-1111-1111-1111-000000000006', 'nextjs',     'Next.js');

-- ---- articles ----
INSERT INTO articles (id, slug, title, body, summary, status, published_at) VALUES
    (
        '22222222-2222-2222-2222-000000000001',
        'go-context-basics',
        'Go の context 入門',
        E'# Go の context 入門\n\ncontext はゴルーチン間でキャンセルや期限を伝播させるための仕組みだ。\n\n## 使い方\n\n`context.Background()` から派生させ、`WithCancel` / `WithTimeout` などでキャンセル可能な子 context を作る。\n\n```go\nctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)\ndefer cancel()\n```\n',
        'Go の context パッケージの基本と、キャンセル・タイムアウトの伝播方法をまとめる。',
        'published',
        '2026-01-15 10:00:00+09'
    ),
    (
        '22222222-2222-2222-2222-000000000002',
        'go-error-handling',
        'Go のエラーハンドリング再考',
        E'# Go のエラーハンドリング再考\n\n`errors.Is` / `errors.As` の使い分けと、独自エラー型を定義する際の指針を整理する。\n\n## errors.Is と errors.As\n\n- `errors.Is`: センチネルエラーとの一致判定\n- `errors.As`: エラー型の取り出し\n',
        'errors.Is / errors.As の使い分けと独自エラー型の設計指針。',
        'published',
        '2026-02-10 10:00:00+09'
    ),
    (
        '22222222-2222-2222-2222-000000000003',
        'typescript-generics',
        'TypeScript のジェネリクス実践',
        E'# TypeScript のジェネリクス実践\n\n型パラメータの制約 (`extends`) と条件付き型を組み合わせると、API レスポンス型の推論を強化できる。\n\n```ts\nfunction pick<T, K extends keyof T>(obj: T, keys: K[]): Pick<T, K> { ... }\n```\n',
        '型パラメータの制約と条件付き型を使った実践的なパターン集。',
        'published',
        '2026-03-05 10:00:00+09'
    ),
    (
        '22222222-2222-2222-2222-000000000004',
        'postgres-index-tips',
        'PostgreSQL の Index 設計 Tips',
        E'# PostgreSQL の Index 設計 Tips\n\n部分 Index (`WHERE status = ''published''`) を使うと、対象行を絞り込んだ上での高速化ができる。\n\n## 部分 Index の例\n\n```sql\nCREATE INDEX articles_published_recent_idx\n  ON articles (published_at DESC)\n  WHERE status = ''published'';\n```\n',
        '部分 Index / 複合 Index / カバリング Index の使いどころ。',
        'published',
        '2026-04-20 10:00:00+09'
    ),
    (
        '22222222-2222-2222-2222-000000000005',
        'docker-compose-dev',
        'Docker Compose で開発環境を作る',
        E'# Docker Compose で開発環境を作る\n\nアプリと DB を 1 つの `compose.yml` にまとめ、ボリュームで Go module / ビルドキャッシュを永続化する。\n',
        'アプリと DB をまとめた開発用 compose.yml のポイント。',
        'published',
        '2026-05-12 10:00:00+09'
    ),
    (
        '22222222-2222-2222-2222-000000000006',
        'nextjs-app-router',
        'Next.js App Router のメモ',
        E'# Next.js App Router のメモ\n\nServer Components と Client Components の境界設計と、`use client` の付け方の指針。\n',
        NULL,
        'draft',
        NULL
    );

-- ---- article_tags ----
INSERT INTO article_tags (article_id, tag_id) VALUES
    ('22222222-2222-2222-2222-000000000001', '11111111-1111-1111-1111-000000000001'), -- go-context-basics x go
    ('22222222-2222-2222-2222-000000000002', '11111111-1111-1111-1111-000000000001'), -- go-error-handling x go
    ('22222222-2222-2222-2222-000000000003', '11111111-1111-1111-1111-000000000002'), -- typescript-generics x typescript
    ('22222222-2222-2222-2222-000000000004', '11111111-1111-1111-1111-000000000005'), -- postgres-index-tips x postgres
    ('22222222-2222-2222-2222-000000000005', '11111111-1111-1111-1111-000000000004'), -- docker-compose-dev x docker
    ('22222222-2222-2222-2222-000000000006', '11111111-1111-1111-1111-000000000006'), -- nextjs-app-router x nextjs
    ('22222222-2222-2222-2222-000000000006', '11111111-1111-1111-1111-000000000003'); -- nextjs-app-router x react

-- ---- series ----
INSERT INTO series (id, slug, title, description, status) VALUES
    (
        '33333333-3333-3333-3333-000000000001',
        'go-fundamentals',
        'Go 基礎講座',
        'Go を書く上で押さえておきたい基本トピックをまとめる連載。',
        'completed'
    ),
    (
        '33333333-3333-3333-3333-000000000002',
        'web-dev-env',
        'Web 開発環境構築',
        'コンテナ・ミドルウェアを組み合わせた開発環境作りの記録。',
        'ongoing'
    );

-- ---- series_articles ----
-- series_articles.PRIMARY KEY は article_id なので 1 記事は 1 series
INSERT INTO series_articles (series_id, article_id, position) VALUES
    ('33333333-3333-3333-3333-000000000001', '22222222-2222-2222-2222-000000000001', 1), -- go-fundamentals / go-context-basics
    ('33333333-3333-3333-3333-000000000001', '22222222-2222-2222-2222-000000000002', 2), -- go-fundamentals / go-error-handling
    ('33333333-3333-3333-3333-000000000002', '22222222-2222-2222-2222-000000000005', 1); -- web-dev-env / docker-compose-dev

-- ---- series_tags ----
INSERT INTO series_tags (series_id, tag_id) VALUES
    ('33333333-3333-3333-3333-000000000001', '11111111-1111-1111-1111-000000000001'), -- go-fundamentals x go
    ('33333333-3333-3333-3333-000000000002', '11111111-1111-1111-1111-000000000004'); -- web-dev-env x docker

COMMIT;
