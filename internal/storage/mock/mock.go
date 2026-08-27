// Package mock содержит настраиваемую заглушку storage.Storage для юнит-тестов.
package mock

import (
	"context"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// Storage — заглушка storage.Storage. Каждый метод делегирует одноимённому
// полю-функции, если оно задано; иначе возвращает нулевые значения.
type Storage struct {
	CreateUserFn           func(ctx context.Context, login, passwordHash string) (int64, error)
	GetUserByLoginFn       func(ctx context.Context, login string) (models.User, error)
	CreateOrderFn          func(ctx context.Context, userID int64, number string) error
	GetOrdersFn            func(ctx context.Context, userID int64) ([]models.Order, error)
	GetUnprocessedOrdersFn func(ctx context.Context) ([]string, error)
	UpdateOrderAccrualFn   func(ctx context.Context, number, status string, accrual *float64) error
	GetBalanceFn           func(ctx context.Context, userID int64) (models.Balance, error)
	WithdrawFn             func(ctx context.Context, userID int64, order string, sum float64) error
	GetWithdrawalsFn       func(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

// Проверка, что заглушка реализует интерфейс на этапе компиляции.
var _ storage.Storage = (*Storage)(nil)

// CreateUser вызывает CreateUserFn, если задана.
func (m *Storage) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, login, passwordHash)
	}
	return 0, nil
}

// GetUserByLogin вызывает GetUserByLoginFn, если задана.
func (m *Storage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	if m.GetUserByLoginFn != nil {
		return m.GetUserByLoginFn(ctx, login)
	}
	return models.User{}, nil
}

// CreateOrder вызывает CreateOrderFn, если задана.
func (m *Storage) CreateOrder(ctx context.Context, userID int64, number string) error {
	if m.CreateOrderFn != nil {
		return m.CreateOrderFn(ctx, userID, number)
	}
	return nil
}

// GetOrders вызывает GetOrdersFn, если задана.
func (m *Storage) GetOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	if m.GetOrdersFn != nil {
		return m.GetOrdersFn(ctx, userID)
	}
	return nil, nil
}

// GetUnprocessedOrders вызывает GetUnprocessedOrdersFn, если задана.
func (m *Storage) GetUnprocessedOrders(ctx context.Context) ([]string, error) {
	if m.GetUnprocessedOrdersFn != nil {
		return m.GetUnprocessedOrdersFn(ctx)
	}
	return nil, nil
}

// UpdateOrderAccrual вызывает UpdateOrderAccrualFn, если задана.
func (m *Storage) UpdateOrderAccrual(ctx context.Context, number, status string, accrual *float64) error {
	if m.UpdateOrderAccrualFn != nil {
		return m.UpdateOrderAccrualFn(ctx, number, status, accrual)
	}
	return nil
}

// GetBalance вызывает GetBalanceFn, если задана.
func (m *Storage) GetBalance(ctx context.Context, userID int64) (models.Balance, error) {
	if m.GetBalanceFn != nil {
		return m.GetBalanceFn(ctx, userID)
	}
	return models.Balance{}, nil
}

// Withdraw вызывает WithdrawFn, если задана.
func (m *Storage) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	if m.WithdrawFn != nil {
		return m.WithdrawFn(ctx, userID, order, sum)
	}
	return nil
}

// GetWithdrawals вызывает GetWithdrawalsFn, если задана.
func (m *Storage) GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	if m.GetWithdrawalsFn != nil {
		return m.GetWithdrawalsFn(ctx, userID)
	}
	return nil, nil
}

// Close — заглушка, ничего не делает.
func (m *Storage) Close() {}
