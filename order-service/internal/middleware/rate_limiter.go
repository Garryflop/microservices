package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter returns a Gin middleware that limits requests per client IP
// using Redis INCR + EXPIRE (fixed-window counter pattern).
func RateLimiter(redisClient *redis.Client, maxRequests int, windowSeconds int) gin.HandlerFunc {
	window := time.Duration(windowSeconds) * time.Second

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate:%s", clientIP)

		ctx := c.Request.Context()

		// increment counter
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("[RateLimit] Redis error: %v, allowing request", err)
			c.Next()
			return
		}

		// set expiry on first request in window
		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		// get TTL for headers
		ttl, _ := redisClient.TTL(ctx, key).Result()
		remaining := int64(maxRequests) - count
		if remaining < 0 {
			remaining = 0
		}

		// set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(int64(ttl.Seconds()), 10))

		if count > int64(maxRequests) {
			log.Printf("[RateLimit] Client %s exceeded limit (%d/%d)", clientIP, count, maxRequests)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, try again later",
			})
			return
		}

		c.Next()
	}
}
