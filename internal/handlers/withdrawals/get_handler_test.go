package withdrawals

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mauth "github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	serviceauth "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/mock"
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
	list := []models.Withdrawal{{Order: "2377225624", Sum: 500, ProcessedAt: time.Now()}}

	cases := []struct {
		name       string
		authed     bool
		result     []models.Withdrawal
		getErr     error
		wantStatus int
	}{
		{"unauthorized", false, nil, nil, http.StatusUnauthorized},
		{"ok", true, list, nil, http.StatusOK},
		{"no content", true, []models.Withdrawal{}, nil, http.StatusNoContent},
		{"error", true, nil, context.DeadlineExceeded, http.StatusInternalServerError},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := &mock.Storage{
				GetWithdrawalsFn: func(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
					return c.result, c.getErr
				},
			}
			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			rec := serve(GetHandler(store), req, c.authed)
			if rec.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, c.wantStatus)
			}
		})
	}
}
