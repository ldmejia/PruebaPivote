package transfer

import "time"

const (
	StatusCompleted = "completed"
	StatusRejected  = "rejected"
)

const (
	MovementOpening = "APERTURA"
	MovementDebit   = "DEBITO"
	MovementCredit  = "CREDITO"
)

type Transfer struct {
	ID          string
	FromAccount string
	ToAccount   string
	Amount      int64
	Currency    string
	Status      string
	Reason      string
	CreatedAt   time.Time
}

type Movement struct {
	ID            int64
	AccountNumber string
	TransferID    string
	Type          string
	Amount        int64
	BalanceAfter  int64
	CreatedAt     time.Time
}

type Order struct {
	ID          string
	FromAccount string
	ToAccount   string
	Amount      int64
	CreatedAt   time.Time
}
