package app

import (
	"context"

	"github.com/JustWorking42/gophermart-yandex/internal/app/logger"
	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/JustWorking42/gophermart-yandex/internal/app/storage/postgress"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	context    context.Context
	Repository repository.AppRepository
	Logger     *zap.Logger
}

func NewApp(ctx context.Context, databaseURI string, logLevel string) (*App, error) {
	dbPool, err := pgxpool.New(ctx, databaseURI)
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

	logger, err := logger.CreateLogger(logLevel)
	if err != nil {
		return nil, err
	}

	return &App{
		context:    ctx,
		Repository: &repository,
		Logger:     logger,
	}, nil
}
