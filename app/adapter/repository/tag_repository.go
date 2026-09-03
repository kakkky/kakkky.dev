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

func (r tagRow) toTag() *domain.Tag {
	return &domain.Tag{
		ID:   domain.TagID(r.ID),
		Slug: domain.Slug(r.Slug),
		Name: r.Name,
	}
}

func (tr *TagRepository) List(ctx context.Context) ([]*domain.Tag, error) {
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
		tags[i] = r.toTag()
	}
	return tags, nil
}

func (tr *TagRepository) Store(ctx context.Context, tag *domain.Tag) error {
	if tag.ID != "" {
		return domain.ErrInternal.With("tag update is not implemented yet")
	}
	var id string
	if err := sqlx.GetContext(ctx, tr.db, &id, `
INSERT INTO tags (slug, name)
VALUES ($1, $2)
RETURNING id::text
`, string(tag.Slug), tag.Name); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists.Wrap(err, "tag slug already exists")
		}
		return domain.ErrInternal.Wrap(err, "insert tag")
	}
	tag.ID = domain.TagID(id)
	return nil
}

func (tr *TagRepository) FindByIDs(ctx context.Context, ids ...domain.TagID) ([]*domain.Tag, error) {
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
ORDER BY name
`, pq.Array(strIDs)); err != nil {
		return nil, domain.ErrInternal.Wrap(err, "find tags by ids")
	}

	tags := make([]*domain.Tag, len(rows))
	for i, r := range rows {
		tags[i] = r.toTag()
	}
	return tags, nil
}
