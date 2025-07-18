package database

import (
	"context"
	"database/sql"
)

type DBInterface interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{})
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Postgres struct {
	db *sql.DB
}

func Newpostgres(db *sql.DB) *Postgres {
	return &Postgres{db}
}
