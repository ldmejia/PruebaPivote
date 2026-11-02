package transfer

import "time"

const (
	StatusCompleted = "completed"
	StatusRejected  = "rejected"
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

type Order struct {
	ID             string
	OwnerID        string
	IdempotencyKey string
	RequestHash    string
	FromAccount    string
	ToAccount      string
	Amount         int64
	CoreReason     string
	CreatedAt      time.Time
}
