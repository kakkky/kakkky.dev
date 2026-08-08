package domain

import "context"

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_repository.go -package=mock

type Repository interface {
	WithTx(ctx context.Context, fn func(Repository) error) error
	NewTagRepository() TagRepository
}

type TagRepository interface {
	FindByIDs(ctx context.Context, ids []TagID) ([]*Tag, error)
}
