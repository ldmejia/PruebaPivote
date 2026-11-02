package auditapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	auditapp "github.com/ldmejia/PruebaPivote/internal/application/audit"
	"github.com/ldmejia/PruebaPivote/internal/domain/audit"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

const validEvent = `{"event_id":"e1","type":"transfer.completed","data":{"transfer_id":"t1"}}`

func TestRecordEventUseCase_Execute_Saves(t *testing.T) {
	store := NewMockAuditStore(t)
	uc := auditapp.NewRecordEventUseCase(store, 3, 0)

	store.EXPECT().
		Save(mock.Anything, mock.MatchedBy(func(ev transfer.Event) bool { return ev.ID == "e1" }), "transfer.completed").
		Return(nil).
		Once()

	assert.NoError(t, uc.Execute(context.Background(), []byte(validEvent), "transfer.completed"))
}

func TestRecordEventUseCase_Execute_DuplicatedIsOk(t *testing.T) {
	store := NewMockAuditStore(t)
	uc := auditapp.NewRecordEventUseCase(store, 3, 0)

	store.EXPECT().Save(mock.Anything, mock.Anything, mock.Anything).Return(audit.ErrDuplicatedEvent).Once()

	assert.NoError(t, uc.Execute(context.Background(), []byte(validEvent), "transfer.completed"))
}

func TestRecordEventUseCase_Execute_InvalidEvent(t *testing.T) {
	store := NewMockAuditStore(t)
	uc := auditapp.NewRecordEventUseCase(store, 3, 0)

	for _, body := range []string{"no es json", `{"type":"transfer.completed"}`} {
		assert.ErrorIs(t, uc.Execute(context.Background(), []byte(body), "transfer.completed"), audit.ErrInvalidEvent)
	}
}

func TestRecordEventUseCase_Execute_StoreDownRetries(t *testing.T) {
	store := NewMockAuditStore(t)
	uc := auditapp.NewRecordEventUseCase(store, 3, 0)

	dbErr := errors.New("mongo caido")
	store.EXPECT().Save(mock.Anything, mock.Anything, mock.Anything).Return(dbErr).Times(3)

	assert.ErrorIs(t, uc.Execute(context.Background(), []byte(validEvent), "transfer.completed"), dbErr)
}
