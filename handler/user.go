package handler

import (
	"day_4_1/pkg/response"
	"day_4_1/service"

	"day_4_1/model"

	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	AccessToken string     `json:"access_token"`
	Tokentype   string     `json:"token_type"`
	ExpiresIn   int64      `json:"expires_in"`
	User        model.User `json:"user"`
}
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=1,max=32"`
	Name     string `json:"name" binding:"required,min=1,max=32"`
	Password string `json:"password" binding:"required,min=8,max=32"`
	Phone    string `json:"phone" binding:"required,len=11"`
}

// type RegisterResp struct {
// 	Id       uint   `json:"id"`
// 	Username string `json:"username"`
// 	Role     string `json:"role"`
// 	Phone    string `json:"phone"`
// }

func Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BindError(c, err)
		return
	}
	user, err := service.RegisterUser(req.Username, req.Name, req.Password, req.Phone)
	if err != nil {
		c.Error(err)
		return
	}
	response.Created(c, user)
}
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BindError(c, err)
		return
	}
	user, token, err := service.Login(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, LoginResp{
		AccessToken: token,
		Tokentype:   "Bearer",
		ExpiresIn:   7200,
		User:        user,
	})
}
