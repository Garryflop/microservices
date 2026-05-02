package handler

import (
	"fmt"
	"log"

	"notification-service/internal/domain"
	"notification-service/internal/store"
)

// NotificationHandler contains the business logic for processing payment events.
// It simulates sending an email by logging the notification details.
// Uses an IdempotencyStore to prevent duplicate processing.
type NotificationHandler struct {
	store *store.IdempotencyStore
}

func NewNotificationHandler(s *store.IdempotencyStore) *NotificationHandler {
	return &NotificationHandler{store: s}
}

// Handle processes a PaymentCompletedEvent and simulates sending an email.
// Returns an error if the notification could not be "sent".
// Duplicate events (same EventID) are skipped and acknowledged.
func (h *NotificationHandler) Handle(event domain.PaymentCompletedEvent) error {
	// Idempotency check — skip if already processed
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

	// Mark as processed only after successful "send"
	h.store.MarkProcessed(event.EventID)

	return nil
}
