package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
)

const (
	ReasonAccountBlocked    = "cuenta_bloqueada"
	ReasonInsufficientFunds = "saldo_insuficiente"
	ReasonDailyLimit        = "limite_diario_excedido"
)

func DailyLimit(currency string) int64 {
	if currency == account.CurrencyUSD {
		return 130000
	}
	return 1000000
}

func Check(from, to *account.Account, amount, usedToday int64) string {
	if from.Blocked || to.Blocked {
		return ReasonAccountBlocked
	}
	if from.Balance < amount {
		return ReasonInsufficientFunds
	}
	if usedToday+amount > DailyLimit(from.Currency) {
		return ReasonDailyLimit
	}
	return ""
}

func RequestHash(fromAccount, toAccount string, amount int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d", fromAccount, toAccount, amount)))
	return hex.EncodeToString(sum[:])
}
