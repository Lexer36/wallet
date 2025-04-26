package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func (s *Storage) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}

func New(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

func Init(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
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
