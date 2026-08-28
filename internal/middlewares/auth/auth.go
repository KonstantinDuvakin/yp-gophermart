// Package auth содержит HTTP-middleware проверки JWT и доступ к userID из контекста.
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
)

type ctxKey int

const userIdKey ctxKey = 0

// Middleware возвращает middleware, которое пропускает только запросы с валидным
// Bearer-токеном и кладёт userID в контекст; иначе отвечает 401.
func Middleware(tm *auth.TokenManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			bearerToken := r.Header.Get("Authorization")
			if bearerToken == "" {
				rw.WriteHeader(http.StatusUnauthorized)
				rw.Write([]byte("Вы не авторизованы"))
				return
			}

			token, ok := strings.CutPrefix(bearerToken, "Bearer ")
			if !ok {
				rw.WriteHeader(http.StatusUnauthorized)
				rw.Write([]byte("Вы не авторизованы"))
				return
			}

			claims, err := tm.ParseJWTString(token)
			if err != nil {
				rw.WriteHeader(http.StatusUnauthorized)
				rw.Write([]byte("Вы не авторизованы"))
				return
			}

			ctx := withIdContext(r.Context(), claims.UserId)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

func withIdContext(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIdKey, id)
}

// UserIdFromContext возвращает userID, положенный Middleware, и признак наличия.
func UserIdFromContext(ctx context.Context) (int64, bool) {
	userId, ok := ctx.Value(userIdKey).(int64)
	return userId, ok
}
