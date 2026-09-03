package orders

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mauth "github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	serviceauth "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/mock"
	"go.uber.org/mock/gomock"
)

const validOrder = "12345678903"

// serve прогоняет запрос через хендлер, при authed=true — сквозь auth-middleware
// с валидным токеном пользователя 7.
func serve(handler http.Handler, req *http.Request, authed bool) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	if authed {
		tm := serviceauth.NewTokenManager("test")
		token, _ := tm.BuildJWTString(7)
		req.Header.Set("Authorization", "Bearer "+token)
		handler = mauth.Middleware(tm)(handler)
	}
	handler.ServeHTTP(rec, req)
	return rec
}

func TestPostHandler(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		authed     bool
		createErr  error
		wantStatus int
	}{
		{"unauthorized", validOrder, false, nil, http.StatusUnauthorized},
		{"accepted", validOrder, true, nil, http.StatusAccepted},
		{"invalid luhn", "123", true, nil, http.StatusUnprocessableEntity},
		{"empty body", "", true, nil, http.StatusBadRequest},
		{"owned by user", validOrder, true, storage.ErrOrderOwnedByUser, http.StatusOK},
		{"owned by other", validOrder, true, storage.ErrOrderOwnedByOther, http.StatusConflict},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			store.EXPECT().
				CreateOrder(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(c.createErr).
				AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(c.body))
			rec := serve(PostHandler(store), req, c.authed)
			if rec.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, c.wantStatus)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	orders := []models.Order{{Number: validOrder, Status: models.OrderStatusNew, UploadedAt: time.Now()}}

	cases := []struct {
		name       string
		authed     bool
		result     []models.Order
		getErr     error
		wantStatus int
	}{
		{"unauthorized", false, nil, nil, http.StatusUnauthorized},
		{"ok", true, orders, nil, http.StatusOK},
		{"no content", true, []models.Order{}, nil, http.StatusNoContent},
		{"error", true, nil, context.DeadlineExceeded, http.StatusInternalServerError},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			store.EXPECT().
				GetOrders(gomock.Any(), gomock.Any()).
				Return(c.result, c.getErr).
				AnyTimes()

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			rec := serve(GetHandler(store), req, c.authed)
			if rec.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, c.wantStatus)
			}
		})
	}
}
