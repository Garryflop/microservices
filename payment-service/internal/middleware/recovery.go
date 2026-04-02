package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery returns a Gin middleware that recovers from panics and returns a 500 JSON error.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
