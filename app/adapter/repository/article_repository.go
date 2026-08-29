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
	UpdatedAt   time.Time      `db:"updated_at"`
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
		UpdatedAt:   r.UpdatedAt.UTC(),
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
       a.updated_at   AS updated_at,
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
       a.updated_at   AS updated_at,
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

func (ar *ArticleRepository) Create(ctx context.Context, article *domain.Article) error {
	var publishedAt sql.NullTime
	if !article.PublishedAt.IsZero() {
		publishedAt = sql.NullTime{Time: article.PublishedAt, Valid: true}
	}

	var row struct {
		ID        string    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}
	if err := sqlx.GetContext(ctx, ar.db, &row, `
INSERT INTO articles (slug, title, body, status, published_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id::text AS id, created_at
`, string(article.Slug), article.Title, article.Body, string(article.Status), publishedAt); err != nil {
		return domain.ErrInternal.Wrap(err, "create article")
	}
	article.ID = domain.ArticleID(row.ID)
	article.CreatedAt = row.CreatedAt.UTC()

	if err := insertArticleTags(ctx, ar.db, article.ID, article.TagIDs); err != nil {
		return err
	}
	return nil
}

func (ar *ArticleRepository) Update(ctx context.Context, article *domain.Article) error {
	var publishedAt sql.NullTime
	if !article.PublishedAt.IsZero() {
		publishedAt = sql.NullTime{Time: article.PublishedAt, Valid: true}
	}

	res, err := ar.db.ExecContext(ctx, `
UPDATE articles
SET slug = $1, title = $2, body = $3, status = $4, published_at = $5, updated_at = now()
WHERE id = $6
`, string(article.Slug), article.Title, article.Body, string(article.Status), publishedAt, string(article.ID))
	if err != nil {
		return domain.ErrInternal.Wrap(err, "update article")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.ErrInternal.Wrap(err, "update article rows affected")
	}
	if n == 0 {
		return domain.ErrNotFound.With("article not found")
	}

	if _, err := ar.db.ExecContext(ctx, `DELETE FROM article_tags WHERE article_id = $1`, string(article.ID)); err != nil {
		return domain.ErrInternal.Wrap(err, "delete article tags")
	}
	if err := insertArticleTags(ctx, ar.db, article.ID, article.TagIDs); err != nil {
		return err
	}
	return nil
}

func insertArticleTags(ctx context.Context, db sqlx.ExtContext, articleID domain.ArticleID, tagIDs []domain.TagID) error {
	for _, tid := range tagIDs {
		if _, err := db.ExecContext(ctx, `
INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2)
`, string(articleID), string(tid)); err != nil {
			return domain.ErrInternal.Wrap(err, "insert article_tag")
		}
	}
	return nil
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
       a.updated_at   AS updated_at,
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
