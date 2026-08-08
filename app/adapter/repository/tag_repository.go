package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

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

func (tr *TagRepository) FindByIDs(ctx context.Context, ids []domain.TagID) ([]*domain.Tag, error) {
	if len(ids) == 0 {
		return []*domain.Tag{}, nil
	}

	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = string(id)
	}

	var rows []tagRow
	if err := sqlx.SelectContext(ctx, tr.db, &rows, `
SELECT id::text AS id,
       slug,
       name
FROM tags
WHERE id = ANY($1::uuid[])
`, pq.Array(strIDs)); err != nil {
		return nil, domain.ErrInternal.Wrap(err, "select tags by ids")
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
