package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"notification-service/internal/domain"
	"notification-service/internal/provider"
	"notification-service/internal/store"
)

type NotificationHandler struct {
	store      *store.RedisIdempotencyStore
	sender     provider.EmailSender
	maxRetries int
}

func NewNotificationHandler(s *store.RedisIdempotencyStore, sender provider.EmailSender, maxRetries int) *NotificationHandler {
	return &NotificationHandler{
		store:      s,
		sender:     sender,
		maxRetries: maxRetries,
	}
}

func (h *NotificationHandler) Handle(ctx context.Context, event domain.PaymentCompletedEvent) error {
	paymentID := event.OrderID // using order_id as the payment identifier

	// idempotency check — skip if already sent
	if h.store.IsProcessed(ctx, paymentID) {
		log.Printf("[Worker] Duplicate event for payment %s, skipping", paymentID)
		return nil
	}

	// mark as pending
	h.store.MarkStatus(ctx, paymentID, store.StatusPending)

	dollars := float64(event.Amount) / 100.0
	subject := fmt.Sprintf("Payment Confirmation for Order #%s", event.OrderID)
	body := fmt.Sprintf("Your payment of $%.2f for Order #%s has been %s.",
		dollars, event.OrderID, event.Status)

	// exponential backoff retry loop
	var lastErr error
	for attempt := 1; attempt <= h.maxRetries; attempt++ {
		err := h.sender.SendNotification(ctx, event.CustomerEmail, subject, body)
		if err == nil {
			// success
			h.store.MarkStatus(ctx, paymentID, store.StatusSent)
			log.Printf("[Worker] Notification sent for payment %s (attempt %d/%d)",
				paymentID, attempt, h.maxRetries)
			return nil
		}

		lastErr = err
		log.Printf("[Worker] Attempt %d/%d failed for payment %s: %v",
			attempt, h.maxRetries, paymentID, err)

		if attempt < h.maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second // 2s, 4s, 8s, 16s, 32s
			log.Printf("[Worker] Retrying in %s...", backoff)
			time.Sleep(backoff)
		}
	}

	// all retries exhausted
	h.store.MarkStatus(ctx, paymentID, store.StatusFailed)
	log.Printf("[Worker] All %d retries exhausted for payment %s", h.maxRetries, paymentID)
	return fmt.Errorf("max retries reached for payment %s: %w", paymentID, lastErr)
}
