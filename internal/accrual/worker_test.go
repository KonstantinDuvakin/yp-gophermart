package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/models"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage/mock"
	"go.uber.org/mock/gomock"
)

func TestMapAccrualStatus(t *testing.T) {
	accrual := 100.0
	cases := []struct {
		name       string
		dto        *Dto
		wantStatus string
		wantNil    bool
	}{
		{"processed", &Dto{Status: StatusProcessed, Accrual: &accrual}, models.OrderStatusProcessed, false},
		{"invalid", &Dto{Status: StatusInvalid}, models.OrderStatusInvalid, true},
		{"processing", &Dto{Status: "PROCESSING"}, models.OrderStatusProcessing, true},
		{"registered", &Dto{Status: "REGISTERED"}, models.OrderStatusProcessing, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, acc := mapAccrualStatus(c.dto)
			if status != c.wantStatus {
				t.Errorf("status = %q, want %q", status, c.wantStatus)
			}
			if c.wantNil && acc != nil {
				t.Errorf("accrual = %v, want nil", acc)
			}
			if !c.wantNil && acc == nil {
				t.Error("accrual = nil, want value")
			}
		})
	}
}

func TestProcessBatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"order":"100","status":"PROCESSED","accrual":50}`))
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().
		GetUnprocessedOrders(gomock.Any()).
		Return([]string{"100"}, nil)

	var gotStatus string
	var gotAcc *float64
	store.EXPECT().
		UpdateOrderAccrual(gomock.Any(), "100", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, status string, accrual *float64) error {
			gotStatus, gotAcc = status, accrual
			return nil
		})

	w := &Worker{store: store, client: NewClient(srv.URL)}
	w.processBatch(context.Background())

	if gotStatus != models.OrderStatusProcessed {
		t.Errorf("status = %q, want %q", gotStatus, models.OrderStatusProcessed)
	}
	if gotAcc == nil || *gotAcc != 50 {
		t.Errorf("accrual = %v, want 50", gotAcc)
	}
}
