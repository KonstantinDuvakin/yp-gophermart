// Package withdraw содержит обработчик списания баллов в счёт оплаты заказа.
package withdraw

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/service/luhn"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// PostHandler обрабатывает POST /api/user/balance/withdraw: списывает баллы в
// счёт указанного заказа. Коды: 200/401/402/422/500.
func PostHandler(store storage.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		userId, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Вы не авторизованы"))
			return
		}

		withdrawDto := &models.WithdrawalDto{}

		dec := json.NewDecoder(r.Body)

		if err := dec.Decode(withdrawDto); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Некорректный JSON"))
			return
		}

		if withdrawDto.Sum <= 0 || withdrawDto.Order == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Поля не должны быть пустыми"))
			return
		}

		isOrderNumCorrect := luhn.CheckNumber(withdrawDto.Order)
		if !isOrderNumCorrect {
			rw.WriteHeader(http.StatusUnprocessableEntity)
			rw.Write([]byte("Неверный номер заказа"))
			return
		}

		err := store.Withdraw(r.Context(), userId, withdrawDto.Order, withdrawDto.Sum)
		if err != nil {
			if errors.Is(err, storage.ErrInsufficientFunds) {
				rw.WriteHeader(http.StatusPaymentRequired)
				rw.Write([]byte("Недостаточно средств"))
				return
			}
			fmt.Printf("ошибка: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Внутренняя ошибка сервиса"))
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}
