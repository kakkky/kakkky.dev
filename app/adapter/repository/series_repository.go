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
