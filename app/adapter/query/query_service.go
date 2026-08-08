package query

import "github.com/jmoiron/sqlx"

type QueryService struct {
	db sqlx.ExtContext
}

func NewQueryService(db *sqlx.DB) *QueryService {
	return &QueryService{db: db}
}
