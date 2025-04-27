package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxIface interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Storage struct {
	db PgxIface
}

func (s *Storage) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}

func New(db PgxIface) *Storage {
	return &Storage{db: db}
}

func Init(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	fmt.Println("Successfully connected to database!")
	return pool, nil
}
