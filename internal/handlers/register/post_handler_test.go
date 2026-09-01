package register

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	serviceauth "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/mock"
	"go.uber.org/mock/gomock"
)

func TestPostHandler(t *testing.T) {
	tm := serviceauth.NewTokenManager("test")

	cases := []struct {
		name        string
		body        string
		createErr   error
		wantStatus  int
		wantAuthHdr bool
	}{
		{"success", `{"login":"user","password":"pass"}`, nil, http.StatusOK, true},
		{"login taken", `{"login":"user","password":"pass"}`, storage.ErrLoginTaken, http.StatusConflict, false},
		{"bad json", `{oops`, nil, http.StatusBadRequest, false},
		{"empty fields", `{"login":"","password":""}`, nil, http.StatusBadRequest, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(int64(1), c.createErr).
				AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(c.body))
			rec := httptest.NewRecorder()

			PostHandler(store, tm).ServeHTTP(rec, req)

			if rec.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, c.wantStatus)
			}
			hasAuth := strings.HasPrefix(rec.Header().Get("Authorization"), "Bearer ")
			if hasAuth != c.wantAuthHdr {
				t.Errorf("Authorization present = %v, want %v", hasAuth, c.wantAuthHdr)
			}
		})
	}
}
