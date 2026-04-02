package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"order-service/internal/domain"
	"order-service/internal/dto"
	"order-service/internal/middleware"
	"order-service/internal/usecase"
)

type Handler struct {
	uc *usecase.OrderUseCase
}

func NewHandler(uc *usecase.OrderUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	idempotencyKey, _ := c.Get(middleware.IdempotencyKeyCtx)
	key, _ := idempotencyKey.(string)

	order, err := h.uc.CreateOrder(c.Request.Context(), req.CustomerID, req.ItemName, req.Amount, key)
	if err != nil {
		if errors.Is(err, usecase.ErrPaymentServiceUnavailable) {
			c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: "payment service unavailable"})
			return
		}
		if errors.Is(err, domain.ErrInvalidAmount) || errors.Is(err, domain.ErrEmptyCustomerID) || errors.Is(err, domain.ErrEmptyItemName) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, toOrderResponse(order))
}

// GET /orders/:id.
func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.GetOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get order"})
		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}

// PATCH /orders/:id/cancel.
func (h *Handler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.CancelOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "order not found"})
			return
		}
		if errors.Is(err, domain.ErrCancelPaidOrder) || errors.Is(err, domain.ErrCancelFailedOrder) || errors.Is(err, domain.ErrAlreadyCancelled) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}

func toOrderResponse(o *domain.Order) dto.OrderResponse {
	return dto.OrderResponse{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		ItemName:   o.ItemName,
		Amount:     o.Amount,
		Status:     o.Status,
		CreatedAt:  o.CreatedAt,
	}
}
