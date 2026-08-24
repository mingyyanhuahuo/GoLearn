package service

import (
	"day_4_1/dao"
	"day_4_1/model"
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/hashpassword"
	"day_4_1/pkg/jwtutil"
)

func RegisterUser(username, password, phone string) (model.User, error) {
	_, err := dao.GetUserByUsername(username)
	if err == nil {
		return model.User{}, errcode.ErrExistingUser
	}
	passwordHash, err := hashpassword.RegisterHashPassword(password)
	if err != nil {
		return model.User{}, errcode.ErrInternalServerError
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
		return "", errcode.ErrBadRequest
	}
	user, err := dao.GetUserByUsername(username)
	if err != nil {
		return "", errcode.ErrBadRequest
	}
	if err := hashpassword.LogincompareHashAndPassword(user.PassHash, password); err != nil {
		return "", errcode.ErrBadRequest
	}
	token, err := jwtutil.GenerateToken(user.Id, user.Role)
	if err != nil {
		return "", errcode.ErrInternalServerError
	}
	return token, nil
}
