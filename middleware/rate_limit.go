package middleware

import (
	"context"
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/redisdb"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("id")
		if !exists {
			c.Error(errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		uid := userID.(uint)
		key := fmt.Sprintf("rate_limit:%d", uid)
		ctx := context.Background()
		n, err := redisdb.Rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if n == 1 {
			redisdb.Rdb.Expire(ctx, key, 2*time.Second)
		}
		if n > int64(10) {
			c.Error(errcode.New(429, 10010, "请求过于频繁，请稍后再试"))
			c.Abort()
			return
		}
		c.Next()
	}
}
