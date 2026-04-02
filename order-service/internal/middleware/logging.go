package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		log.Printf("[%s] %s %s | %d | %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Proto,
			c.Writer.Status(),
			latency,
		)
	}
}
