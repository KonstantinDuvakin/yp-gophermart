package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
)

// testStore подключается к БД из TEST_DATABASE_URI. Если переменная не задана,
// интеграционный тест пропускается (например, в окружении без PostgreSQL).
func testStore(t *testing.T) *PostgresStorage {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URI not set; skipping storage integration test")
	}
	st, err := NewPostgresStorage(context.Background(), dsn)
	if err != nil {
		t.Fatalf("NewPostgresStorage: %v", err)
	}
	t.Cleanup(st.Close)
	return st
}

func TestStorageFlow(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	uniq := time.Now().UnixNano()
	login := fmt.Sprintf("user_%d", uniq)
	order := fmt.Sprintf("%d", uniq)

	// --- пользователи ---
	id, err := st.CreateUser(ctx, login, "hash")
	if err != nil || id == 0 {
		t.Fatalf("CreateUser: id=%d err=%v", id, err)
	}
	if _, err := st.CreateUser(ctx, login, "hash"); !errors.Is(err, ErrLoginTaken) {
		t.Errorf("duplicate login: err=%v, want ErrLoginTaken", err)
	}
	if u, err := st.GetUserByLogin(ctx, login); err != nil || u.ID != id {
		t.Errorf("GetUserByLogin: u=%+v err=%v", u, err)
	}
	if _, err := st.GetUserByLogin(ctx, "missing_"+login); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("missing user: err=%v, want ErrUserNotFound", err)
	}

	// --- заказы ---
	if err := st.CreateOrder(ctx, id, order); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if err := st.CreateOrder(ctx, id, order); !errors.Is(err, ErrOrderOwnedByUser) {
		t.Errorf("re-upload same user: err=%v, want ErrOrderOwnedByUser", err)
	}
	id2, _ := st.CreateUser(ctx, "other_"+login, "hash")
	if err := st.CreateOrder(ctx, id2, order); !errors.Is(err, ErrOrderOwnedByOther) {
		t.Errorf("re-upload other user: err=%v, want ErrOrderOwnedByOther", err)
	}
	if orders, err := st.GetOrders(ctx, id); err != nil || len(orders) != 1 {
		t.Errorf("GetOrders: n=%d err=%v", len(orders), err)
	}
	if nums, err := st.GetUnprocessedOrders(ctx); err != nil || !contains(nums, order) {
		t.Errorf("GetUnprocessedOrders: %v err=%v", nums, err)
	}

	// --- начисление и баланс ---
	acc := 100.0
	if err := st.UpdateOrderAccrual(ctx, order, models.OrderStatusProcessed, &acc); err != nil {
		t.Fatalf("UpdateOrderAccrual: %v", err)
	}
	if bal, err := st.GetBalance(ctx, id); err != nil || bal.Current != 100 {
		t.Errorf("GetBalance after accrual: %+v err=%v", bal, err)
	}

	// --- списание ---
	if err := st.Withdraw(ctx, id, order, 30); err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
	if err := st.Withdraw(ctx, id, order, 1000); !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("over-withdraw: err=%v, want ErrInsufficientFunds", err)
	}
	if bal, err := st.GetBalance(ctx, id); err != nil || bal.Current != 70 || bal.Withdrawn != 30 {
		t.Errorf("GetBalance after withdraw: %+v err=%v", bal, err)
	}
	if ws, err := st.GetWithdrawals(ctx, id); err != nil || len(ws) != 1 || ws[0].Sum != 30 {
		t.Errorf("GetWithdrawals: %+v err=%v", ws, err)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
