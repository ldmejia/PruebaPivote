package transfer

import (
	"time"

	"github.com/google/uuid"

	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
)

const (
	EventCompleted = "transfer.completed"
	EventRejected  = "transfer.rejected"
)

type Event struct {
	ID         string    `json:"event_id" bson:"_id"`
	Type       string    `json:"type" bson:"type"`
	OccurredAt time.Time `json:"occurred_at" bson:"occurred_at"`
	Data       EventData `json:"data" bson:"data"`
}

type EventData struct {
	TransferID  string `json:"transfer_id" bson:"transfer_id"`
	FromAccount string `json:"from_account" bson:"from_account"`
	ToAccount   string `json:"to_account" bson:"to_account"`
	Amount      string `json:"amount" bson:"amount"`
	Currency    string `json:"currency" bson:"currency"`
	Status      string `json:"status" bson:"status"`
	Reason      string `json:"reason,omitempty" bson:"reason,omitempty"`
}

func NewEvent(t *Transfer) Event {
	eventType := EventCompleted
	if t.Status == StatusRejected {
		eventType = EventRejected
	}
	return Event{
		ID:         uuid.NewString(),
		Type:       eventType,
		OccurredAt: t.CreatedAt,
		Data: EventData{
			TransferID:  t.ID,
			FromAccount: t.FromAccount,
			ToAccount:   t.ToAccount,
			Amount:      shared.FormatAmount(t.Amount),
			Currency:    t.Currency,
			Status:      t.Status,
			Reason:      t.Reason,
		},
	}
}
