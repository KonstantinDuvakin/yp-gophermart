// Package orders содержит обработчики загрузки и получения заказов пользователя.
package orders

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/service/luhn"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// PostHandler обрабатывает POST /api/user/orders: принимает номер заказа,
// проверяет его по Луну и привязывает к пользователю. Коды: 202/200/400/409/422/500.
func PostHandler(store storage.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		userId, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Вы не авторизованы"))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("ошибка: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		orderNum := strings.TrimSpace(string(body))

		if orderNum == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Отсутствует номер заказа"))
			return
		}

		isOrderCorrect := luhn.CheckNumber(orderNum)
		if !isOrderCorrect {
			rw.WriteHeader(http.StatusUnprocessableEntity)
			rw.Write([]byte("Номер заказа неверный"))
			return
		}

		err = store.CreateOrder(r.Context(), userId, orderNum)
		switch {
		case err == nil:
			rw.WriteHeader(http.StatusAccepted)
		case errors.Is(err, storage.ErrOrderOwnedByUser):
			rw.WriteHeader(http.StatusOK)
		case errors.Is(err, storage.ErrOrderOwnedByOther):
			rw.WriteHeader(http.StatusConflict)
		default:
			fmt.Printf("ошибка: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}
