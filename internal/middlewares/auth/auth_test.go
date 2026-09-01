package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	serviceauth "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
)

func TestMiddleware(t *testing.T) {
	tm := serviceauth.NewTokenManager("test-secret")
	validToken, _ := tm.BuildJWTString(99)

	cases := []struct {
		name       string
		header     string
		wantStatus int
		wantNext   bool
	}{
		{"no header", "", http.StatusUnauthorized, false},
		{"wrong scheme", "Token " + validToken, http.StatusUnauthorized, false},
		{"invalid token", "Bearer garbage", http.StatusUnauthorized, false},
		{"valid token", "Bearer " + validToken, http.StatusOK, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			nextCalled := false
			var gotID int64
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotID, _ = UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if c.header != "" {
				req.Header.Set("Authorization", c.header)
			}
			rec := httptest.NewRecorder()

			Middleware(tm)(next).ServeHTTP(rec, req)

			if rec.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, c.wantStatus)
			}
			if nextCalled != c.wantNext {
				t.Errorf("nextCalled = %v, want %v", nextCalled, c.wantNext)
			}
			if c.wantNext && gotID != 99 {
				t.Errorf("userID in context = %d, want 99", gotID)
			}
		})
	}
}
