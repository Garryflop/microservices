package handler

import (
	"fmt"
	"log"

	"notification-service/internal/domain"
	"notification-service/internal/store"
)

type NotificationHandler struct {
	store *store.IdempotencyStore
}

func NewNotificationHandler(s *store.IdempotencyStore) *NotificationHandler {
	return &NotificationHandler{store: s}
}

func (h *NotificationHandler) Handle(event domain.PaymentCompletedEvent) error {
	// idempotency check
	if h.store.IsProcessed(event.EventID) {
		log.Printf("[Notification] Duplicate event %s, skipping", event.EventID)
		return nil
	}

	dollars := float64(event.Amount) / 100.0

	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
		event.CustomerEmail,
		event.OrderID,
		dollars,
	)

	fmt.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f\n",
		event.CustomerEmail,
		event.OrderID,
		dollars,
	)

	h.store.MarkProcessed(event.EventID)

	return nil
}
