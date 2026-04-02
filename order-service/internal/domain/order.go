package domain

import (
	"errors"
	"time"
)

const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

var (
	ErrInvalidAmount     = errors.New("amount must be greater than 0")
	ErrEmptyCustomerID   = errors.New("customer_id is required")
	ErrEmptyItemName     = errors.New("item_name is required")
	ErrOrderNotFound     = errors.New("order not found")
	ErrCancelPaidOrder   = errors.New("cannot cancel a paid order")
	ErrCancelFailedOrder = errors.New("cannot cancel a failed order")
	ErrAlreadyCancelled  = errors.New("order is already cancelled")
)

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64
	Status         string
	IdempotencyKey string
	CreatedAt      time.Time
}

func NewOrder(customerID, itemName string, amount int64, idempotencyKey string) (*Order, error) {
	o := &Order{
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         StatusPending,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return ErrEmptyCustomerID
	}
	if o.ItemName == "" {
		return ErrEmptyItemName
	}
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

func (o *Order) MarkPaid() {
	o.Status = StatusPaid
}

func (o *Order) MarkFailed() {
	o.Status = StatusFailed
}

func (o *Order) Cancel() error {
	switch o.Status {
	case StatusPaid:
		return ErrCancelPaidOrder
	case StatusFailed:
		return ErrCancelFailedOrder
	case StatusCancelled:
		return ErrAlreadyCancelled
	}
	o.Status = StatusCancelled
	return nil
}
