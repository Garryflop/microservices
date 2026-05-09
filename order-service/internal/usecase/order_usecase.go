package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"order-service/internal/domain"
)

var ErrPaymentServiceUnavailable = errors.New("payment service unavailable")

type OrderUseCase struct {
	repo          OrderRepository
	paymentClient PaymentClient
	cache         OrderCache
}

func NewOrderUseCase(repo OrderRepository, paymentClient PaymentClient, cache OrderCache) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
		cache:         cache,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, error) {
	if idempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	order, err := domain.NewOrder(customerID, itemName, amount, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("invalid order: %w", err)
	}
	order.ID = uuid.New().String()

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	paymentResp, err := uc.paymentClient.AuthorizePayment(ctx, order.ID, order.Amount)
	if err != nil {
		order.MarkFailed()
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.StatusFailed)
		_ = uc.cache.Delete(ctx, order.ID) // invalidate cache
		return order, ErrPaymentServiceUnavailable
	}

	if paymentResp.Status == "Authorized" {
		order.MarkPaid()
		if err := uc.repo.UpdateStatus(ctx, order.ID, domain.StatusPaid); err != nil {
			return nil, fmt.Errorf("failed to update order status: %w", err)
		}
		_ = uc.cache.Delete(ctx, order.ID) // invalidate cache
	} else {
		order.MarkFailed()
		if err := uc.repo.UpdateStatus(ctx, order.ID, domain.StatusFailed); err != nil {
			return nil, fmt.Errorf("failed to update order status: %w", err)
		}
		_ = uc.cache.Delete(ctx, order.ID) // invalidate cache
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	// cache-aside: check cache first
	if cached, err := uc.cache.Get(ctx, id); err == nil {
		return cached, nil
	}

	// cache miss: query database
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// populate cache for next read
	_ = uc.cache.Set(ctx, id, order)

	return order, nil
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := order.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, domain.StatusCancelled); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}
	_ = uc.cache.Delete(ctx, order.ID) // invalidate cache

	return order, nil
}
