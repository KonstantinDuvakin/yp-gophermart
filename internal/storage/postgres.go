package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PostgresStorage — реализация Storage поверх пула соединений pgx.
type PostgresStorage struct {
	pool *pgxpool.Pool
}

// NewPostgresStorage открывает пул соединений по строке dsn, применяет
// схему БД и возвращает готовое хранилище.
func NewPostgresStorage(ctx context.Context, dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err = migrations.RunMigrations(sqlDB); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migration error: %w", err)
	}

	return &PostgresStorage{pool: pool}, nil
}

// Close закрывает пул соединений.
func (s *PostgresStorage) Close() {
	s.pool.Close()
}

// CreateUser создаёт пользователя. При конфликте по логину возвращает ErrLoginTaken.
func (s *PostgresStorage) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`,
		login, passwordHash,
	).Scan(&id)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" { // unique_violation
				return 0, ErrLoginTaken
			}
		}
		return 0, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

// GetUserByLogin возвращает пользователя по логину или ErrUserNotFound.
func (s *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, login, password_hash FROM users WHERE login = $1`, login,
	).Scan(&u.ID, &u.Login, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("get user by login: %w", err)
	}
	return u, nil
}

// CreateOrder привязывает заказ к пользователю, различая повторную загрузку
// тем же и другим пользователем.
func (s *PostgresStorage) CreateOrder(ctx context.Context, userID int64, number string) error {
	tag, err := s.pool.Exec(ctx,
		`INSERT INTO orders (number, user_id) VALUES ($1, $2) ON CONFLICT (number) DO NOTHING`,
		number, userID,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	// Заказ уже существует — выясняем владельца.
	var ownerID int64
	if err = s.pool.QueryRow(ctx,
		`SELECT user_id FROM orders WHERE number = $1`, number,
	).Scan(&ownerID); err != nil {
		return fmt.Errorf("check order owner: %w", err)
	}
	if ownerID == userID {
		return ErrOrderOwnedByUser
	}
	return ErrOrderOwnedByOther
}

// GetOrders возвращает заказы пользователя от новых к старым.
func (s *PostgresStorage) GetOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT number, status, accrual, uploaded_at
		   FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var o models.Order
		if err = rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// GetUnprocessedOrders возвращает номера заказов в статусах NEW/PROCESSING.
func (s *PostgresStorage) GetUnprocessedOrders(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT number FROM orders WHERE status IN ($1, $2) ORDER BY uploaded_at`,
		models.OrderStatusNew, models.OrderStatusProcessing,
	)
	if err != nil {
		return nil, fmt.Errorf("query unprocessed orders: %w", err)
	}
	defer rows.Close()

	numbers := make([]string, 0)
	for rows.Next() {
		var n string
		if err = rows.Scan(&n); err != nil {
			return nil, fmt.Errorf("scan order number: %w", err)
		}
		numbers = append(numbers, n)
	}
	return numbers, rows.Err()
}

// UpdateOrderAccrual обновляет статус заказа и атомарно пополняет баланс
// владельца, если начисление положительно.
func (s *PostgresStorage) UpdateOrderAccrual(ctx context.Context, number, status string, accrual *float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx,
		`UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 RETURNING user_id`,
		status, accrual, number,
	).Scan(&userID)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	if accrual != nil && *accrual > 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET current_balance = current_balance + $1 WHERE id = $2`,
			*accrual, userID,
		); err != nil {
			return fmt.Errorf("credit balance: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetBalance возвращает текущий баланс и сумму списаний пользователя.
func (s *PostgresStorage) GetBalance(ctx context.Context, userID int64) (models.Balance, error) {
	var b models.Balance
	err := s.pool.QueryRow(ctx,
		`SELECT current_balance, withdrawn FROM users WHERE id = $1`, userID,
	).Scan(&b.Current, &b.Withdrawn)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Balance{}, ErrUserNotFound
	}
	if err != nil {
		return models.Balance{}, fmt.Errorf("get balance: %w", err)
	}
	return b, nil
}

// Withdraw атомарно списывает баллы: блокирует строку пользователя, проверяет
// достаточность средств, обновляет баланс и записывает факт списания.
func (s *PostgresStorage) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var current float64
	err = tx.QueryRow(ctx,
		`SELECT current_balance FROM users WHERE id = $1 FOR UPDATE`, userID,
	).Scan(&current)
	if err != nil {
		return fmt.Errorf("lock user balance: %w", err)
	}
	if current < sum {
		return ErrInsufficientFunds
	}

	if _, err := tx.Exec(ctx,
		`UPDATE users SET current_balance = current_balance - $1, withdrawn = withdrawn + $1 WHERE id = $2`,
		sum, userID,
	); err != nil {
		return fmt.Errorf("update balance: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO withdrawals (order_number, user_id, sum) VALUES ($1, $2, $3)`,
		order, userID, sum,
	); err != nil {
		return fmt.Errorf("insert withdrawal: %w", err)
	}

	return tx.Commit(ctx)
}

// GetWithdrawals возвращает списания пользователя от новых к старым.
func (s *PostgresStorage) GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT order_number, sum, processed_at
		   FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query withdrawals: %w", err)
	}
	defer rows.Close()

	ws := make([]models.Withdrawal, 0)
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		ws = append(ws, w)
	}
	return ws, rows.Err()
}
