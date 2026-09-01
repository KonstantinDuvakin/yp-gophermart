// Package balance содержит обработчик получения баланса пользователя.
package balance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// GetHandler обрабатывает GET /api/user/balance: возвращает текущий баланс и
// сумму списаний пользователя. Коды: 200/401/500.
func GetHandler(store storage.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		userId, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Вы не авторизованы"))
			return
		}

		balance, err := store.GetBalance(r.Context(), userId)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				fmt.Printf("ошибка: %v", err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte("Внутренняя ошибка сервиса"))
				return
			}
			fmt.Printf("ошибка: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		data, err := json.Marshal(balance)
		if err != nil {
			fmt.Printf("ошибка: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}
}
