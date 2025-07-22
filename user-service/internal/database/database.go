package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/wycliff-ochieng/internal/config"
)

type DBInterface interface {
	//QueryRowContext(ctx context.Context, query string, args ...interface{})
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Postgres struct {
	db  *sql.DB
	cfg *config.Config
}

func Newpostgres(cfg *config.Config) (*Postgres, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBsslmode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	//defer db.Close()
	//set goose dialect
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("error setting dialect:%v", err)
	}

	if err := goose.Up(db, "./internal/database/migrations"); err != nil {
		log.Fatalf("error spinning up goose:%v", err)
	}

	log.Println(">>>>migrations run successfully...")
	return &Postgres{db: db}, nil
}

//func (p *Postgres) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
//	return p.db.QueryRowContext(ctx, query, args...)
//}

func (p *Postgres) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}
