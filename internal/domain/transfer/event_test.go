package transfer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

func TestNewEvent(t *testing.T) {
	ok := transfer.NewEvent(&transfer.Transfer{ID: "t1", Amount: 2550, Status: transfer.StatusCompleted})
	rejected := transfer.NewEvent(&transfer.Transfer{ID: "t2", Status: transfer.StatusRejected, Reason: transfer.ReasonInsufficientFunds})

	assert.Equal(t, transfer.EventCompleted, ok.Type)
	assert.Equal(t, "25.50", ok.Data.Amount)
	assert.Equal(t, transfer.EventRejected, rejected.Type)
	assert.Equal(t, transfer.ReasonInsufficientFunds, rejected.Data.Reason)
	assert.NotEqual(t, ok.ID, rejected.ID)
}
