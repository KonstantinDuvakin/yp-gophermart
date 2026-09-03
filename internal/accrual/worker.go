package accrual

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// Worker периодически опрашивает систему расчёта начислений по незавершённым
// заказам и обновляет их статусы и баланс пользователей.
type Worker struct {
	store      storage.Storage
	client     *Client
	interval   time.Duration
	pauseUntil atomic.Int64
}

// NewWorker создаёт воркер, опрашивающий accrual раз в секунду.
func NewWorker(store storage.Storage, client *Client) *Worker {
	return &Worker{store: store, client: client, interval: time.Second}
}

// Run запускает цикл опроса и работает до отмены ctx.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	orderNums, err := w.store.GetUnprocessedOrders(ctx)
	if err != nil {
		slog.Error("worker: get unprocessed orders", "error", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

Orders:
	for _, num := range orderNums {
		select {
		case <-ctx.Done():
			break Orders
		case sem <- struct{}{}:
			wg.Add(1)
			go func(num string) {
				defer wg.Done()
				defer func() { <-sem }()
				err := w.processOne(ctx, num)
				if err != nil {
					slog.Error("worker: process order", "error", err)
				}
			}(num)
		}
	}

	wg.Wait()
}

func (w *Worker) processOne(ctx context.Context, num string) error {
	if until := w.pauseUntil.Load(); until > 0 {
		if wait := time.Until(time.Unix(0, until)); wait > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}

	accrualRes, td, err := w.client.GetOrderAccrual(ctx, num)
	switch {
	case errors.Is(err, ErrOrderNotRegistered):
		return ErrOrderNotRegistered
	case errors.Is(err, ErrTooManyRequests):
		w.pauseUntil.Store(time.Now().Add(td).UnixNano())
		return ErrTooManyRequests
	case err != nil:
		return fmt.Errorf("worker: order %s: %w", num, err)
	}

	status, accrual := mapAccrualStatus(accrualRes)

	err = w.store.UpdateOrderAccrual(ctx, num, status, accrual)
	if err != nil {
		return fmt.Errorf("worker: fail update order %s: %w", num, err)
	}

	return nil
}

func mapAccrualStatus(dto *Dto) (string, *float64) {
	switch dto.Status {
	case StatusInvalid:
		return models.OrderStatusInvalid, nil
	case StatusProcessed:
		return models.OrderStatusProcessed, dto.Accrual
	default:
		return models.OrderStatusProcessing, nil
	}
}
