package withdraw

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mauth "github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	serviceauth "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/mock"
	"go.uber.org/mock/gomock"
)

const validOrder = "12345678903"

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
		name        string
		body        string
		authed      bool
		withdrawErr error
		wantStatus  int
	}{
		{"unauthorized", `{"order":"` + validOrder + `","sum":100}`, false, nil, http.StatusUnauthorized},
		{"success", `{"order":"` + validOrder + `","sum":100}`, true, nil, http.StatusOK},
		{"insufficient funds", `{"order":"` + validOrder + `","sum":100}`, true, storage.ErrInsufficientFunds, http.StatusPaymentRequired},
		{"invalid luhn", `{"order":"123","sum":100}`, true, nil, http.StatusUnprocessableEntity},
		{"non-positive sum", `{"order":"` + validOrder + `","sum":0}`, true, nil, http.StatusBadRequest},
		{"bad json", `{oops`, true, nil, http.StatusBadRequest},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			store.EXPECT().
				Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(c.withdrawErr).
				AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(c.body))
			rec := serve(PostHandler(store), req, c.authed)
			if rec.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, c.wantStatus)
			}
		})
	}
}
