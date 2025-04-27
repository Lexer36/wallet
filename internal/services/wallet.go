package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"wallet/internal/model/wallet"
)

type WalletStorage interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)                          // BeginTx starts a new database transaction
	Deposit(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount int64) (int64, error)  // Deposit returns updated balance
	Withdraw(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount int64) (int64, error) // Withdraw returns updated balance
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)                        // GetBalance returns balance
}

type WalletCache interface {
	Get(ctx context.Context, key string) (int64, bool)
	Set(ctx context.Context, key string, balance int64)
	Delete(ctx context.Context, key string)
}

type WalletService struct {
	log   *slog.Logger
	repo  WalletStorage
	cache WalletCache
}

func NewWalletService(repo WalletStorage, cache WalletCache, log *slog.Logger) *WalletService {
	return &WalletService{
		repo:  repo,
		cache: cache,
		log:   log,
	}
}

func (ws *WalletService) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error {
	tx, err := ws.repo.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		ws.log.Error("Error starting transaction", "walletID", walletID, "amount", amount, "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	balance, err := ws.repo.Deposit(ctx, tx, walletID, amount)
	if err != nil {
		ws.log.Error("Error during deposit", "walletID", walletID, "amount", amount, "error", err)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		ws.log.Error("Error committing transaction", "walletID", walletID, "amount", amount, "error", err)
		return err
	}

	ws.cache.Set(ctx, walletID.String(), balance)

	return nil
}

func (ws *WalletService) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	tx, err := ws.repo.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		ws.log.Error("Error starting transaction", "walletID", walletID, "amount", amount, "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	balance, err := ws.repo.Withdraw(ctx, tx, walletID, amount)
	if err != nil {
		ws.log.Error("Error during withdrawal", "walletID", walletID, "amount", amount, "error", err)
		return err
	}

	if balance < 0 {
		ws.log.Error("Insufficient funds", "walletID", walletID, "amount", amount, "balance", balance)
		return wallet.ErrNotEnoughMoney
	}

	err = tx.Commit(ctx)
	if err != nil {
		ws.log.Error("Error committing transaction", "walletID", walletID, "amount", amount, "error", err)
		return err
	}

	ws.cache.Set(ctx, walletID.String(), balance)
	ws.log.Info("Withdrawal completed", "walletID", walletID, "amount", amount, "newBalance", balance)
	return nil
}

func (ws *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	balance, ok := ws.cache.Get(ctx, walletID.String())
	if ok {
		return balance, nil
	}

	balance, err := ws.repo.GetBalance(ctx, walletID)
	if err != nil {
		ws.log.Error("Error fetching balance from DB", "walletID", walletID, "error", err)
		return 0, err
	}

	ws.cache.Set(ctx, walletID.String(), balance)
	return balance, nil
}
