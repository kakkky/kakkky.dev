package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kakkky/kakkky.dev/domain"
)

type ArticleRepository struct {
	db sqlx.ExtContext
}

func (r *Repository) NewArticleRepository() domain.ArticleRepository {
	return &ArticleRepository{db: r.db}
}

type articleRow struct {
	ID          string         `db:"id"`
	Slug        string         `db:"slug"`
	Title       string         `db:"title"`
	Body        string         `db:"body"`
	Status      string         `db:"status"`
	PublishedAt sql.NullTime   `db:"published_at"`
	CreatedAt   time.Time      `db:"created_at"`
	TagIDs      pq.StringArray `db:"tag_ids"`
}

func (r articleRow) toArticle() *domain.Article {
	tagIDs := make([]domain.TagID, len(r.TagIDs))
	for i, s := range r.TagIDs {
		tagIDs[i] = domain.TagID(s)
	}
	var publishedAt time.Time
	if r.PublishedAt.Valid {
		publishedAt = r.PublishedAt.Time.UTC()
	}
	return &domain.Article{
		ID:          domain.ArticleID(r.ID),
		Slug:        domain.Slug(r.Slug),
		Title:       r.Title,
		Body:        r.Body,
		Status:      domain.ArticleStatus(r.Status),
		PublishedAt: publishedAt,
		CreatedAt:   r.CreatedAt.UTC(),
		TagIDs:      tagIDs,
	}
}

func (ar *ArticleRepository) FindBySlug(ctx context.Context, slug domain.Slug) (*domain.Article, error) {
	var row articleRow
	if err := sqlx.GetContext(ctx, ar.db, &row, `
SELECT a.id::text     AS id,
       a.slug         AS slug,
       a.title        AS title,
       a.body         AS body,
       a.status       AS status,
       a.published_at AS published_at,
       a.created_at   AS created_at,
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

func (ar *ArticleRepository) FindByIDs(ctx context.Context, ids ...domain.ArticleID) ([]*domain.Article, error) {
	if len(ids) == 0 {
		return []*domain.Article{}, nil
	}

	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = string(id)
	}

	var rows []articleRow
	if err := sqlx.SelectContext(ctx, ar.db, &rows, `
SELECT a.id::text     AS id,
       a.slug         AS slug,
       a.title        AS title,
       a.body         AS body,
       a.status       AS status,
       a.published_at AS published_at,
       a.created_at   AS created_at,
       ARRAY(SELECT tag_id::text FROM article_tags WHERE article_id = a.id ORDER BY tag_id) AS tag_ids
FROM articles a
WHERE a.id = ANY($1::uuid[])
`, pq.Array(strIDs)); err != nil {
		return nil, domain.ErrInternal.Wrap(err, "find articles by ids")
	}

	articles := make([]*domain.Article, len(rows))
	for i, r := range rows {
		articles[i] = r.toArticle()
	}
	return articles, nil
}

func (ar *ArticleRepository) List(
	ctx context.Context,
	afterID domain.ArticleID,
	afterCreatedAt time.Time,
	limit int,
) ([]*domain.Article, error) {
	var createdAtArg, idArg any
	if afterID != "" && !afterCreatedAt.IsZero() {
		createdAtArg = afterCreatedAt
		idArg = string(afterID)
	}

	var rows []articleRow
	if err := sqlx.SelectContext(ctx, ar.db, &rows, `
SELECT a.id::text     AS id,
       a.slug         AS slug,
       a.title        AS title,
       a.body         AS body,
       a.status       AS status,
       a.published_at AS published_at,
       a.created_at   AS created_at,
       ARRAY(SELECT tag_id::text FROM article_tags WHERE article_id = a.id ORDER BY tag_id) AS tag_ids
FROM articles a
WHERE (a.created_at, a.id) < (
  COALESCE($1::timestamptz, 'infinity'),
  COALESCE($2::uuid, 'ffffffff-ffff-ffff-ffff-ffffffffffff')
)
ORDER BY a.created_at DESC, a.id DESC
LIMIT $3
`, createdAtArg, idArg, limit); err != nil {
		return nil, domain.ErrInternal.Wrap(err, "list articles")
	}

	articles := make([]*domain.Article, len(rows))
	for i, r := range rows {
		articles[i] = r.toArticle()
	}
	return articles, nil
}
