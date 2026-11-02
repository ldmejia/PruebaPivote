package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

func coreServer(delay time.Duration, status int, approved bool, calls *int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			return
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"approved": approved})
	}))
}

func sample() *transfer.Transfer {
	return &transfer.Transfer{ID: "t1", FromAccount: "01", ToAccount: "02", Amount: 600000, Currency: "GTQ"}
}

func TestValidateBeneficiary(t *testing.T) {
	tests := []struct {
		name     string
		delay    time.Duration
		status   int
		approved bool
		want     error
	}{
		{"aprobado", 0, http.StatusOK, true, nil},
		{"rechazado", 0, http.StatusOK, false, transfer.ErrBeneficiaryRejected},
		{"error del core", 0, http.StatusServiceUnavailable, false, transfer.ErrCoreUnavailable},
		{"timeout", 500 * time.Millisecond, http.StatusOK, true, transfer.ErrCoreUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int32
			srv := coreServer(tt.delay, tt.status, tt.approved, &calls)
			defer srv.Close()
			c := NewClient(srv.URL, 100*time.Millisecond, NewBreaker(5, time.Minute))

			start := time.Now()
			err := c.ValidateBeneficiary(context.Background(), sample())

			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.want)
			}
			assert.Less(t, time.Since(start), 300*time.Millisecond, "nunca espera mas que el timeout")
		})
	}
}

func TestBreakerAvoidsCallingCore(t *testing.T) {
	var calls int32
	srv := coreServer(0, http.StatusServiceUnavailable, false, &calls)
	defer srv.Close()
	c := NewClient(srv.URL, time.Second, NewBreaker(3, time.Minute))

	for i := 0; i < 10; i++ {
		_ = c.ValidateBeneficiary(context.Background(), sample())
	}

	assert.Equal(t, int32(3), calls, "con el breaker abierto ya no se llama al core")
}

func TestBreakerOpensAndRecovers(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	b := NewBreaker(3, 10*time.Second)
	b.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		assert.True(t, b.Allow())
		b.Failure()
	}
	assert.False(t, b.Allow())
	assert.Equal(t, "open", b.State())

	now = now.Add(11 * time.Second)
	assert.True(t, b.Allow(), "despues del cooldown deja pasar una prueba")
	assert.False(t, b.Allow(), "solo una prueba a la vez")
	b.Success()
	assert.Equal(t, "closed", b.State())
}
