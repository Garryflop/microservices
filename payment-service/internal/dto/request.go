package dto

// incoming payload for processing a payment.
type CreatePaymentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount" binding:"required"`
}
