package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	model "wallet/internal/model/handler"
	"wallet/internal/model/wallet"

	"github.com/google/uuid"
)

type WalletService interface {
	Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error
	Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
}

type WalletHandler struct {
	svc WalletService
}

func NewWalletHandler(svc WalletService) *WalletHandler {
	return &WalletHandler{
		svc: svc,
	}
}

func (h *WalletHandler) WalletOperation(w http.ResponseWriter, r *http.Request) {
	var req model.WalletOperationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, model.ErrInvalidRequest)
		return
	}

	if req.Amount <= 0 {
		h.handleError(w, model.ErrInvalidAmount)
		return
	}

	switch req.OperationType {
	case model.OperationDeposit:
		if err := h.svc.Deposit(r.Context(), req.WalletID, req.Amount); err != nil {
			h.handleError(w, err)
			return
		}
	case model.OperationWithdraw:
		if err := h.svc.Withdraw(r.Context(), req.WalletID, req.Amount); err != nil {
			h.handleError(w, err)
			return
		}
	default:
		http.Error(w, "invalid operation type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	// getting wallet uuid
	walletIDStr := r.URL.Path[len("/api/v1/wallets/"):]
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		h.handleError(w, model.ErrInvalidRequest)
		return
	}

	balance, err := h.svc.GetBalance(r.Context(), walletID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := model.WalletOperationResponse{WalletID: walletID, Balance: balance}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(resp)
}

func (h *WalletHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, wallet.ErrWalletNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, wallet.ErrNotEnoughMoney):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, model.ErrInvalidAmount):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, model.ErrInvalidRequest):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
