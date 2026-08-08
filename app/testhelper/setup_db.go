package testhelper

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/kakkky/kakkky.dev/driver/db/schema"
)

func SetupDB(ctx context.Context) (*sqlx.DB, func(), error) {
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("run postgres container: %w", err)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, fmt.Errorf("connection string: %w", err)
	}

	sqlxDB, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %w", err)
	}

	if _, err := sqlxDB.ExecContext(ctx, schema.SQL); err != nil {
		return nil, nil, fmt.Errorf("apply schema: %w", err)
	}

	cleanup := func() {
		_ = sqlxDB.Close()
		_ = pg.Terminate(ctx)
	}
	return sqlxDB, cleanup, nil
}
