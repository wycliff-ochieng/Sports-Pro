package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/wycliff-ochieng/internal/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type DBInterface interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	BeginTx(ctx context.Context, opt *sql.TxOptions) (*sql.Tx, error)
}

type PostgresDB struct {
	db  *sql.DB
	cfg *config.Config
}

var migrationsFS embed.FS

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {

	//dsn := "user=" + cfg.DBUser + "password=" + cfg.DBPassword + "host=" + cfg.DBHost + "port=" + strconv.Itoa(cfg.DBPort) + "dbname=" + cfg.DBName + "sslmode= " + cfg.DBsslmode
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

	if err := goose.Up(db, "internal/database/migrations"); err != nil {
		log.Fatalf("error spinning up goose:%v", err)
	}

	log.Println(">>>>migrations run successfully...")

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

func (p *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}
func (p *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

func (p *PostgresDB) BeginTx(ctx context.Context, opt *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opt)
}
