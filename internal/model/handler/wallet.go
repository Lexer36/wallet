package handler

import "github.com/google/uuid"

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidAmount  = errors.New("invalid amount")
)

type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

type WalletOperationRequest struct {
	WalletID      uuid.UUID     `json:"walletId"`
	OperationType OperationType `json:"operationType"`
	Amount        int64         `json:"amount"`
}

type WalletOperationResponse struct {
	WalletID uuid.UUID `json:"walletId"`
	Balance  int64     `json:"balance"`
}
