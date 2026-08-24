package middleware

import (
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/response"
	"errors"

	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		errs := c.Errors
		if len(errs) == 0 {
			return
		}
		err := errs.Last().Err
		var bizErr *errcode.BizError
		if ok := errors.As(err, &bizErr); ok {
			response.Err(c, bizErr.HttpStatus, bizErr.Message)
			return
		}
		response.Err(c, 500, "服务器内部错误")
	}
}
