package postgres_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"wallet/internal/model/repository"
	"wallet/internal/repository/postgres"
)

func TestStorage_Withdraw(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	walletID := uuid.New()

	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	storage := postgres.New(dbConn)

	tests := []struct {
		name            string
		amount          int64
		expectedError   error
		expectedBalance int64
		walletID        uuid.UUID
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
			expectedError:   repo.ErrWalletNotFound,
			expectedBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//t.Parallel()

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`
					UPDATE wallets
					SET balance = balance + $1
					WHERE wallet_id = $2
					RETURNING balance;
				`)).
				WithArgs(-tt.amount, walletID).
				WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(tt.expectedBalance)).
				WillReturnError(tt.expectedError)

			// просто инициализируем транзакцию (без моков Begin/Commit/Rollback)
			tx, err := dbConn.Begin()
			require.NoError(t, err)

			balance, err := storage.Withdraw(ctx, tx, walletID, tt.amount)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedBalance, balance)

			// обязательно закрываем транзакцию
			_ = tx.Rollback()
		})
	}

	require.NoError(t, mock.ExpectationsWereMet())
}
