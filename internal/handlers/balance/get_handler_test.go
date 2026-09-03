package balance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	mauth "github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	serviceauth "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/mock"
	"go.uber.org/mock/gomock"
)

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

func TestGetHandler(t *testing.T) {
	cases := []struct {
		name       string
		authed     bool
		balance    models.Balance
		getErr     error
		wantStatus int
	}{
		{"unauthorized", false, models.Balance{}, nil, http.StatusUnauthorized},
		{"ok", true, models.Balance{Current: 500.5, Withdrawn: 42}, nil, http.StatusOK},
		{"error", true, models.Balance{}, context.DeadlineExceeded, http.StatusInternalServerError},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			store.EXPECT().
				GetBalance(gomock.Any(), gomock.Any()).
				Return(c.balance, c.getErr).
				AnyTimes()

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			rec := serve(GetHandler(store), req, c.authed)
			if rec.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, c.wantStatus)
			}
		})
	}
}
