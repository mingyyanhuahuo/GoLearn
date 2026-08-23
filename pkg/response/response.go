package response

// 参考 gin 的状态码找ai问出来的
//     StatusOK                   = 200  // "200 OK"
//     StatusBadRequest           = 400  // "400 Bad Request"
//     StatusUnauthorized         = 401  // "401 Unauthorized"  ← 你现在这个
//     StatusForbidden            = 403  // "403 Forbidden"
//     StatusNotFound             = 404  // "404 Not Found"
//     StatusInternalServerError  = 500  // "500 Internal Server Error"
import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data})
}
func Err(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"code": code,
		"msg":  msg,
		"data": nil,
	})
}
