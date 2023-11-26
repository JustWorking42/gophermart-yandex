package postgress

import (
	"context"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/model/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ErrCodeDuplicateObject     = "23505"
	ErrCodeForeignKeyViolation = "23503"
)

type PostgresOrderStorage struct {
	db *pgxpool.Pool
}

func NewPostgresOrderStorage(pool *pgxpool.Pool) (*PostgresOrderStorage, error) {
	return &PostgresOrderStorage{db: pool}, nil
}

func (p *PostgresOrderStorage) Init(ctx context.Context) error {
	_, err := p.db.Exec(ctx, `
	DO $$ BEGIN
    	CREATE TYPE status_type AS ENUM ('NEW', 'REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED');
	EXCEPTION
    	WHEN duplicate_object THEN null;
	END $$;
		CREATE TABLE IF NOT EXISTS orders (
			order_id TEXT PRIMARY KEY UNIQUE,
			user_id INT NOT NULL,
			status status_type NOT NULL DEFAULT 'NEW',
			accrual NUMERIC(10,2) DEFAULT 0,
			uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)
	`)
	return err
}

func (p *PostgresOrderStorage) RegisterOrder(ctx context.Context, tx pgx.Tx, order model.RegisterOrderModel) error {
	_, err := tx.Exec(ctx, "INSERT INTO orders (user_id, order_id) VALUES ($1, $2)", order.UserID, order.OrderID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == ErrCodeDuplicateObject {
				var userID int
				err := p.db.QueryRow(ctx, "SELECT user_id FROM orders WHERE order_id = $1", order.OrderID).Scan(&userID)
				if err != nil {
					return err
				}
				if userID == order.UserID {
					return &apperrors.ErrAlreadyRegisteredByThisUser{}
				} else {
					return &apperrors.ErrAlreadyRegisteredByAnotherUser{}
				}
			}
			if pgErr.Code == ErrCodeForeignKeyViolation {
				return &apperrors.ErrUserDoesNotExist{}
			}
		}
		return err
	}
	return nil
}

func (p *PostgresOrderStorage) GetOrdersByUser(ctx context.Context, tx pgx.Tx, userID int) ([]model.OrderModel, error) {
	rows, err := tx.Query(ctx, "SELECT order_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.OrderModel
	for rows.Next() {
		var order model.OrderModel
		if err := rows.Scan(&order.OrderID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (p *PostgresOrderStorage) GetNonProcessedOrdersID(ctx context.Context) ([]string, error) {
	rows, err := p.db.Query(ctx, "SELECT order_id FROM orders WHERE status IN ('NEW', 'REGISTERED', 'PROCESSING')")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []string
	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			return nil, err
		}
		orders = append(orders, orderID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (p *PostgresOrderStorage) UpdateOrder(ctx context.Context, tx pgx.Tx, orderID string, accrual float64, status string) error {
	_, err := tx.Exec(ctx, `
        UPDATE orders 
        SET accrual = $2, status = $3 
        WHERE order_id = $1
    `, orderID, accrual, status)

	return err
}

func (p *PostgresOrderStorage) GetUserIdByOrderID(ctx context.Context, tx pgx.Tx, orderID string) (int, error) {
	var userID int
	err := tx.QueryRow(ctx, "SELECT user_id FROM orders WHERE order_id = $1", orderID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
