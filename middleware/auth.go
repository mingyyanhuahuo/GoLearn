package middleware

import (
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/jwtutil"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			c.Error(errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		tokenString := authHeader[7:]
		claims, err := jwtutil.ParseToken(tokenString)
		if err != nil {
			c.Error(errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set("id", claims.UserId)
		c.Set("role", claims.Role)
		c.Next()
	}
}
