package accrual

import (
	"VladBag2022/gophermart/internal/storage"
	"context"
	"time"
)

type Daemon struct {
	repository storage.Repository
	accrualAddress string
}

func NewDaemon(repository storage.Repository, accrualAddress string) Daemon {
	return Daemon{
		repository: repository,
		accrualAddress:     accrualAddress,
	}
}

func (d Daemon) Start(ctx context.Context) error {
	for {
		select {
		case <- ctx.Done():
			return nil
		default:
			orders, err := d.repository.AccrualOrders(ctx)
			if err != nil {
				return err
			}
			for _, order := range orders {
				info, retryAfter, err := d.orderInfo(order)
				if err != nil {
					return err
				}
				if info == nil {
					select{
					case <-ctx.Done():
						return nil
					case <-time.After(time.Duration(retryAfter) * time.Second):
						continue
					}
				}
				err = d.repository.UpdateOrder(ctx, order, info.Status, info.Accrual)
				if err != nil {
					return nil
				}
			}
		}
	}
}