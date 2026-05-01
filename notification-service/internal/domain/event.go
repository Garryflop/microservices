package domain

import "time"

// PaymentCompletedEvent represents the event published by the Payment Service
// after a successful payment. The Notification Service is fully decoupled —
// it only knows about this event structure, not about Order or Payment services.
type PaymentCompletedEvent struct {
	EventID       string    `json:"event_id"`
	OrderID       string    `json:"order_id"`
	Amount        int64     `json:"amount"`
	CustomerEmail string    `json:"customer_email"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}
