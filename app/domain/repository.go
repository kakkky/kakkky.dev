package domain

import (
	"context"
	"time"
)

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_repository.go -package=mock

type Repository interface {
	WithTx(ctx context.Context, fn func(Repository) error) error
	NewTagRepository() TagRepository
	NewArticleRepository() ArticleRepository
	NewSeriesRepository() SeriesRepository
}

type ArticleRepository interface {
	FindBySlug(ctx context.Context, slug Slug) (*Article, error)
	FindByIDs(ctx context.Context, ids ...ArticleID) ([]*Article, error)
	List(ctx context.Context, afterID ArticleID, afterCreatedAt time.Time, limit int) ([]*Article, error)
	Create(ctx context.Context, article *Article) error
	Update(ctx context.Context, article *Article) error
}

type SeriesRepository interface {
	FindBySlug(ctx context.Context, slug Slug) (*Series, error)
	List(ctx context.Context, afterID SeriesID, afterCreatedAt time.Time, limit int) ([]*Series, error)
}

type TagRepository interface {
	ListAll(ctx context.Context) ([]*Tag, error)
	FindByIDs(ctx context.Context, ids ...TagID) ([]*Tag, error)
	FindByNames(ctx context.Context, names ...string) ([]*Tag, error)
	Create(ctx context.Context, tag *Tag) error
}
