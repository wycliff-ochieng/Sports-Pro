package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/pressly/goose/v3"
)

type DBInterface interface {
	QueryRowContext(ctx context.Context)
}

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(db *sql.DB) (*PostgresDB, error) {

	dsn := "user=postgres password=password host=localhost port=5432 dbname=User sslmode=require"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	//defer db.Close()
	//set goose dialect
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("error setting dialect:%v", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("error spinning up goose:%v", err)
	}

	log.Println(">>>>migrations run successfully...")

	return &PostgresDB{db: db}, nil
}
