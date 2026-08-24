package middleware

import (
	"day_4_1/pkg/errcode"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.Error(errcode.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}

}
