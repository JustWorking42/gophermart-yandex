package repository

import (
	"context"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
)

type AppRepository interface {
	Register(ctx context.Context, user model.UserModel) error
	GetByUsername(ctx context.Context, username string) (*model.UserModel, error)
	RegisterOrder(ctx context.Context, order model.RegisterOrderModel, username string) error
	GetOrdersByUser(ctx context.Context, username string) ([]model.OrderModel, error)
	GetBalanceAndWithdrawnInCentsByUser(ctx context.Context, username string) (int, int, error)
	WithdrawBalance(ctx context.Context, username string, withdrawal model.WithdrawalModel) error
	GetWithdrawalsByUser(ctx context.Context, username string) ([]model.WithdrawalModel, error)
	GetNonProcessedOrdersID(ctx context.Context) ([]string, error)
	UpdateOrderStatus(ctx context.Context, orderID string, accrual float64, status string) error
}
