package usecase

import (
	"context"
	"time"

	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
}

// PaymentCompletedEvent is the event published after a successful payment.
type PaymentCompletedEvent struct {
	EventID       string    `json:"event_id"`
	OrderID       string    `json:"order_id"`
	Amount        int64     `json:"amount"`
	CustomerEmail string    `json:"customer_email"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}

// EventPublisher is the port for publishing payment events to a message broker.
// The infrastructure layer provides the concrete implementation (e.g., RabbitMQ).
type EventPublisher interface {
	PublishPaymentCompleted(ctx context.Context, event PaymentCompletedEvent) error
}
