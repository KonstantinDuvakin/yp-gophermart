package accrual

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
)

// Worker периодически опрашивает систему расчёта начислений по незавершённым
// заказам и обновляет их статусы и баланс пользователей.
type Worker struct {
	store    storage.Storage
	client   *Client
	interval time.Duration
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
		log.Printf("worker: can't get unprocessed orders. Error: %v", err)
		return
	}

	for _, num := range orderNums {
		accrualRes, td, err := w.client.GetOrderAccrual(ctx, num)
		switch {
		case errors.Is(err, ErrOrderNotRegistered):
			continue
		case errors.Is(err, ErrToManyRequests):
			time.Sleep(td)
			return
		case err != nil:
			log.Printf("worker: order %s: %v", num, err)
			continue
		}

		status, accrual := mapAccrualStatus(accrualRes)

		err = w.store.UpdateOrderAccrual(ctx, num, status, accrual)
		if err != nil {
			log.Printf("worker: fail update order %s: %v", num, err)
		}
	}
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
