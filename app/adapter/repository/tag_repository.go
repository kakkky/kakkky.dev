package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/kakkky/kakkky.dev/domain"
)

type TagRepository struct {
	db sqlx.ExtContext
}

func (r *Repository) NewTagRepository() domain.TagRepository {
	return &TagRepository{db: r.db}
}

type tagRow struct {
	ID   string `db:"id"`
	Slug string `db:"slug"`
	Name string `db:"name"`
}

func (tr *TagRepository) ListAll(ctx context.Context) ([]*domain.Tag, error) {
	var rows []tagRow
	if err := sqlx.SelectContext(ctx, tr.db, &rows, `
SELECT id::text AS id,
       slug,
       name
FROM tags
ORDER BY name
`); err != nil {
		return nil, domain.ErrInternal.Wrap(err, "list all tags")
	}

	tags := make([]*domain.Tag, len(rows))
	for i, r := range rows {
		tags[i] = &domain.Tag{
			ID:   domain.TagID(r.ID),
			Slug: domain.Slug(r.Slug),
			Name: r.Name,
		}
	}
	return tags, nil
}
