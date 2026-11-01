package transfer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

func TestCheck(t *testing.T) {
	gtq := func(balance int64) *account.Account {
		return &account.Account{Currency: account.CurrencyGTQ, Balance: balance}
	}
	usd := &account.Account{Currency: account.CurrencyUSD, Balance: 500000}

	tests := []struct {
		name   string
		from   *account.Account
		to     *account.Account
		amount int64
		used   int64
		want   string
	}{
		{"pasa", gtq(1000), gtq(0), 500, 0, ""},
		{"saldo exacto", gtq(1000), gtq(0), 1000, 0, ""},
		{"sin saldo", gtq(1000), gtq(0), 1001, 0, transfer.ReasonInsufficientFunds},
		{"origen bloqueada", &account.Account{Balance: 1000, Blocked: true}, gtq(0), 10, 0, transfer.ReasonAccountBlocked},
		{"destino bloqueada", gtq(1000), &account.Account{Blocked: true}, 10, 0, transfer.ReasonAccountBlocked},
		{"gtq justo en el limite", gtq(5000000), gtq(0), 400000, 600000, ""},
		{"gtq un centavo arriba del limite", gtq(5000000), gtq(0), 400001, 600000, transfer.ReasonDailyLimit},
		{"usd limite 1300", usd, usd, 1, 130000, transfer.ReasonDailyLimit},
		{"usd justo en el limite", usd, usd, 130000, 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, transfer.Check(tt.from, tt.to, tt.amount, tt.used))
		})
	}
}

func TestRequestHash(t *testing.T) {
	assert.Equal(t, transfer.RequestHash("01", "02", 100), transfer.RequestHash("01", "02", 100))
	assert.NotEqual(t, transfer.RequestHash("01", "02", 100), transfer.RequestHash("01", "02", 101))
}
