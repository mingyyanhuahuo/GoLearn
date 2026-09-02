package middleware

import (
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/logger"
	"day_4_1/pkg/response"
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
			response.Err(c, bizErr)
			return
		}
		logger.Logger.Error("未知错误", zap.Error(err))
		response.Err(c, errcode.ErrInternalServerError)
	}
}
