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
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'go-tips-' || lpad(n::text, 3, '0'),
    'Go Tips #' || lpad(n::text, 3, '0'),
    E'# Go Tips #' || lpad(n::text, 3, '0') || E'\n\nGo に関する Tips 記事 (No.' || n || E')。\n\n```go\nfunc main() { fmt.Println("hello, go") }\n```\n',
    'Go に関する ' || n || ' 本目の Tips 記事。',
    CASE WHEN n = ANY(ARRAY[5, 17]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[5, 17]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(1, 25) n;

-- TypeScript: 026-045
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'ts-tips-' || lpad(n::text, 3, '0'),
    'TypeScript Tips #' || lpad(n::text, 3, '0'),
    E'# TypeScript Tips #' || lpad(n::text, 3, '0') || E'\n\nTypeScript に関する Tips 記事 (No.' || n || E')。\n\n```ts\nfunction pick<T, K extends keyof T>(o: T, k: K[]): Pick<T, K> { return {} as any }\n```\n',
    'TypeScript に関する ' || n || ' 本目の Tips 記事。',
    CASE WHEN n = ANY(ARRAY[33, 42]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[33, 42]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(26, 45) n;

-- React: 046-060
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'react-tips-' || lpad(n::text, 3, '0'),
    'React Tips #' || lpad(n::text, 3, '0'),
    E'# React Tips #' || lpad(n::text, 3, '0') || E'\n\nReact に関する Tips 記事 (No.' || n || E')。\n\n```tsx\nexport function Hello() { return <p>hi</p> }\n```\n',
    'React に関する ' || n || ' 本目の Tips 記事。',
    CASE WHEN n = ANY(ARRAY[55]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[55]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(46, 60) n;

-- PostgreSQL: 061-072
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'pg-tips-' || lpad(n::text, 3, '0'),
    'PostgreSQL Tips #' || lpad(n::text, 3, '0'),
    E'# PostgreSQL Tips #' || lpad(n::text, 3, '0') || E'\n\nPostgreSQL に関する Tips 記事 (No.' || n || E')。\n\n```sql\nSELECT count(*) FROM articles WHERE status = ''published'';\n```\n',
    'PostgreSQL に関する ' || n || ' 本目の Tips 記事。',
    CASE WHEN n = ANY(ARRAY[66]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[66]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(61, 72) n;

-- Docker: 073-082
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'docker-tips-' || lpad(n::text, 3, '0'),
    'Docker Tips #' || lpad(n::text, 3, '0'),
    E'# Docker Tips #' || lpad(n::text, 3, '0') || E'\n\nDocker に関する Tips 記事 (No.' || n || E')。\n\n```yaml\nservices:\n  app:\n    build: .\n```\n',
    'Docker に関する ' || n || ' 本目の Tips 記事。',
    CASE WHEN n = ANY(ARRAY[78]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[78]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(73, 82) n;

-- Next.js: 083-092
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'nextjs-tips-' || lpad(n::text, 3, '0'),
    'Next.js Tips #' || lpad(n::text, 3, '0'),
    E'# Next.js Tips #' || lpad(n::text, 3, '0') || E'\n\nNext.js に関する Tips 記事 (No.' || n || E')。\n\n```tsx\nexport default function Page() { return <main>hi</main> }\n```\n',
    'Next.js に関する ' || n || ' 本目の Tips 記事。',
    CASE WHEN n = ANY(ARRAY[89]) THEN 'draft' ELSE 'published' END,
    CASE WHEN n = ANY(ARRAY[89]) THEN NULL
         ELSE ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval) END
FROM generate_series(83, 92) n;

-- AWS: 093-097
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'aws-tips-' || lpad(n::text, 3, '0'),
    'AWS Tips #' || lpad(n::text, 3, '0'),
    E'# AWS Tips #' || lpad(n::text, 3, '0') || E'\n\nAWS に関する Tips 記事 (No.' || n || E')。\n',
    'AWS に関する ' || n || ' 本目の Tips 記事。',
    'published',
    ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval)
FROM generate_series(93, 97) n;

-- Testing: 098-100
INSERT INTO articles (id, slug, title, body, summary, status, published_at)
SELECT
    ('22222222-2222-2222-2222-' || lpad(n::text, 12, '0'))::uuid,
    'testing-tips-' || lpad(n::text, 3, '0'),
    'Testing Tips #' || lpad(n::text, 3, '0'),
    E'# Testing Tips #' || lpad(n::text, 3, '0') || E'\n\nテストに関する Tips 記事 (No.' || n || E')。\n',
    'テストに関する ' || n || ' 本目の Tips 記事。',
    'published',
    ('2024-01-01 10:00:00+09'::timestamptz + ((n - 1) * 8 || ' days')::interval)
FROM generate_series(98, 100) n;

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
