package middleware

import (
	"day_4_1/pkg/jwtutil"
	"day_4_1/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			response.Err(c, http.StatusUnauthorized, "无效的授权头")
			c.Abort()
			return
		}
		tokenString := authHeader[7:]
		claims, err := jwtutil.ParseToken(tokenString)
		if err != nil {
			response.Err(c, http.StatusUnauthorized, "无效的token")
			c.Abort()
			return
		}
		c.Set("id", claims.UserId)
		c.Set("role", claims.Role)
		c.Next()
	}
}
