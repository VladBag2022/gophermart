package storage

import (
	"context"
)

type OrderInfo struct {
	Number     int64   `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type BalanceInfo struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawalInfo struct {
	Order       int64   `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type Repository interface {
	IsLoginAvailable(
		ctx context.Context,
		login string,
	) (available bool, err error)

	Register(
		ctx context.Context,
		login, password string,
	) error

	Login(
		ctx context.Context,
		login, password string,
	) (success bool, err error)

	OrderOwner(
		ctx context.Context,
		order int64,
	) (login string, err error)

	UploadOrder(
		ctx context.Context,
		login string,
		order int64,
	) error

	Orders(
		ctx context.Context,
		login string,
	) (orders []OrderInfo, err error)

	AccrualOrders(
		ctx context.Context,
	) (orders []int64, err error)

	UpdateOrder(
		ctx context.Context,
		order int64,
		status string,
		accrual float64,
	) error

	Balance(
		ctx context.Context,
		login string,
	) (balance BalanceInfo, err error)

	Withdraw(
		ctx context.Context,
		login string,
		order int64,
		sum float64,
	) error

	Withdrawals(
		ctx context.Context,
		login string,
	) (withdrawals []WithdrawalInfo, err error)

	Close() error
}
