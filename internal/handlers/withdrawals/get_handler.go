// Package withdrawals содержит обработчик истории списаний пользователя.
package withdrawals

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// GetHandler обрабатывает GET /api/user/withdrawals: возвращает списания
// пользователя от новых к старым. Коды: 200/204/401/500.
func GetHandler(store storage.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		userId, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Вы не авторизованы"))
			return
		}

		withdrawals, err := store.GetWithdrawals(r.Context(), userId)
		if err != nil {
			slog.Error("withdrawals: get withdrawals", "error", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		if len(withdrawals) == 0 {
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		data, err := json.Marshal(withdrawals)
		if err != nil {
			slog.Error("withdrawals: get withdrawals", "error", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}
}
