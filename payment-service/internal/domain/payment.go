package domain

import (
	"errors"
)

const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"
)

const MaxPaymentAmount int64 = 100000

var (
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
	ErrPaymentNotFound = errors.New("payment not found")
	ErrAmountExceeded  = errors.New("amount exceeds the allowed limit")
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
}

func NewPayment(orderID string, amount int64) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	status := StatusAuthorized
	if amount > MaxPaymentAmount {
		status = StatusDeclined
	}

	return &Payment{
		OrderID: orderID,
		Amount:  amount,
		Status:  status,
	}, nil
}
