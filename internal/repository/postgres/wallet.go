package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"wallet/internal/model/wallet"
)

func (s *Storage) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	query := `
		SELECT balance 
		FROM wallets 
		WHERE id = $1
	`

	var balance int64

	err := s.db.QueryRow(ctx, query, walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, wallet.ErrWalletNotFound
		}
		return 0, err
	}

	return balance, nil
}

func (s *Storage) Deposit(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount int64) (int64, error) {
	return s.updateBalance(ctx, tx, walletID, amount)
}

func (s *Storage) Withdraw(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount int64) (int64, error) {
	return s.updateBalance(ctx, tx, walletID, -amount)
}

func (s *Storage) updateBalance(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, delta int64) (int64, error) {
	query := `
		UPDATE wallets
		SET balance = balance + $1
		WHERE id = $2
		RETURNING balance;
		`

	var balance int64
	err := tx.QueryRow(ctx, query, delta, walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, wallet.ErrWalletNotFound
		}
		return 0, err
	}

	return balance, nil
}
