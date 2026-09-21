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
	Count(ctx context.Context) (int, error)
	Store(ctx context.Context, article *Article) error
	Update(ctx context.Context, article *Article) error
	Delete(ctx context.Context, id ArticleID) error
}

type SeriesRepository interface {
	FindBySlug(ctx context.Context, slug Slug) (*Series, error)
	FindByArticleID(ctx context.Context, articleID ArticleID) (*Series, error)
	FindByArticleIDs(ctx context.Context, ids ...ArticleID) ([]*Series, error)
	List(ctx context.Context, afterID SeriesID, afterCreatedAt time.Time, limit int) ([]*Series, error)
	Count(ctx context.Context) (int, error)
	Store(ctx context.Context, series *Series) error
	Update(ctx context.Context, series *Series) error
}

type TagRepository interface {
	List(ctx context.Context) ([]*Tag, error)
	FindByIDs(ctx context.Context, ids ...TagID) ([]*Tag, error)
	Store(ctx context.Context, tag *Tag) error
}
