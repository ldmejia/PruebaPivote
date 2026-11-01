package transferapp

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

const (
	lockTTL       = 30 * time.Second
	waitForResult = 3 * time.Second
)

type TransferUseCases struct {
	Create *CreateTransferUseCase
}

func NewTransferUseCases(accounts account.Repository, transfers transfer.Repository, locker Locker) *TransferUseCases {
	return &TransferUseCases{
		Create: NewCreateTransferUseCase(accounts, transfers, locker),
	}
}

type CreateTransferInput struct {
	UserID         string
	IdempotencyKey string
	FromAccount    string
	ToAccount      string
	Amount         int64
}

type CreateTransferOutput struct {
	Transfer *transfer.Transfer
	Replayed bool
}

type CreateTransferUseCase struct {
	accounts  account.Repository
	transfers transfer.Repository
	locker    Locker
}

func NewCreateTransferUseCase(accounts account.Repository, transfers transfer.Repository, locker Locker) *CreateTransferUseCase {
	return &CreateTransferUseCase{accounts: accounts, transfers: transfers, locker: locker}
}

func (u *CreateTransferUseCase) Execute(ctx context.Context, input CreateTransferInput) (*CreateTransferOutput, error) {
	if input.IdempotencyKey == "" || len(input.IdempotencyKey) > 100 {
		return nil, transfer.ErrIdempotencyKeyRequired
	}
	hash := transfer.RequestHash(input.FromAccount, input.ToAccount, input.Amount)
	if out, err := u.replay(ctx, input, hash); out != nil || err != nil {
		return out, err
	}

	lockKey := "idem:" + input.UserID + ":" + input.IdempotencyKey
	token, locked, err := u.locker.Lock(ctx, lockKey, lockTTL)
	if err == nil && !locked {
		return u.wait(ctx, input, hash)
	}
	if locked {
		defer u.locker.Unlock(context.WithoutCancel(ctx), lockKey, token)
	}

	order, err := u.newOrder(ctx, input, hash)
	if err != nil {
		return nil, err
	}
	t, err := u.transfers.Execute(ctx, order)
	if errors.Is(err, transfer.ErrDuplicatedKey) {
		return u.wait(ctx, input, hash)
	}
	if err != nil {
		return nil, err
	}
	return &CreateTransferOutput{Transfer: t}, nil
}

func (u *CreateTransferUseCase) newOrder(ctx context.Context, input CreateTransferInput, hash string) (transfer.Order, error) {
	if input.Amount <= 0 {
		return transfer.Order{}, transfer.ErrInvalidAmount
	}
	if input.FromAccount == input.ToAccount {
		return transfer.Order{}, transfer.ErrSameAccount
	}
	from, err := u.accounts.GetByNumber(ctx, input.FromAccount)
	if err != nil {
		return transfer.Order{}, err
	}
	if !from.BelongsTo(input.UserID) {
		return transfer.Order{}, account.ErrAccountNotFound
	}
	to, err := u.accounts.GetByNumber(ctx, input.ToAccount)
	if err != nil {
		return transfer.Order{}, err
	}
	if from.Currency != to.Currency {
		return transfer.Order{}, transfer.ErrCurrencyMismatch
	}
	return transfer.Order{
		ID:             uuid.NewString(),
		OwnerID:        input.UserID,
		IdempotencyKey: input.IdempotencyKey,
		RequestHash:    hash,
		FromAccount:    from.Number,
		ToAccount:      to.Number,
		Amount:         input.Amount,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

func (u *CreateTransferUseCase) replay(ctx context.Context, input CreateTransferInput, hash string) (*CreateTransferOutput, error) {
	t, savedHash, err := u.transfers.FindByIdempotencyKey(ctx, input.UserID, input.IdempotencyKey)
	if errors.Is(err, transfer.ErrTransferNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if savedHash != hash {
		return nil, transfer.ErrIdempotencyKeyReused
	}
	return &CreateTransferOutput{Transfer: t, Replayed: true}, nil
}

func (u *CreateTransferUseCase) wait(ctx context.Context, input CreateTransferInput, hash string) (*CreateTransferOutput, error) {
	deadline := time.Now().Add(waitForResult)
	for time.Now().Before(deadline) {
		if out, err := u.replay(ctx, input, hash); out != nil || err != nil {
			return out, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return nil, transfer.ErrRequestInProgress
}
