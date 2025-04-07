package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

func RateLimit(redis *redis.Client, requests int, duration time.Duration) gin.HandlerFunc {

	return func(c *gin.Context) {
		key := "ratelimit:" + c.ClientIP()
		ctx := context.Background()

		count, err := redis.Incr(ctx, key).Result()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": " rate limit error"})

			c.Abort()
			return
		}

		if count == 1 {
			redis.Expire(ctx, key, duration)
		}

		if count > int64(requests) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}

}
