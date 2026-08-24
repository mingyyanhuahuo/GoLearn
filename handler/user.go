package handler

import (
	"day_4_1/pkg/response"
	"day_4_1/service"

	"day_4_1/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type LoginResp struct {
	Token     string `json:"token"`
	Tokentype string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"`
}
type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required,len=11"`
}
type RegisterResp struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Phone    string `json:"phone"`
}

func Register(c *gin.Context) {
	// 调用 service 层的注册函数
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errcode.New(400, 10000, "请求参数错误: "+err.Error()))
		return
	}
	user, err := service.RegisterUser(req.Username, req.Password, req.Phone)
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, RegisterResp{
		Id:       user.Id,
		Username: user.Username,
		Phone:    user.PhoneNumber,
		Role:     user.Role,
	})
}
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errcode.New(400, 10000, "请求参数错误: "+err.Error()))
		return
	}
	resp, err := service.Login(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, LoginResp{
		Token:     resp,
		Tokentype: "Bearer",
		ExpiresIn: 7200, // 假设 token 有效期为 2 小时
	})
}
