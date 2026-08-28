// Package register содержит обработчик регистрации пользователя.
package register

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// PostHandler обрабатывает POST /api/user/register: создаёт пользователя и
// выдаёт JWT в заголовке Authorization. Коды: 200/400/409/500.
func PostHandler(store storage.Storage, tm *auth.TokenManager) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		userDto := models.UserDto{}

		dec := json.NewDecoder(r.Body)

		if err := dec.Decode(&userDto); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Некорректный JSON"))
			return
		}

		if userDto.Password == "" || userDto.Login == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Поля не должны быть пустыми"))
			return
		}

		hashedPassword, err := auth.HashPassword(userDto.Password)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		userId, err := store.CreateUser(r.Context(), userDto.Login, hashedPassword)
		if err != nil {
			if errors.Is(err, storage.ErrLoginTaken) {
				rw.WriteHeader(http.StatusConflict)
				rw.Write([]byte(storage.ErrLoginTaken.Error()))
				return
			}
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		jwt, err := tm.BuildJWTString(userId)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		rw.Header().Set("Authorization", "Bearer "+jwt)
		rw.WriteHeader(http.StatusOK)
	}
}
