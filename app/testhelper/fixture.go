package testhelper

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/kakkky/kakkky.dev/domain"
)

type Fixtures struct {
	Tags     []*domain.Tag
	Articles []*domain.Article
	Series   []*domain.Series
}

func Insert(t *testing.T, ctx context.Context, db sqlx.ExtContext, f Fixtures) {
	t.Helper()

	for _, tag := range f.Tags {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO tags (id, slug, name) VALUES ($1, $2, $3)`,
			string(tag.ID), string(tag.Slug), tag.Name,
		); err != nil {
			t.Fatalf("insert tag: %v", err)
		}
	}

	for _, a := range f.Articles {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO articles (id, slug, title, body, status, published_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, now()), COALESCE($8, now()))`,
			string(a.ID), string(a.Slug), a.Title, a.Body, string(a.Status), a.PublishedAt,
			nullTime(a.CreatedAt),
			nullTime(a.UpdatedAt),
		); err != nil {
			t.Fatalf("insert article: %v", err)
		}
		for _, tid := range a.TagIDs {
			if _, err := db.ExecContext(ctx,
				`INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2)`,
				string(a.ID), string(tid),
			); err != nil {
				t.Fatalf("insert article_tag: %v", err)
			}
		}
	}

	for _, s := range f.Series {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO series (id, slug, title, description, status, published_at, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, now()))`,
			string(s.ID), string(s.Slug), s.Title, s.Description, string(s.Status), s.PublishedAt,
			nullTime(s.CreatedAt),
		); err != nil {
			t.Fatalf("insert series: %v", err)
		}
		for _, sa := range s.Articles {
			if _, err := db.ExecContext(ctx,
				`INSERT INTO series_articles (series_id, article_id, position) VALUES ($1, $2, $3)`,
				string(s.ID), string(sa.ArticleID), sa.Position,
			); err != nil {
				t.Fatalf("insert series_article: %v", err)
			}
		}
		for _, tid := range s.TagIDs {
			if _, err := db.ExecContext(ctx,
				`INSERT INTO series_tags (series_id, tag_id) VALUES ($1, $2)`,
				string(s.ID), string(tid),
			); err != nil {
				t.Fatalf("insert series_tag: %v", err)
			}
		}
	}
}

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func TruncateAll(t *testing.T, ctx context.Context, db sqlx.ExecerContext) {
	t.Helper()
	if _, err := db.ExecContext(ctx,
		`TRUNCATE articles, series, tags, article_tags, series_tags, series_articles RESTART IDENTITY CASCADE`,
	); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
