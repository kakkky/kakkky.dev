package db

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kakkky/kakkky.dev/config"
)

func NewDB(ctx context.Context, cfg *config.Config) (db *sqlx.DB, cleanup func(), err error) {
	db, err = sqlx.ConnectContext(ctx, "pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("sqlx.ConnectContext: %w", err)
	}
	return db,
		func() { _ = db.Close() },
		nil
}
