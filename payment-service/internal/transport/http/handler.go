package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"payment-service/internal/domain"
	"payment-service/internal/dto"
	"payment-service/internal/usecase"
)

type Handler struct {
	uc *usecase.PaymentUseCase
}

func NewHandler(uc *usecase.PaymentUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) ProcessPayment(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	payment, err := h.uc.ProcessPayment(c.Request.Context(), req.OrderID, req.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to process payment"})
		return
	}

	c.JSON(http.StatusCreated, toPaymentResponse(payment))
}

func (h *Handler) GetPayment(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.uc.GetPayment(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get payment"})
		return
	}

	c.JSON(http.StatusOK, toPaymentResponse(payment))
}

func toPaymentResponse(p *domain.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		TransactionID: p.TransactionID,
		Amount:        p.Amount,
		Status:        p.Status,
	}
}
