package dto

// outgoing payload for a payment
type PaymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id,omitempty"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
