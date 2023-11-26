package postgress

import (
	"context"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWalletStorage struct {
	db *pgxpool.Pool
}

func NewPostgresWalletStorage(pool *pgxpool.Pool) (*PostgresWalletStorage, error) {
	return &PostgresWalletStorage{db: pool}, nil
}

func (p *PostgresWalletStorage) Init(ctx context.Context) error {
	_, err := p.db.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS wallets (
            wallet_id SERIAL PRIMARY KEY,
            amount INT NOT NULL,
			withdrawn INT DEFAULT 0
        )
    `)
	return err
}

func (p *PostgresWalletStorage) CreateWalletForUser(ctx context.Context, tx pgx.Tx) (int, error) {
	var walletID int
	err := tx.QueryRow(ctx, "INSERT INTO wallets (amount) VALUES ($1) RETURNING wallet_id", 0).Scan(&walletID)
	return walletID, err
}

func (p *PostgresWalletStorage) GetBalanceAndWithdrawnInCentsByUser(ctx context.Context, tx pgx.Tx, walletID int) (int, int, error) {
	var balance int
	var withdrawn int
	err := tx.QueryRow(ctx, "SELECT amount, withdrawn FROM wallets WHERE wallet_id = $1", walletID).Scan(&balance, &withdrawn)
	if err != nil {
		return 0, 0, err
	}
	return balance, withdrawn, nil
}

func (p *PostgresWalletStorage) Deposit(ctx context.Context, tx pgx.Tx, walletID int, amount float64) error {
	cents := int(amount * 100)
	balance, _, err := p.GetBalanceAndWithdrawnInCentsByUser(ctx, tx, walletID)
	if err != nil {
		return err
	}
	balance = cents + balance
	_, err = p.db.Exec(ctx, "UPDATE wallets SET amount = $1 WHERE wallet_id = $2", balance, walletID)
	return err
}

func (p *PostgresWalletStorage) Withdraw(ctx context.Context, tx pgx.Tx, walletID int, amount float64) error {
	cents := int(amount * 100)
	balance, withdrawn, err := p.GetBalanceAndWithdrawnInCentsByUser(ctx, tx, walletID)
	if err != nil {
		return err
	}
	if balance < cents {
		return &apperrors.ErrInsufficientBalance{}
	}

	balance -= cents
	withdrawn += cents

	_, err = tx.Exec(ctx, "UPDATE wallets SET amount = $1, withdrawn = $2 WHERE wallet_id = $3", balance, withdrawn, walletID)
	if err != nil {
		return err
	}
	return nil
}
