package response

// 参考 gin 的状态码找ai问出来的
//     StatusOK                   = 200  // "200 OK"
//     StatusBadRequest           = 400  // "400 Bad Request"
//     StatusUnauthorized         = 401  // "401 Unauthorized"  ← 你现在这个
//     StatusForbidden            = 403  // "403 Forbidden"
//     StatusNotFound             = 404  // "404 Not Found"
//     StatusInternalServerError  = 500  // "500 Internal Server Error"
import (
	"day_4_1/pkg/errcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

type resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, resp{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, resp{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}
func Err(c *gin.Context, bizErr *errcode.BizError) {
	c.JSON(bizErr.HttpStatus, resp{
		Code: bizErr.Code,
		Msg:  bizErr.Message,
		Data: nil,
	})
}
