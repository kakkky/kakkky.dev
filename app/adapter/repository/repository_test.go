package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/kakkky/kakkky.dev/testhelper"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	db, cleanup, err := testhelper.SetupDB(ctx)
	if err != nil {
		panic(err)
	}
	testDB = db

	code := m.Run()
	cleanup()
	os.Exit(code)
}
