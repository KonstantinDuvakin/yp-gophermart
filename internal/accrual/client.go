// Package accrual реализует клиента и фонового воркера для взаимодействия
// с внешней системой расчёта начислений баллов лояльности.
package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Dto — ответ системы расчёта начислений по конкретному заказу.
type Dto struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

var (
	// ErrOrderNotRegistered возвращается, когда accrual ещё не знает о заказе (HTTP 204).
	ErrOrderNotRegistered = errors.New("заказ не зарегистрирован")
	// ErrToManyRequests возвращается при превышении лимита запросов к accrual (HTTP 429).
	ErrToManyRequests = errors.New("слишком много запросов")
)

const (
	// StatusInvalid — accrual отказал в расчёте (финальный статус).
	StatusInvalid = "INVALID"
	// StatusProcessed — расчёт начисления завершён (финальный статус).
	StatusProcessed = "PROCESSED"
)

// Client — HTTP-клиент к системе расчёта начислений.
type Client struct {
	url  string
	http *http.Client
}

// NewClient создаёт клиента, обращающегося к accrual по базовому адресу url.
func NewClient(url string) *Client {
	return &Client{url: url, http: &http.Client{Timeout: time.Second * 5}}
}

// GetOrderAccrual запрашивает статус расчёта по номеру заказа. Возвращает
// ErrOrderNotRegistered при 204 и ErrToManyRequests (со временем ожидания
// из Retry-After) при 429.
func (c *Client) GetOrderAccrual(ctx context.Context, number string) (*Dto, time.Duration, error) {
	url := c.url + "/api/orders/" + number

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		dto := &Dto{}
		if err = json.NewDecoder(resp.Body).Decode(dto); err != nil {
			return nil, 0, err
		}
		return dto, 0, nil
	case http.StatusNoContent:
		return nil, 0, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		d, err := strconv.Atoi(resp.Header.Get("Retry-After"))
		if err != nil {
			d = 30
		}
		return nil, time.Duration(d) * time.Second, ErrToManyRequests
	default:
		return nil, 0, fmt.Errorf("accrual status: %d", resp.StatusCode)
	}
}
