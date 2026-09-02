package service

import (
	"day_4_1/dao"
	"day_4_1/model"
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/hashpassword"
	"day_4_1/pkg/jwtutil"
	"strings"
)

func RegisterUser(username, name, password, phone string) (model.User, error) {
	if username == "" || password == "" || phone == "" || name == "" {
		return model.User{}, errcode.ErrBadRequest
	}
	_, err := dao.GetUserByUsername(username)
	if err == nil {
		return model.User{}, errcode.ErrExistingUser
	}
	if _, err := dao.GetUserByPhone(phone); err == nil {
		return model.User{}, errcode.ErrExistingUser
	}
	passwordHash, err := hashpassword.RegisterHashPassword(password)
	if err != nil {
		return model.User{}, errcode.ErrInternalServerError
	}
	user := &model.User{
		Name:        name,
		Username:    username,
		PassHash:    passwordHash,
		PhoneNumber: phone,
		Role:        "student",
	}
	return *user, dao.CreateUser(user)
}
func Login(username, password string) (model.User, string, error) {
	if username == "" || password == "" {
		return model.User{}, "", errcode.ErrBadRequest
	}
	user, err := dao.GetUserByUsername(username)
	if err != nil {
		return model.User{}, "", errcode.ErrWrongCredentials
	}
	if err := hashpassword.LogincompareHashAndPassword(user.PassHash, password); err != nil {
		return model.User{}, "", errcode.ErrWrongCredentials
	}
	token, err := jwtutil.GenerateToken(user.Id, user.Role)
	if err != nil {
		return model.User{}, "", errcode.ErrInternalServerError
	}
	return user, token, nil
}
func IsAdmin(role string) bool {
	return strings.EqualFold(role, "admin")
}
