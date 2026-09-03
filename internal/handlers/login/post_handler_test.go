package login

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	serviceauth "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/mock"
	"go.uber.org/mock/gomock"
)

func TestPostHandler(t *testing.T) {
	tm := serviceauth.NewTokenManager("test")
	hash, _ := serviceauth.HashPassword("pass")

	cases := []struct {
		name        string
		body        string
		user        models.User
		getErr      error
		wantStatus  int
		wantAuthHdr bool
	}{
		{"success", `{"login":"user","password":"pass"}`, models.User{ID: 1, Login: "user", PasswordHash: hash}, nil, http.StatusOK, true},
		{"user not found", `{"login":"user","password":"pass"}`, models.User{}, storage.ErrUserNotFound, http.StatusUnauthorized, false},
		{"wrong password", `{"login":"user","password":"wrong"}`, models.User{ID: 1, Login: "user", PasswordHash: hash}, nil, http.StatusUnauthorized, false},
		{"bad json", `{oops`, models.User{}, nil, http.StatusBadRequest, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			store.EXPECT().
				GetUserByLogin(gomock.Any(), gomock.Any()).
				Return(c.user, c.getErr).
				AnyTimes()

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(c.body))
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
