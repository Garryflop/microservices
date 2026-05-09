package usecase

import (
	"context"

	"order-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
}

type OrderCache interface {
	Get(ctx context.Context, id string) (*domain.Order, error)
	Set(ctx context.Context, id string, order *domain.Order) error
	Delete(ctx context.Context, id string) error
}

type PaymentResponse struct {
	Status        string
	TransactionID string
}

type PaymentClient interface {
	AuthorizePayment(ctx context.Context, orderID string, amount int64) (*PaymentResponse, error)
}
