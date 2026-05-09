package http

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"order-service/internal/middleware"
)

func NewRouter(handler *Handler, redisClient *redis.Client, rateLimitMax, rateLimitWindow int) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	//middleware stack
	r.Use(middleware.Recovery())
	r.Use(middleware.Logging())
	r.Use(middleware.RateLimiter(redisClient, rateLimitMax, rateLimitWindow))
	r.Use(middleware.IdempotencyKey())

	// Order routes
	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)

	return r
}

