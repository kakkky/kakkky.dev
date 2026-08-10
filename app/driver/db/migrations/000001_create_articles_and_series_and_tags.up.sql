-- series
CREATE TABLE series (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         VARCHAR(20)  NOT NULL UNIQUE,
    title        VARCHAR(100) NOT NULL,
    description  TEXT,
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft'
                 CHECK (status IN ('draft', 'published_ongoing', 'published_completed')),
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX series_published_recent_idx
    ON series (published_at DESC)
    WHERE status LIKE 'published_%';

-- articles
CREATE TABLE articles (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          VARCHAR(20)  NOT NULL UNIQUE,
    title         VARCHAR(100) NOT NULL,
    body          TEXT         NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'draft'
                  CHECK (status IN ('draft', 'published')),
    published_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX articles_published_recent_idx
    ON articles (published_at DESC)
    WHERE status = 'published';

-- series_articles: 連載への所属 (1 記事 = 最大 1 シリーズ)
CREATE TABLE series_articles (
    series_id   UUID         NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    article_id  UUID         PRIMARY KEY REFERENCES articles(id) ON DELETE CASCADE,
    position    INTEGER      NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (series_id, position)
);

-- tags
CREATE TABLE tags (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(20)  NOT NULL UNIQUE,
    name        VARCHAR(30)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- article_tags: 単発記事のタグ (多対多)
CREATE TABLE article_tags (
    article_id  UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id      UUID         NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (article_id, tag_id)
);

CREATE INDEX article_tags_tag_id_idx ON article_tags (tag_id);

-- series_tags: 連載のタグ (多対多)
CREATE TABLE series_tags (
    series_id   UUID         NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    tag_id      UUID         NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (series_id, tag_id)
);

CREATE INDEX series_tags_tag_id_idx ON series_tags (tag_id);
