package repository

import (
	"context"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/storage/postgress"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppRepository struct {
	pool              *pgxpool.Pool
	orderStorage      *postgress.PostgresOrderStorage
	userStorage       *postgress.PostgresUserStorage
	walletStorage     *postgress.PostgresWalletStorage
	withdrawalStorage *postgress.PostgresWithdrawalsStorage
}

func NewAppRepository(pool *pgxpool.Pool, orderStorage *postgress.PostgresOrderStorage, userStorage *postgress.PostgresUserStorage, walletStorage *postgress.PostgresWalletStorage, withdrawalStorage *postgress.PostgresWithdrawalsStorage) AppRepository {
	return AppRepository{
		pool:              pool,
		orderStorage:      orderStorage,
		userStorage:       userStorage,
		walletStorage:     walletStorage,
		withdrawalStorage: withdrawalStorage,
	}
}

func (ar *AppRepository) Register(ctx context.Context, user model.UserModel) error {
	tgx, err := ar.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tgx.Rollback(ctx)

	err = ar.userStorage.Register(ctx, tgx, user)
	if err != nil {
		return err
	}

	walletID, err := ar.walletStorage.CreateWalletForUser(ctx, tgx)
	if err != nil {
		return err
	}

	err = ar.userStorage.BindUserWallet(ctx, tgx, user.Username, walletID)
	if err != nil {
		return err
	}

	err = tgx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (ar *AppRepository) GetByUsername(ctx context.Context, username string) (model.UserModel, error) {
	return ar.userStorage.GetByLogin(ctx, username)
}

func (ar *AppRepository) RegisterOrder(ctx context.Context, order model.RegisterOrderModel, username string) error {
	tgx, err := ar.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tgx.Rollback(ctx)

	userID, err := ar.userStorage.GetUserIdByLogin(ctx, tgx, username)
	if err != nil {
		return err
	}

	order.UserID = userID

	err = ar.orderStorage.RegisterOrder(ctx, tgx, order)
	if err != nil {
		return err
	}

	err = tgx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (ar *AppRepository) GetOrdersByUser(ctx context.Context, username string) ([]model.OrderModel, error) {
	tgx, err := ar.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tgx.Rollback(ctx)

	userID, err := ar.userStorage.GetUserIdByLogin(ctx, tgx, username)
	if err != nil {
		return nil, err
	}

	orders, err := ar.orderStorage.GetOrdersByUser(ctx, tgx, userID)
	if err != nil {
		return nil, err
	}

	err = tgx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (ar *AppRepository) GetBalanceAndWithdrawnInCentsByUser(ctx context.Context, username string) (int, int, error) {
	tgx, err := ar.pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tgx.Rollback(ctx)

	userID, err := ar.userStorage.GetUserIdByLogin(ctx, tgx, username)
	if err != nil {
		return 0, 0, err
	}

	walletID, err := ar.userStorage.GetWalletIdByUserId(ctx, tgx, userID)
	if err != nil {
		return 0, 0, err
	}

	balance, withdrawn, err := ar.walletStorage.GetBalanceAndWithdrawnInCentsByUser(ctx, tgx, walletID)
	if err != nil {
		return 0, 0, err
	}

	err = tgx.Commit(ctx)
	if err != nil {
		return 0, 0, err
	}

	return balance, withdrawn, nil
}

func (ar *AppRepository) WithdrawBalance(ctx context.Context, username string, withdrawal model.WithdrawalModel) error {
	tgx, err := ar.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tgx.Rollback(ctx)

	userID, err := ar.userStorage.GetUserIdByLogin(ctx, tgx, username)
	if err != nil {
		return err
	}

	walletID, err := ar.userStorage.GetWalletIdByUserId(ctx, tgx, userID)
	if err != nil {
		return err
	}

	err = ar.walletStorage.Withdraw(ctx, tgx, walletID, withdrawal.Sum)
	if err != nil {
		return err
	}

	err = ar.withdrawalStorage.WithdrawBalance(ctx, tgx, walletID, withdrawal)
	if err != nil {
		return err
	}

	err = tgx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (ar *AppRepository) GetWithdrawalsByUser(ctx context.Context, username string) ([]model.WithdrawalModel, error) {
	tgx, err := ar.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tgx.Rollback(ctx)

	userID, err := ar.userStorage.GetUserIdByLogin(ctx, tgx, username)
	if err != nil {
		return nil, err
	}

	walletID, err := ar.userStorage.GetWalletIdByUserId(ctx, tgx, userID)
	if err != nil {
		return nil, err
	}

	withdrawals, err := ar.withdrawalStorage.GetWithdrawalsByUser(ctx, tgx, walletID)
	if err != nil {
		return nil, err
	}

	err = tgx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (ar *AppRepository) GetNonProcessedOrdersID(ctx context.Context) ([]string, error) {
	return ar.orderStorage.GetNonProcessedOrdersID(ctx)
}

func (ar *AppRepository) UpdateOrderStatus(ctx context.Context, orderID string, accrual float64, status string) error {
	tx, err := ar.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = ar.orderStorage.UpdateOrder(ctx, tx, orderID, accrual, status)
	if err != nil {
		return err
	}

	userID, err := ar.orderStorage.GetUserIdByOrderID(ctx, tx, orderID)
	if err != nil {
		return err
	}

	walletID, err := ar.userStorage.GetWalletIdByUserId(ctx, tx, userID)
	if err != nil {
		return err
	}

	err = ar.walletStorage.Deposit(ctx, tx, walletID, accrual)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
