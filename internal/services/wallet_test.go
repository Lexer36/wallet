package services_test

import (
	"context"
	"errors"
	"wallet/internal/model/wallet"

	"go.uber.org/mock/gomock"
	"log/slog"
	"testing"

	"wallet/internal/services"
	"wallet/internal/services/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockgen -destination=mock_walletstorage.go -package=services wallet/internal/services WalletStorage
// mockgen -destination=mock_walletcache.go -package=services wallet/internal/services WalletCache
// mockgen -destination=mock_tx.go -package=services github.com/jackc/pgx/v5 Tx

func TestWalletService_Deposit_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWalletStorage(ctrl)
	cache := mocks.NewMockWalletCache(ctrl)
	logger := slog.Default()

	service := services.NewWalletService(repo, cache, logger)

	walletID := uuid.New()
	amount := int64(100)
	updatedBalance := int64(200)

	tx := mocks.NewMockTx(ctrl)

	repo.EXPECT().
		BeginTx(gomock.Any(), gomock.Any()).
		Return(tx, nil)

	repo.EXPECT().
		Deposit(gomock.Any(), tx, walletID, amount).
		Return(updatedBalance, nil)

	tx.EXPECT().
		Commit(gomock.Any()).
		Return(nil)

	cache.EXPECT().
		Set(gomock.Any(), walletID.String(), updatedBalance)

	tx.EXPECT().
		Rollback(gomock.Any()).AnyTimes()

	err := service.Deposit(context.Background(), walletID, amount)
	require.NoError(t, err)
}

func TestWalletService_Deposit_DepositError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWalletStorage(ctrl)
	cache := mocks.NewMockWalletCache(ctrl)
	logger := slog.Default()

	service := services.NewWalletService(repo, cache, logger)

	walletID := uuid.New()
	amount := int64(100)

	tx := mocks.NewMockTx(ctrl)

	repo.EXPECT().
		BeginTx(gomock.Any(), gomock.Any()).
		Return(tx, nil)

	repo.EXPECT().
		Deposit(gomock.Any(), tx, walletID, amount).
		Return(int64(0), errors.New("deposit error"))

	tx.EXPECT().
		Rollback(gomock.Any())

	err := service.Deposit(context.Background(), walletID, amount)
	require.Error(t, err)
}

func TestWalletService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWalletStorage(ctrl)
	cache := mocks.NewMockWalletCache(ctrl)
	logger := slog.Default()

	service := services.NewWalletService(repo, cache, logger)

	walletID := uuid.New()
	amount := int64(50)

	tests := []struct {
		name           string
		withdrawReturn int64
		withdrawError  error
		commitError    error
		expectError    error
	}{
		{
			name:           "successful withdraw",
			withdrawReturn: 150,
			expectError:    nil,
		},
		{
			name:           "not enough money",
			withdrawReturn: -10,
			expectError:    wallet.ErrNotEnoughMoney,
		},
		{
			name:          "withdraw error",
			withdrawError: errors.New("withdraw failed"),
			expectError:   errors.New("withdraw failed"),
		},
		{
			name:           "commit error",
			withdrawReturn: 100,
			commitError:    errors.New("commit failed"),
			expectError:    errors.New("commit failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := mocks.NewMockTx(ctrl)

			repo.EXPECT().
				BeginTx(gomock.Any(), gomock.Any()).
				Return(tx, nil)

			repo.EXPECT().
				Withdraw(gomock.Any(), tx, walletID, amount).
				Return(tt.withdrawReturn, tt.withdrawError)

			if tt.withdrawError == nil && tt.withdrawReturn >= 0 && tt.commitError == nil {
				tx.EXPECT().
					Commit(gomock.Any()).
					Return(nil)

				cache.EXPECT().
					Set(gomock.Any(), walletID.String(), tt.withdrawReturn)
			} else if tt.withdrawError == nil && tt.withdrawReturn >= 0 {
				// Commit вызовется, но ошибка вернется
				tx.EXPECT().
					Commit(gomock.Any()).
					Return(tt.commitError)
			}

			tx.EXPECT().
				Rollback(gomock.Any()).AnyTimes()

			err := service.Withdraw(context.Background(), walletID, amount)

			if tt.expectError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWalletService_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWalletStorage(ctrl)
	cache := mocks.NewMockWalletCache(ctrl)
	logger := slog.Default()

	service := services.NewWalletService(repo, cache, logger)

	walletID := uuid.New()

	tests := []struct {
		name            string
		cacheHit        bool
		cacheBalance    int64
		dbBalance       int64
		dbError         error
		expectedBalance int64
		expectError     bool
	}{
		{
			name:            "balance from cache",
			cacheHit:        true,
			cacheBalance:    500,
			expectedBalance: 500,
			expectError:     false,
		},
		{
			name:            "balance from db",
			cacheHit:        false,
			dbBalance:       300,
			expectedBalance: 300,
			expectError:     false,
		},
		{
			name:        "db returns error",
			cacheHit:    false,
			dbError:     errors.New("db error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cache.EXPECT().
				Get(ctx, walletID.String()).
				Return(tt.cacheBalance, tt.cacheHit)

			if !tt.cacheHit {
				repo.EXPECT().
					GetBalance(ctx, walletID).
					Return(tt.dbBalance, tt.dbError)

				if tt.dbError == nil {
					cache.EXPECT().
						Set(ctx, walletID.String(), tt.dbBalance)
				}
			}

			balance, err := service.GetBalance(ctx, walletID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedBalance, balance)
			}
		})
	}
}
