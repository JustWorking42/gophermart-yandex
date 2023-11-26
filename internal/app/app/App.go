package app

import (
	"context"

	"github.com/JustWorking42/gophermart-yandex/internal/app/config"
	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/JustWorking42/gophermart-yandex/internal/app/storage/postgress"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	context    context.Context
	Repository *repository.AppRepository
}

func NewApp(ctx context.Context, config config.Config) (*App, error) {

	dbPool, err := pgxpool.New(ctx, config.DatabaseURI)
	if err != nil {
		return nil, err
	}
	orderStorage, err := postgress.NewPostgresOrderStorage(dbPool)
	if err != nil {
		return nil, err
	}
	userStorage, err := postgress.NewPostgresUserStorage(dbPool)
	if err != nil {
		return nil, err
	}
	walletStorage, err := postgress.NewPostgresWalletStorage(dbPool)
	if err != nil {
		return nil, err
	}

	withdrawalsStorage, err := postgress.NewPostgresWithdrawalsStorage(dbPool)
	if err != nil {
		return nil, err
	}

	err = walletStorage.Init(ctx)
	if err != nil {
		return nil, err
	}
	err = userStorage.Init(ctx)
	if err != nil {
		return nil, err
	}
	err = orderStorage.Init(ctx)
	if err != nil {
		return nil, err
	}
	err = withdrawalsStorage.Init(ctx)
	if err != nil {
		return nil, err
	}

	repository := repository.NewAppRepository(dbPool, orderStorage, userStorage, walletStorage, withdrawalsStorage)

	return &App{
		context:    ctx,
		Repository: &repository,
	}, nil
}
