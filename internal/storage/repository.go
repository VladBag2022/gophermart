package storage

import (
	"context"
)

type OrderInfo struct {
	Number     int     `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type BalanceInfo struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawalInfo struct {
	order       int
	sum         float64
	processedAt string
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
		order int,
	) (login string, err error)

	UploadOrder(
		ctx context.Context,
		login string,
		order int,
	) error

	Orders(
		ctx context.Context,
		login string,
	) (orders []OrderInfo, err error)

	Balance(
		ctx context.Context,
		login string,
	) (balance BalanceInfo, err error)

	Withdraw(
		ctx context.Context,
		login string,
		order int,
		sum float64,
	) error

	Withdrawals(
		ctx context.Context,
		login string,
	) (withdrawals []WithdrawalInfo, err error)

	Close() error
}
