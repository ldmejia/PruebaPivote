package transfer

import (
	"time"

	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type CreateTransferRequest struct {
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      string `json:"amount"`
}

type TransferResponse struct {
	ID          string    `json:"id"`
	FromAccount string    `json:"from_account"`
	ToAccount   string    `json:"to_account"`
	Amount      string    `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func toTransferResponse(t *transfer.Transfer) TransferResponse {
	return TransferResponse{
		ID:          t.ID,
		FromAccount: t.FromAccount,
		ToAccount:   t.ToAccount,
		Amount:      shared.FormatAmount(t.Amount),
		Currency:    t.Currency,
		Status:      t.Status,
		Reason:      t.Reason,
		CreatedAt:   t.CreatedAt,
	}
}
