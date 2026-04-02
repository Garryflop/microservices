package repository

import (
	"context"
	"database/sql"
	"fmt"

	"payment-service/internal/domain"
)

// using PostgreSQL
type PostgresPaymentRepository struct {
	db *sql.DB
}

// creates a new repository backed by PostgreSQL
func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// inserts a new payment into the database
func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

// retrieves a payment by its associated order ID
func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments WHERE order_id = $1
	`
	row := r.db.QueryRowContext(ctx, query, orderID)

	var p domain.Payment
	if err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("scan payment: %w", err)
	}
	return &p, nil
}
