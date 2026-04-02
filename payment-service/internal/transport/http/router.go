package http

import (
	"github.com/gin-gonic/gin"

	"payment-service/internal/middleware"
)

func NewRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// middleware stack
	r.Use(middleware.Recovery())
	r.Use(middleware.Logging())

	// Payment routes
	r.POST("/payments", handler.ProcessPayment)
	r.GET("/payments/:order_id", handler.GetPayment)

	return r
}
