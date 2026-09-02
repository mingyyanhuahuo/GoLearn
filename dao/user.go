package dao

import (
	"day_4_1/model"
)

func CreateUser(user *model.User) error {
	return db.Create(user).Error
}
func GetUserById(id uint) (model.User, error) {
	var user model.User
	result := db.First(&user, id)
	return user, result.Error
}
func GetUserByUsername(username string) (model.User, error) {
	var user model.User
	result := db.Where("username = ?", username).First(&user)
	return user, result.Error
}
func GetUserByPhone(phone string) (model.User, error) {
	var user model.User
	result := db.Where("phone_number = ?", phone).First(&user)
	return user, result.Error
}
