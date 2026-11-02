package account

import "time"

const (
	MovementOpening = "APERTURA"
	MovementDebit   = "DEBITO"
	MovementCredit  = "CREDITO"
)

type Movement struct {
	ID            int64
	AccountNumber string
	TransferID    string
	Type          string
	Amount        int64
	BalanceAfter  int64
	CreatedAt     time.Time
}
