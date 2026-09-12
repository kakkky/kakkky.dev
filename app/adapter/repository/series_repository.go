package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kakkky/kakkky.dev/domain"
)

type SeriesRepository struct {
	db sqlx.ExtContext
}

func (r *Repository) NewSeriesRepository() domain.SeriesRepository {
	return &SeriesRepository{db: r.db}
}

type seriesRow struct {
	ID          string         `db:"id"`
	Slug        string         `db:"slug"`
	Title       string         `db:"title"`
	Description string         `db:"description"`
	Status      string         `db:"status"`
	PublishedAt sql.NullTime   `db:"published_at"`
	CreatedAt   time.Time      `db:"created_at"`
	TagIDs      pq.StringArray `db:"tag_ids"`
	ArticleIDs  pq.StringArray `db:"article_ids"`
	Positions   pq.Int64Array  `db:"positions"`
}

func (r seriesRow) toSeries() *domain.Series {
	var publishedAt time.Time
	if r.PublishedAt.Valid {
		publishedAt = r.PublishedAt.Time.UTC()
	}
	tagIDs := make([]domain.TagID, len(r.TagIDs))
	for i, s := range r.TagIDs {
		tagIDs[i] = domain.TagID(s)
	}
	articles := make([]domain.SeriesArticle, len(r.ArticleIDs))
	for i, id := range r.ArticleIDs {
		articles[i] = domain.SeriesArticle{
			ArticleID: domain.ArticleID(id),
			Position:  int(r.Positions[i]),
		}
	}
	return &domain.Series{
		ID:          domain.SeriesID(r.ID),
		Slug:        domain.Slug(r.Slug),
		Title:       r.Title,
		Description: r.Description,
		Status:      domain.SeriesStatus(r.Status),
		PublishedAt: publishedAt,
		CreatedAt:   r.CreatedAt.UTC(),
		TagIDs:      tagIDs,
		Articles:    articles,
	}
}

func (sr *SeriesRepository) FindBySlug(ctx context.Context, slug domain.Slug) (*domain.Series, error) {
	var row seriesRow
	if err := sqlx.GetContext(ctx, sr.db, &row, `
SELECT s.id::text                 AS id,
       s.slug                     AS slug,
       s.title                    AS title,
       s.description              AS description,
       s.status                   AS status,
       s.published_at             AS published_at,
       s.created_at               AS created_at,
       ARRAY(SELECT tag_id::text     FROM series_tags     WHERE series_id = s.id ORDER BY tag_id)   AS tag_ids,
       ARRAY(SELECT article_id::text FROM series_articles WHERE series_id = s.id ORDER BY position) AS article_ids,
       ARRAY(SELECT position         FROM series_articles WHERE series_id = s.id ORDER BY position) AS positions
FROM series s
WHERE s.slug = $1
`, string(slug)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound.With("series not found")
		}
		return nil, domain.ErrInternal.Wrap(err, "find series by slug")
	}
	return row.toSeries(), nil
}

func (sr *SeriesRepository) FindByArticleID(ctx context.Context, articleID domain.ArticleID) (*domain.Series, error) {
	var row seriesRow
	if err := sqlx.GetContext(ctx, sr.db, &row, `
SELECT s.id::text                 AS id,
       s.slug                     AS slug,
       s.title                    AS title,
       s.description              AS description,
       s.status                   AS status,
       s.published_at             AS published_at,
       s.created_at               AS created_at,
       ARRAY(SELECT tag_id::text     FROM series_tags     WHERE series_id = s.id ORDER BY tag_id)   AS tag_ids,
       ARRAY(SELECT article_id::text FROM series_articles WHERE series_id = s.id ORDER BY position) AS article_ids,
       ARRAY(SELECT position         FROM series_articles WHERE series_id = s.id ORDER BY position) AS positions
FROM series s
JOIN series_articles sa ON sa.series_id = s.id
WHERE sa.article_id = $1
`, string(articleID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound.With("series not found")
		}
		return nil, domain.ErrInternal.Wrap(err, "find series by article id")
	}
	return row.toSeries(), nil
}

func (sr *SeriesRepository) Store(ctx context.Context, series *domain.Series) error {
	var id string
	if err := sqlx.GetContext(ctx, sr.db, &id, `
INSERT INTO series (slug, title, description, status, published_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id::text
`,
		string(series.Slug),
		series.Title,
		series.Description,
		string(series.Status),
		nullTime(series.PublishedAt),
	); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists.Wrap(err, "series slug already exists")
		}
		return domain.ErrInternal.Wrap(err, "insert series")
	}
	series.ID = domain.SeriesID(id)

	if len(series.TagIDs) > 0 {
		values := make([]string, len(series.TagIDs))
		args := make([]any, 0, 1+len(series.TagIDs))
		args = append(args, id)
		for i, tid := range series.TagIDs {
			values[i] = fmt.Sprintf("($1, $%d)", i+2)
			args = append(args, string(tid))
		}
		query := "INSERT INTO series_tags (series_id, tag_id) VALUES " + strings.Join(values, ", ")
		if _, err := sr.db.ExecContext(ctx, query, args...); err != nil {
			return domain.ErrInternal.Wrap(err, "insert series_tags")
		}
	}

	if len(series.Articles) > 0 {
		values := make([]string, len(series.Articles))
		args := make([]any, 0, 1+2*len(series.Articles))
		args = append(args, id)
		for i, sa := range series.Articles {
			values[i] = fmt.Sprintf("($1, $%d, $%d)", 2*i+2, 2*i+3)
			args = append(args, string(sa.ArticleID), sa.Position)
		}
		query := "INSERT INTO series_articles (series_id, article_id, position) VALUES " + strings.Join(values, ", ")
		if _, err := sr.db.ExecContext(ctx, query, args...); err != nil {
			return domain.ErrInternal.Wrap(err, "insert series_articles")
		}
	}

	return nil
}

func (sr *SeriesRepository) Update(ctx context.Context, series *domain.Series) error {
	res, err := sr.db.ExecContext(ctx, `
UPDATE series
SET title        = $2,
    description  = $3,
    status       = $4,
    published_at = $5,
    updated_at   = now()
WHERE id = $1
`,
		string(series.ID),
		series.Title,
		series.Description,
		string(series.Status),
		nullTime(series.PublishedAt),
	)
	if err != nil {
		return domain.ErrInternal.Wrap(err, "update series")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.ErrInternal.Wrap(err, "update series rows affected")
	}
	if n == 0 {
		return domain.ErrNotFound.With("series not found")
	}

	tagIDs := make([]string, len(series.TagIDs))
	for i, id := range series.TagIDs {
		tagIDs[i] = string(id)
	}
	if _, err := sr.db.ExecContext(ctx, `
INSERT INTO series_tags (series_id, tag_id)
SELECT $1, unnest($2::uuid[])
ON CONFLICT DO NOTHING
`, string(series.ID), pq.Array(tagIDs)); err != nil {
		return domain.ErrInternal.Wrap(err, "attach series_tags")
	}
	if _, err := sr.db.ExecContext(ctx, `
DELETE FROM series_tags
WHERE series_id = $1 AND tag_id <> ALL($2::uuid[])
`, string(series.ID), pq.Array(tagIDs)); err != nil {
		return domain.ErrInternal.Wrap(err, "detach series_tags")
	}
	return nil
}

func (sr *SeriesRepository) List(
	ctx context.Context,
	afterID domain.SeriesID,
	afterCreatedAt time.Time,
	limit int,
) ([]*domain.Series, error) {
	var createdAtArg, idArg any
	if afterID != "" && !afterCreatedAt.IsZero() {
		createdAtArg = afterCreatedAt
		idArg = string(afterID)
	}

	var rows []seriesRow
	if err := sqlx.SelectContext(ctx, sr.db, &rows, `
SELECT s.id::text                 AS id,
       s.slug                     AS slug,
       s.title                    AS title,
       s.description              AS description,
       s.status                   AS status,
       s.published_at             AS published_at,
       s.created_at               AS created_at,
       ARRAY(SELECT tag_id::text     FROM series_tags     WHERE series_id = s.id ORDER BY tag_id)   AS tag_ids,
       ARRAY(SELECT article_id::text FROM series_articles WHERE series_id = s.id ORDER BY position) AS article_ids,
       ARRAY(SELECT position         FROM series_articles WHERE series_id = s.id ORDER BY position) AS positions
FROM series s
WHERE (s.created_at, s.id) < (
  COALESCE($1::timestamptz, 'infinity'),
  COALESCE($2::uuid, 'ffffffff-ffff-ffff-ffff-ffffffffffff')
)
ORDER BY s.created_at DESC, s.id DESC
LIMIT $3
`, createdAtArg, idArg, limit); err != nil {
		return nil, domain.ErrInternal.Wrap(err, "list series")
	}

	series := make([]*domain.Series, len(rows))
	for i, r := range rows {
		series[i] = r.toSeries()
	}
	return series, nil
}
