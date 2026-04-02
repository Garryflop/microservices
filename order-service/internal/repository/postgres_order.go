package repository

import (
	"context"
	"database/sql"
	"fmt"

	"order-service/internal/domain"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		nullableString(order.IdempotencyKey),
		order.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key, ''), created_at
		FROM orders WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var o domain.Order
	if err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.IdempotencyKey, &o.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("scan order: %w", err)
	}
	return &o, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// looks up an order by its idempotency key
func (r *PostgresOrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key, ''), created_at
		FROM orders WHERE idempotency_key = $1
	`
	row := r.db.QueryRowContext(ctx, query, key)

	var o domain.Order
	if err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.IdempotencyKey, &o.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("scan order by idempotency key: %w", err)
	}
	return &o, nil
}

// converts an empty string to a sql.NullString
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
