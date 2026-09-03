// Package storage описывает контракт хранилища данных «Гофермарт» и его
// реализацию поверх PostgreSQL.
package storage

import (
	"context"
	"errors"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
)

// Ошибки хранилища, на которые опирается бизнес-логика хендлеров.
var (
	// ErrLoginTaken возвращается, когда логин уже занят другим пользователем.
	ErrLoginTaken = errors.New("login already taken")
	// ErrUserNotFound возвращается, когда пользователь не найден.
	ErrUserNotFound = errors.New("user not found")
	// ErrOrderOwnedByUser — заказ уже был загружен этим же пользователем.
	ErrOrderOwnedByUser = errors.New("order already uploaded by this user")
	// ErrOrderOwnedByOther — заказ уже был загружен другим пользователем.
	ErrOrderOwnedByOther = errors.New("order already uploaded by another user")
	// ErrInsufficientFunds — на счету недостаточно баллов для списания.
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// Storage — абстракция над хранилищем данных сервиса.
//
//go:generate mockgen -source=storage.go -destination=mock/mock.go -package=mock
type Storage interface {
	// CreateUser создаёт пользователя и возвращает его идентификатор.
	// Если логин занят, возвращает ErrLoginTaken.
	CreateUser(ctx context.Context, login, passwordHash string) (int64, error)
	// GetUserByLogin возвращает пользователя по логину или ErrUserNotFound.
	GetUserByLogin(ctx context.Context, login string) (models.User, error)

	// CreateOrder привязывает номер заказа к пользователю. Возвращает
	// ErrOrderOwnedByUser или ErrOrderOwnedByOther при повторной загрузке.
	CreateOrder(ctx context.Context, userID int64, number string) error
	// GetOrders возвращает заказы пользователя, отсортированные от новых к старым.
	GetOrders(ctx context.Context, userID int64) ([]models.Order, error)
	// GetUnprocessedOrders возвращает номера заказов в статусах NEW/PROCESSING
	// для дообработки фоновым воркером.
	GetUnprocessedOrders(ctx context.Context) ([]string, error)
	// UpdateOrderAccrual обновляет статус заказа и, при начислении, атомарно
	// пополняет баланс пользователя-владельца.
	UpdateOrderAccrual(ctx context.Context, number, status string, accrual *float64) error

	// GetBalance возвращает текущий баланс и сумму списаний пользователя.
	GetBalance(ctx context.Context, userID int64) (models.Balance, error)
	// Withdraw списывает баллы в счёт заказа. Возвращает ErrInsufficientFunds,
	// если баллов не хватает. Операция атомарна.
	Withdraw(ctx context.Context, userID int64, order string, sum float64) error
	// GetWithdrawals возвращает списания пользователя от новых к старым.
	GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error)

	// Close освобождает ресурсы хранилища.
	Close()
}
