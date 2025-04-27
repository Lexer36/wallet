package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	handlerModel "wallet/internal/model/handler"
	walletModel "wallet/internal/model/wallet"
	"wallet/internal/rest"
	"wallet/internal/rest/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockgen -destination=mock_wallet_service.go -package=mocks wallet/internal/rest WalletService

func TestWalletHandler_WalletOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockWalletService(ctrl)
	handler := rest.NewWalletHandler(svc)

	walletID := uuid.New()

	tests := []struct {
		name           string
		request        handlerModel.WalletOperationRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "successful deposit",
			request: handlerModel.WalletOperationRequest{
				WalletID:      walletID,
				Amount:        100,
				OperationType: handlerModel.OperationDeposit,
			},
			setupMock: func() {
				svc.EXPECT().
					Deposit(gomock.Any(), walletID, int64(100)).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful withdraw",
			request: handlerModel.WalletOperationRequest{
				WalletID:      walletID,
				Amount:        50,
				OperationType: handlerModel.OperationWithdraw,
			},
			setupMock: func() {
				svc.EXPECT().
					Withdraw(gomock.Any(), walletID, int64(50)).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid operation",
			request: handlerModel.WalletOperationRequest{
				WalletID:      walletID,
				Amount:        50,
				OperationType: "invalid",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid amount",
			request: handlerModel.WalletOperationRequest{
				WalletID:      walletID,
				Amount:        0,
				OperationType: handlerModel.OperationDeposit,
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "withdraw not enough money",
			request: handlerModel.WalletOperationRequest{
				WalletID:      walletID,
				Amount:        200,
				OperationType: handlerModel.OperationWithdraw,
			},
			setupMock: func() {
				svc.EXPECT().
					Withdraw(gomock.Any(), walletID, int64(200)).
					Return(walletModel.ErrNotEnoughMoney)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "service error",
			request: handlerModel.WalletOperationRequest{
				WalletID:      walletID,
				Amount:        150,
				OperationType: handlerModel.OperationDeposit,
			},
			setupMock: func() {
				svc.EXPECT().
					Deposit(gomock.Any(), walletID, int64(150)).
					Return(errors.New("some service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/operation", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.WalletOperation(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			require.Equal(t, tt.expectedStatus, res.StatusCode)
		})
	}
}

func TestWalletHandler_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockWalletService(ctrl)
	handler := rest.NewWalletHandler(svc)

	walletID := uuid.New()

	tests := []struct {
		name           string
		walletID       string
		setupMock      func()
		expectedStatus int
		expectedBody   *handlerModel.WalletOperationResponse
	}{
		{
			name:     "successful get balance",
			walletID: walletID.String(),
			setupMock: func() {
				svc.EXPECT().
					GetBalance(gomock.Any(), walletID).
					Return(int64(1000), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &handlerModel.WalletOperationResponse{
				WalletID: walletID,
				Balance:  1000,
			},
		},
		{
			name:     "invalid uuid format",
			walletID: "invalid-uuid",
			setupMock: func() {
				// no call to service
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name:     "wallet not found",
			walletID: walletID.String(),
			setupMock: func() {
				svc.EXPECT().
					GetBalance(gomock.Any(), walletID).
					Return(int64(0), walletModel.ErrWalletNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
		},
		{
			name:     "internal service error",
			walletID: walletID.String(),
			setupMock: func() {
				svc.EXPECT().
					GetBalance(gomock.Any(), walletID).
					Return(int64(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+tt.walletID, nil)
			rec := httptest.NewRecorder()

			handler.GetBalance(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			require.Equal(t, tt.expectedStatus, res.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				var resp handlerModel.WalletOperationResponse
				err := json.NewDecoder(res.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody.WalletID, resp.WalletID)
				require.Equal(t, tt.expectedBody.Balance, resp.Balance)
			}
		})
	}
}
