package service

import (
	"day_4_1/dao"
	"day_4_1/model"
	"day_4_1/pkg/hashpassword"
	"day_4_1/pkg/jwtutil"
	"fmt"
)

func RegisterUser(username, password, phone string) (model.User, error) {
	_, err := dao.GetUserByUsername(username)
	if err == nil {
		return model.User{}, fmt.Errorf("用户名已存在")
	}
	passwordHash, err := hashpassword.RegisterHashPassword(password)
	if err != nil {
		return model.User{}, fmt.Errorf("密码哈希失败")
	}
	user := &model.User{
		Username:    username,
		PassHash:    passwordHash,
		PhoneNumber: phone,
		Role:        "user",
	}
	return *user, dao.CreateUser(user)
}
func Login(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", fmt.Errorf("用户名或密码不能为空")
	}
	user, err := dao.GetUserByUsername(username)
	if err != nil {
		return "", fmt.Errorf("用户或密码错误")
	}
	if err := hashpassword.LogincompareHashAndPassword(user.PassHash, password); err != nil {
		return "", fmt.Errorf("用户或密码错误")
	}
	token, err := jwtutil.GenerateToken(user.Id, user.Role)
	if err != nil {
		return "", fmt.Errorf("生成token失败")
	}
	return token, nil
}
