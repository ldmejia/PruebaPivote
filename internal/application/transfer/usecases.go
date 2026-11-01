package transferapp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type TransferUseCases struct {
	Create *CreateTransferUseCase
}

func NewTransferUseCases(accounts account.Repository, transfers transfer.Repository) *TransferUseCases {
	return &TransferUseCases{
		Create: NewCreateTransferUseCase(accounts, transfers),
	}
}

type CreateTransferInput struct {
	UserID      string
	FromAccount string
	ToAccount   string
	Amount      int64
}

type CreateTransferUseCase struct {
	accounts  account.Repository
	transfers transfer.Repository
}

func NewCreateTransferUseCase(accounts account.Repository, transfers transfer.Repository) *CreateTransferUseCase {
	return &CreateTransferUseCase{accounts: accounts, transfers: transfers}
}

func (u *CreateTransferUseCase) Execute(ctx context.Context, input CreateTransferInput) (*transfer.Transfer, error) {
	if input.Amount <= 0 {
		return nil, transfer.ErrInvalidAmount
	}
	if input.FromAccount == input.ToAccount {
		return nil, transfer.ErrSameAccount
	}
	from, err := u.accounts.GetByNumber(ctx, input.FromAccount)
	if err != nil {
		return nil, err
	}
	if !from.BelongsTo(input.UserID) {
		return nil, account.ErrAccountNotFound
	}
	to, err := u.accounts.GetByNumber(ctx, input.ToAccount)
	if err != nil {
		return nil, err
	}
	if from.Currency != to.Currency {
		return nil, transfer.ErrCurrencyMismatch
	}
	return u.transfers.Execute(ctx, transfer.Order{
		ID:          uuid.NewString(),
		FromAccount: from.Number,
		ToAccount:   to.Number,
		Amount:      input.Amount,
		CreatedAt:   time.Now().UTC(),
	})
}
