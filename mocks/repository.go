// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	storage "VladBag2022/gophermart/internal/storage"
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// AccrualOrders provides a mock function with given fields: ctx
func (_m *Repository) AccrualOrders(ctx context.Context) ([]int64, error) {
	ret := _m.Called(ctx)

	var r0 []int64
	if rf, ok := ret.Get(0).(func(context.Context) []int64); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int64)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Balance provides a mock function with given fields: ctx, login
func (_m *Repository) Balance(ctx context.Context, login string) (storage.BalanceInfo, error) {
	ret := _m.Called(ctx, login)

	var r0 storage.BalanceInfo
	if rf, ok := ret.Get(0).(func(context.Context, string) storage.BalanceInfo); ok {
		r0 = rf(ctx, login)
	} else {
		r0 = ret.Get(0).(storage.BalanceInfo)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Close provides a mock function with given fields:
func (_m *Repository) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsLoginAvailable provides a mock function with given fields: ctx, login
func (_m *Repository) IsLoginAvailable(ctx context.Context, login string) (bool, error) {
	ret := _m.Called(ctx, login)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, login)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Login provides a mock function with given fields: ctx, login, password
func (_m *Repository) Login(ctx context.Context, login string, password string) (bool, error) {
	ret := _m.Called(ctx, login, password)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, login, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrderOwner provides a mock function with given fields: ctx, order
func (_m *Repository) OrderOwner(ctx context.Context, order int64) (string, error) {
	ret := _m.Called(ctx, order)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, int64) string); ok {
		r0 = rf(ctx, order)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, order)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Orders provides a mock function with given fields: ctx, login
func (_m *Repository) Orders(ctx context.Context, login string) ([]storage.OrderInfo, error) {
	ret := _m.Called(ctx, login)

	var r0 []storage.OrderInfo
	if rf, ok := ret.Get(0).(func(context.Context, string) []storage.OrderInfo); ok {
		r0 = rf(ctx, login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]storage.OrderInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Register provides a mock function with given fields: ctx, login, password
func (_m *Repository) Register(ctx context.Context, login string, password string) error {
	ret := _m.Called(ctx, login, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateOrder provides a mock function with given fields: ctx, order, status, accrual
func (_m *Repository) UpdateOrder(ctx context.Context, order int64, status string, accrual float64) error {
	ret := _m.Called(ctx, order, status, accrual)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string, float64) error); ok {
		r0 = rf(ctx, order, status, accrual)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UploadOrder provides a mock function with given fields: ctx, login, order
func (_m *Repository) UploadOrder(ctx context.Context, login string, order int64) error {
	ret := _m.Called(ctx, login, order)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) error); ok {
		r0 = rf(ctx, login, order)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Withdraw provides a mock function with given fields: ctx, login, order, sum
func (_m *Repository) Withdraw(ctx context.Context, login string, order int64, sum float64) error {
	ret := _m.Called(ctx, login, order, sum)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64, float64) error); ok {
		r0 = rf(ctx, login, order, sum)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Withdrawals provides a mock function with given fields: ctx, login
func (_m *Repository) Withdrawals(ctx context.Context, login string) ([]storage.WithdrawalInfo, error) {
	ret := _m.Called(ctx, login)

	var r0 []storage.WithdrawalInfo
	if rf, ok := ret.Get(0).(func(context.Context, string) []storage.WithdrawalInfo); ok {
		r0 = rf(ctx, login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]storage.WithdrawalInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepository(t mockConstructorTestingTNewRepository) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
