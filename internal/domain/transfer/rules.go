package transfer

import "github.com/ldmejia/PruebaPivote/internal/domain/account"

const (
	ReasonAccountBlocked    = "cuenta_bloqueada"
	ReasonInsufficientFunds = "saldo_insuficiente"
)

func Check(from, to *account.Account, amount int64) string {
	if from.Blocked || to.Blocked {
		return ReasonAccountBlocked
	}
	if from.Balance < amount {
		return ReasonInsufficientFunds
	}
	return ""
}
