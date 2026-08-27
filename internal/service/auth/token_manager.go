package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — полезная нагрузка JWT: стандартные поля и идентификатор пользователя.
type Claims struct {
	jwt.RegisteredClaims
	UserId int64
}

// TokenManager выпускает и проверяет JWT, храня секрет и срок жизни токена.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager создаёт менеджер токенов с заданным секретом и TTL 24 часа.
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: time.Hour * 24}
}

// BuildJWTString выпускает подписанный токен для пользователя userId.
func (tm *TokenManager) BuildJWTString(userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.ttl)),
		},
		UserId: userId,
	})

	tokenString, err := token.SignedString(tm.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWTString проверяет подпись и срок токена и возвращает его claims.
func (tm *TokenManager) ParseJWTString(tokenString string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tm.secret, nil
	})

	if err != nil {
		return claims, err
	}

	return claims, nil
}
