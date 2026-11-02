package auditapp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ldmejia/PruebaPivote/internal/domain/audit"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type RecordEventUseCase struct {
	store     audit.Store
	retries   int
	retryWait time.Duration
}

func NewRecordEventUseCase(store audit.Store, retries int, retryWait time.Duration) *RecordEventUseCase {
	return &RecordEventUseCase{store: store, retries: retries, retryWait: retryWait}
}

func (u *RecordEventUseCase) Execute(ctx context.Context, body []byte, routingKey string) error {
	var ev transfer.Event
	if err := json.Unmarshal(body, &ev); err != nil || ev.ID == "" {
		return audit.ErrInvalidEvent
	}
	var err error
	for i := 1; i <= u.retries; i++ {
		err = u.store.Save(ctx, ev, routingKey)
		if err == nil || errors.Is(err, audit.ErrDuplicatedEvent) {
			return nil
		}
		time.Sleep(time.Duration(i) * u.retryWait)
	}
	return err
}
