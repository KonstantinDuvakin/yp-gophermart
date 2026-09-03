// Package login содержит обработчик аутентификации пользователя.
package login

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// PostHandler обрабатывает POST /api/user/login: проверяет пару логин/пароль и
// выдаёт JWT в заголовке Authorization. Коды: 200/400/401/500.
func PostHandler(store storage.Storage, tm *auth.TokenManager) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		userDto := models.UserDto{}

		dec := json.NewDecoder(r.Body)

		if err := dec.Decode(&userDto); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Некорректный JSON"))
			return
		}

		if userDto.Login == "" || userDto.Password == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Поля не должны быть пустыми"))
			return
		}

		user, err := store.GetUserByLogin(r.Context(), userDto.Login)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				rw.WriteHeader(http.StatusUnauthorized)
				rw.Write([]byte("Пользователя с таким логином не существует"))
				return
			}
			slog.Error("login: get user", "error", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		isCorrectPassword, err := auth.CheckPasswordHash(user.PasswordHash, userDto.Password)
		if err != nil {
			slog.Error("login: check password", "error", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		if !isCorrectPassword {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Неверный пароль"))
			return
		}

		jwt, err := tm.BuildJWTString(user.ID)
		if err != nil {
			slog.Error("login: build token", "error", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		rw.Header().Set("Authorization", "Bearer "+jwt)
		rw.WriteHeader(http.StatusOK)
	}
}
