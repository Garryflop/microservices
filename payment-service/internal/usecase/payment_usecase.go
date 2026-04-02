package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"payment-service/internal/domain"
)

// implements the business logic for payment operations
type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(repo PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
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

	return payment, nil
}

func (uc *PaymentUseCase) GetPayment(ctx context.Context, orderID string) (*domain.Payment, error) {
	payment, err := uc.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return payment, nil
}
