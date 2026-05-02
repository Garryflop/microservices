package domain

import "time"

// PaymentCompletedEvent from the Payment Service
type PaymentCompletedEvent struct {
	EventID       string    `json:"event_id"`
	OrderID       string    `json:"order_id"`
	Amount        int64     `json:"amount"`
	CustomerEmail string    `json:"customer_email"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}
