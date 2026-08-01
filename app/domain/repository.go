package domain

import "context"

type Repository interface {
	WithTx(ctx context.Context, fn func(Repository) error) error
}
