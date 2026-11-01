package transfer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name   string
		from   account.Account
		to     account.Account
		amount int64
		want   string
	}{
		{"pasa", account.Account{Balance: 1000}, account.Account{}, 500, ""},
		{"saldo exacto", account.Account{Balance: 1000}, account.Account{}, 1000, ""},
		{"sin saldo", account.Account{Balance: 1000}, account.Account{}, 1001, transfer.ReasonInsufficientFunds},
		{"origen bloqueada", account.Account{Balance: 1000, Blocked: true}, account.Account{}, 10, transfer.ReasonAccountBlocked},
		{"destino bloqueada", account.Account{Balance: 1000}, account.Account{Blocked: true}, 10, transfer.ReasonAccountBlocked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, transfer.Check(&tt.from, &tt.to, tt.amount))
		})
	}
}
