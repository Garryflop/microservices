package handler

import (
	"fmt"
	"log"

	"notification-service/internal/domain"
)

// NotificationHandler contains the business logic for processing payment events.
// It simulates sending an email by logging the notification details.
type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// Handle processes a PaymentCompletedEvent and simulates sending an email.
// Returns an error if the notification could not be "sent".
func (h *NotificationHandler) Handle(event domain.PaymentCompletedEvent) error {
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

	return nil
}
