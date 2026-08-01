package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kakkky/kakkky.dev/domain"
)

type Repository struct {
	db sqlx.ExtContext
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(ctx context.Context, fn func(domain.Repository) error) (err error) {
	db, ok := r.db.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("WithTx: nested transaction not supported")
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("BeginTxx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	err = fn(&Repository{db: tx})
	return err
}
