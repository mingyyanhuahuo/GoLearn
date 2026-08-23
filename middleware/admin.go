package middleware

import (
	"day_4_1/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Err(c, http.StatusForbidden, "权限不足，管理员才能访问")
			c.Abort()
			return
		}
		c.Next()
	}

}
