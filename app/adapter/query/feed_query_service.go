package query

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kakkky/kakkky.dev/domain"
)

type FeedQueryService struct {
	db sqlx.ExtContext
}

func (qs *QueryService) NewFeedQueryService() domain.FeedQueryService {
	return &FeedQueryService{db: qs.db}
}

type feedItemRow struct {
	Kind         string         `db:"kind"`
	ID           string         `db:"id"`
	Slug         string         `db:"slug"`
	Title        string         `db:"title"`
	PublishedAt  time.Time      `db:"published_at"`
	TagIDs       pq.StringArray `db:"tag_ids"`
	ArticleCount int            `db:"article_count"` // article は 0
	SeriesStatus string         `db:"series_status"` // article は ''
}

func (fqs *FeedQueryService) ListFeedItems(
	ctx context.Context,
	afterID domain.FeedItemID,
	afterPublishedAt time.Time,
	limit int,
) ([]domain.FeedItem, error) {
	var publishedAtArg, idArg any
	if afterID != "" && !afterPublishedAt.IsZero() {
		publishedAtArg = afterPublishedAt
		idArg = string(afterID)
	}

	var rows []feedItemRow
	if err := sqlx.SelectContext(ctx, fqs.db, &rows, `
SELECT
  'article'::text AS kind,
  a.id::text      AS id,
  a.slug          AS slug,
  a.title         AS title,
  a.published_at  AS published_at,
  ARRAY(SELECT tag_id::text FROM article_tags WHERE article_id = a.id ORDER BY tag_id) AS tag_ids,
  0               AS article_count,
  ''              AS series_status
FROM articles a
WHERE a.status = 'published'
  AND NOT EXISTS (SELECT 1 FROM series_articles WHERE article_id = a.id)
  AND (a.published_at, a.id) < (
    COALESCE($1::timestamptz, 'infinity'),
    COALESCE($2::uuid, 'ffffffff-ffff-ffff-ffff-ffffffffffff')
  )

UNION ALL

SELECT
  'series'::text  AS kind,
  s.id::text      AS id,
  s.slug          AS slug,
  s.title         AS title,
  s.published_at  AS published_at,
  ARRAY(SELECT tag_id::text FROM series_tags WHERE series_id = s.id ORDER BY tag_id) AS tag_ids,
  (SELECT count(*) FROM series_articles WHERE series_id = s.id)::int AS article_count,
  s.status        AS series_status
FROM series s
WHERE s.status LIKE 'published_%'
  AND (s.published_at, s.id) < (
    COALESCE($1::timestamptz, 'infinity'),
    COALESCE($2::uuid, 'ffffffff-ffff-ffff-ffff-ffffffffffff')
  )

ORDER BY published_at DESC, id DESC
LIMIT $3
`, publishedAtArg, idArg, limit); err != nil {
		return nil, domain.ErrInternal.Wrap(err, "select feed items")
	}

	items := make([]domain.FeedItem, len(rows))
	for i, r := range rows {
		items[i] = r.toFeedItem()
	}
	return items, nil
}

func (r feedItemRow) toFeedItem() domain.FeedItem {
	tagIDs := make([]domain.TagID, len(r.TagIDs))
	for i, s := range r.TagIDs {
		tagIDs[i] = domain.TagID(s)
	}
	return domain.FeedItem{
		Kind:         domain.FeedItemKind(r.Kind),
		ID:           domain.FeedItemID(r.ID),
		Slug:         domain.Slug(r.Slug),
		Title:        r.Title,
		PublishedAt:  r.PublishedAt.UTC(),
		TagIDs:       tagIDs,
		ArticleCount: r.ArticleCount,
		SeriesStatus: domain.SeriesStatus(r.SeriesStatus),
	}
}
