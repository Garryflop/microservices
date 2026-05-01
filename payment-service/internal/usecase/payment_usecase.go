package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"payment-service/internal/domain"
)

// implements the business logic for payment operations
type PaymentUseCase struct {
	repo      PaymentRepository
	publisher EventPublisher
}

func NewPaymentUseCase(repo PaymentRepository, publisher EventPublisher) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, publisher: publisher}
}

// creates and stores a new payment, applying business rules
func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error) {
	payment, err := domain.NewPayment(orderID, amount)
	if err != nil {
		return nil, err
	}

	payment.ID = uuid.New().String()

	if payment.Status == domain.StatusAuthorized {
		payment.TransactionID = uuid.New().String()
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to store payment: %w", err)
	}

	// Publish event after successful DB commit (at-least-once delivery)
	if payment.Status == domain.StatusAuthorized {
		event := PaymentCompletedEvent{
			EventID:       uuid.New().String(),
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			CustomerEmail: "user@example.com",
			Status:        payment.Status,
			Timestamp:     time.Now(),
		}

		if err := uc.publisher.PublishPaymentCompleted(ctx, event); err != nil {
			// Log the error but don't fail the payment — the message broker
			// might be temporarily unavailable. In production, use an outbox pattern.
			log.Printf("[PaymentUseCase] WARNING: failed to publish event: %v", err)
		}
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetPayment(ctx context.Context, orderID string) (*domain.Payment, error) {
	payment, err := uc.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

