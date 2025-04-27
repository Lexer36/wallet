package postgres_test

import (
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"wallet/internal/model/wallet"
	"wallet/internal/repository/postgres"
)

func TestStorage_Withdraw(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	walletID := uuid.New()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	storage := postgres.New(mockPool)

	tests := []struct {
		name            string
		amount          int64
		expectedError   error
		expectedBalance int64
	}{
		{
			name:            "successful withdraw",
			amount:          100,
			expectedError:   nil,
			expectedBalance: 900,
		},
		{
			name:            "negative balance",
			amount:          1000,
			expectedError:   nil,
			expectedBalance: -50,
		},
		{
			name:            "wallet not found",
			amount:          100,
			expectedError:   wallet.ErrWalletNotFound,
			expectedBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool.ExpectBegin()
			mockTx, err := mockPool.Begin(ctx)
			require.NoError(t, err)

			mockPool.ExpectQuery(regexp.QuoteMeta(`
				UPDATE wallets
				SET balance = balance + $1
				WHERE id = $2
				RETURNING balance;
			`)).
				WithArgs(-tt.amount, walletID).
				WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(tt.expectedBalance)).
				WillReturnError(tt.expectedError)

			balance, err := storage.Withdraw(ctx, mockTx, walletID, tt.amount)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedBalance, balance)

			_ = mockTx.Rollback(ctx)
		})
	}

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestStorage_Deposit(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	walletID := uuid.New()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	storage := postgres.New(mockPool)

	tests := []struct {
		name            string
		amount          int64
		expectedError   error
		expectedBalance int64
	}{
		{
			name:            "successful deposit",
			amount:          100,
			expectedError:   nil,
			expectedBalance: 900,
		},
		{
			name:            "wallet not found",
			amount:          100,
			expectedError:   wallet.ErrWalletNotFound,
			expectedBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool.ExpectBegin()
			mockTx, err := mockPool.Begin(ctx)
			require.NoError(t, err)

			mockPool.ExpectQuery(regexp.QuoteMeta(`
				UPDATE wallets
				SET balance = balance + $1
				WHERE id = $2
				RETURNING balance;
			`)).
				WithArgs(tt.amount, walletID).
				WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(tt.expectedBalance)).
				WillReturnError(tt.expectedError)

			balance, err := storage.Deposit(ctx, mockTx, walletID, tt.amount)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedBalance, balance)

			_ = mockTx.Rollback(ctx)
		})
	}

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestStorage_GetBalance(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	walletID := uuid.New()

	// Замокаем pgxpool.Pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	storage := postgres.New(mockPool)

	tests := []struct {
		name            string
		expectedError   error
		expectedBalance int64
	}{
		{
			name:            "successful get balance",
			expectedError:   nil,
			expectedBalance: 100,
		},
		{
			name:            "wallet not found",
			expectedError:   wallet.ErrWalletNotFound,
			expectedBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool.ExpectQuery(regexp.QuoteMeta(`
				SELECT balance 
				FROM wallets 
				WHERE id = $1
			`)).
				WithArgs(walletID).
				WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(tt.expectedBalance)).
				WillReturnError(tt.expectedError)

			balance, err := storage.GetBalance(ctx, walletID)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedBalance, balance)

			// Проверка выполнения всех моков
			require.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
