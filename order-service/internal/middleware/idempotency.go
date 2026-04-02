package middleware

import (
	"github.com/gin-gonic/gin"
)

const IdempotencyKeyCtx = "idempotency_key"

// extracts the IdempotencyKey header and stores it in the Gin context
func IdempotencyKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key != "" {
			c.Set(IdempotencyKeyCtx, key)
		}
		c.Next()
	}
}
