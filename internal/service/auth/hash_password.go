// Package auth предоставляет хеширование паролей (bcrypt) и работу с JWT.
package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword возвращает bcrypt-хеш пароля.
func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

// CheckPasswordHash сообщает, соответствует ли пароль хешу. Несовпадение —
// это (false, nil); ненулевая ошибка означает реальный сбой (например, битый хеш).
func CheckPasswordHash(hashedPassword, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
