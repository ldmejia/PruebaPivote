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

func RequestHash(fromAccount, toAccount string, amount int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d", fromAccount, toAccount, amount)))
	return hex.EncodeToString(sum[:])
}
