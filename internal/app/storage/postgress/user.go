package postgress

import (
	"context"
	"errors"
	"fmt"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserStorage struct {
	db *pgxpool.Pool
}

const uniqueViolation = "23505"

func NewPostgresUserStorage(pool *pgxpool.Pool) (*PostgresUserStorage, error) {
	return &PostgresUserStorage{db: pool}, nil
}

func (p *PostgresUserStorage) Init(ctx context.Context) error {
	_, err := p.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			hashed_password TEXT NOT NULL, 
			wallet_id INT,
			FOREIGN KEY (wallet_id) REFERENCES wallets(wallet_id)
		)
	`)
	return err
}

func (p *PostgresUserStorage) Register(ctx context.Context, tx pgx.Tx, user model.UserModel) error {
	_, err := tx.Exec(ctx, "INSERT INTO users (username, hashed_password) VALUES ($1, $2)", user.Username, user.HashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueViolation {
				return fmt.Errorf("username %s already exists", user.Username)
			}
		}
		return err
	}
	return nil
}

func (p *PostgresUserStorage) GetByLogin(ctx context.Context, username string) (model.UserModel, error) {
	row := p.db.QueryRow(ctx, "SELECT username, hashed_password, wallet_id FROM users WHERE username = $1", username)
	user := model.UserModel{}
	err := row.Scan(&user.Username, &user.HashedPassword, &user.WalletId)
	return user, err
}

func (p *PostgresUserStorage) BindUserWallet(ctx context.Context, tx pgx.Tx, username string, walletID int) error {
	_, err := tx.Exec(ctx, "UPDATE users SET wallet_id = $1 WHERE username = $2", walletID, username)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresUserStorage) GetUserIdByLogin(ctx context.Context, tx pgx.Tx, username string) (int, error) {
	var userID int
	err := tx.QueryRow(ctx, "SELECT user_id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (p *PostgresUserStorage) GetWalletIdByUserId(ctx context.Context, tx pgx.Tx, userID int) (int, error) {
	var walletID int
	err := tx.QueryRow(ctx, "SELECT wallet_id FROM users WHERE user_id = $1", userID).Scan(&walletID)
	if err != nil {
		return 0, err
	}
	return walletID, nil
}
