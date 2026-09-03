package accrual

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/ok"):
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"order":"ok","status":"PROCESSED","accrual":50}`))
		case strings.HasSuffix(r.URL.Path, "/none"):
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/limit"):
			w.Header().Set("Retry-After", "3")
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
}

func TestGetOrderAccrual_OK(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	dto, _, err := NewClient(srv.URL).GetOrderAccrual(context.Background(), "ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.Status != StatusProcessed {
		t.Errorf("status = %q, want %q", dto.Status, StatusProcessed)
	}
	if dto.Accrual == nil || *dto.Accrual != 50 {
		t.Errorf("accrual = %v, want 50", dto.Accrual)
	}
}

func TestGetOrderAccrual_NoContent(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	_, _, err := NewClient(srv.URL).GetOrderAccrual(context.Background(), "none")
	if !errors.Is(err, ErrOrderNotRegistered) {
		t.Errorf("err = %v, want ErrOrderNotRegistered", err)
	}
}

func TestGetOrderAccrual_TooManyRequests(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	_, retryAfter, err := NewClient(srv.URL).GetOrderAccrual(context.Background(), "limit")
	if !errors.Is(err, ErrTooManyRequests) {
		t.Errorf("err = %v, want ErrTooManyRequests", err)
	}
	if retryAfter != 3*time.Second {
		t.Errorf("retryAfter = %v, want 3s", retryAfter)
	}
}

func TestGetOrderAccrual_ServerError(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	_, _, err := NewClient(srv.URL).GetOrderAccrual(context.Background(), "boom")
	if err == nil || errors.Is(err, ErrOrderNotRegistered) || errors.Is(err, ErrTooManyRequests) {
		t.Errorf("err = %v, want generic error", err)
	}
}
