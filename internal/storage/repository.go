package storage

import (
	"context"
)

type OrderInfo struct {
	number 		int
	status 		string
	uploadedAt 	string
}

type WithdrawalInfo struct {
	order 		int
	sum 		int
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
		order int,
		login string,
	) error

	Orders(
		ctx context.Context,
		login string,
	) (orders []OrderInfo, err error)

	Balance(
		ctx context.Context,
		login string,
	) (current, withdrawn int, err error)

	Withdraw(
		ctx context.Context,
		login string,
		order int,
		sum int,
	) error

	Withdrawals(
		ctx context.Context,
		login string,
	) (withdrawals []WithdrawalInfo, err error)

	Close() error
}
