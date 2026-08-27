// Package models описывает доменные сущности накопительной системы
// лояльности «Гофермарт»: пользователей, заказы, списания и баланс.
package models

import "time"

// Статусы обработки заказа во внутренней системе.
const (
	// OrderStatusNew — заказ загружен, но ещё не отправлен в расчёт.
	OrderStatusNew = "NEW"
	// OrderStatusProcessing — начисление за заказ рассчитывается.
	OrderStatusProcessing = "PROCESSING"
	// OrderStatusInvalid — система расчёта отказала в начислении (финальный).
	OrderStatusInvalid = "INVALID"
	// OrderStatusProcessed — начисление рассчитано и получено (финальный).
	OrderStatusProcessed = "PROCESSED"
)

// UserDto — тело запроса регистрации и аутентификации (логин/пароль).
type UserDto struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// WithdrawalDto — тело запроса на списание баллов: номер заказа и сумма.
type WithdrawalDto struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// User — зарегистрированный пользователь системы.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

// Order — загруженный пользователем номер заказа и его статус расчёта.
type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Balance — текущий баланс баллов и сумма списаний пользователя.
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Withdrawal — факт списания баллов в счёт оплаты заказа.
type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
