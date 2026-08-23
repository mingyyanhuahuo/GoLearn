package hashpassword

import (
	"golang.org/x/crypto/bcrypt"
)

func LogincompareHashAndPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
func RegisterHashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.
		GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}
