package postgress

import (
	"context"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWithdrawalsStorage struct {
	db *pgxpool.Pool
}

func NewPostgresWithdrawalsStorage(pool *pgxpool.Pool) (*PostgresWithdrawalsStorage, error) {
	return &PostgresWithdrawalsStorage{db: pool}, nil
}

func (p *PostgresWithdrawalsStorage) Init(ctx context.Context) error {
	_, err := p.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS withdrawals (
			order_id TEXT PRIMARY KEY UNIQUE,
			amount NUMERIC(10,2) NOT NULL,
			processed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			wallet_id INT NOT NULL,
			FOREIGN KEY (wallet_id) REFERENCES wallets(wallet_id)
		)
	`)
	return err
}

func (p *PostgresWithdrawalsStorage) WithdrawBalance(ctx context.Context, tx pgx.Tx, walletID int, withdrawal model.WithdrawalModel) error {
	_, err := tx.Exec(ctx, "INSERT INTO withdrawals (order_id, amount, wallet_id) VALUES ($1, $2, $3)",
		withdrawal.Order, withdrawal.Sum, walletID)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresWithdrawalsStorage) GetWithdrawalsByUser(ctx context.Context, tx pgx.Tx, walletID int) ([]model.WithdrawalModel, error) {
	rows, err := tx.Query(ctx, "SELECT order_id, amount, processed_at FROM withdrawals WHERE wallet_id = $1 ORDER BY processed_at ASC", walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []model.WithdrawalModel
	for rows.Next() {
		var withdrawal model.WithdrawalModel
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}
