package accountapp

import (
	"context"
	"errors"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
)

type AccountUseCases struct {
	Open    *OpenAccountUseCase
	Get     *GetAccountUseCase
	List    *ListAccountsUseCase
	Block   *BlockAccountUseCase
	Unblock *UnblockAccountUseCase
}

func NewAccountUseCases(repo account.Repository) *AccountUseCases {
	return &AccountUseCases{
		Open:    NewOpenAccountUseCase(repo),
		Get:     NewGetAccountUseCase(repo),
		List:    NewListAccountsUseCase(repo),
		Block:   NewBlockAccountUseCase(repo),
		Unblock: NewUnblockAccountUseCase(repo),
	}
}

type OpenAccountInput struct {
	OwnerID        string
	Currency       string
	InitialBalance int64
}

type OpenAccountUseCase struct{ repo account.Repository }

func NewOpenAccountUseCase(repo account.Repository) *OpenAccountUseCase {
	return &OpenAccountUseCase{repo: repo}
}

func (u *OpenAccountUseCase) Execute(ctx context.Context, input OpenAccountInput) (*account.Account, error) {
	for i := 0; i < 5; i++ {
		a, err := account.New(input.OwnerID, input.Currency, input.InitialBalance)
		if err != nil {
			return nil, err
		}
		err = u.repo.Create(ctx, a)
		if errors.Is(err, account.ErrDuplicatedNumber) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return a, nil
	}
	return nil, account.ErrDuplicatedNumber
}

type GetAccountInput struct {
	UserID string
	Number string
}

type GetAccountUseCase struct{ repo account.Repository }

func NewGetAccountUseCase(repo account.Repository) *GetAccountUseCase {
	return &GetAccountUseCase{repo: repo}
}

func (u *GetAccountUseCase) Execute(ctx context.Context, input GetAccountInput) (*account.Account, error) {
	a, err := u.repo.GetByNumber(ctx, input.Number)
	if err != nil {
		return nil, err
	}
	if !a.BelongsTo(input.UserID) {
		return nil, account.ErrAccountNotFound
	}
	return a, nil
}

type ListAccountsUseCase struct{ repo account.Repository }

func NewListAccountsUseCase(repo account.Repository) *ListAccountsUseCase {
	return &ListAccountsUseCase{repo: repo}
}

func (u *ListAccountsUseCase) Execute(ctx context.Context, ownerID string) ([]*account.Account, error) {
	return u.repo.ListByOwner(ctx, ownerID)
}

type BlockAccountUseCase struct{ get *GetAccountUseCase }

func NewBlockAccountUseCase(repo account.Repository) *BlockAccountUseCase {
	return &BlockAccountUseCase{get: NewGetAccountUseCase(repo)}
}

func (u *BlockAccountUseCase) Execute(ctx context.Context, input GetAccountInput) (*account.Account, error) {
	a, err := u.get.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	a.Block()
	if err := u.get.repo.UpdateBlocked(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

type UnblockAccountUseCase struct{ get *GetAccountUseCase }

func NewUnblockAccountUseCase(repo account.Repository) *UnblockAccountUseCase {
	return &UnblockAccountUseCase{get: NewGetAccountUseCase(repo)}
}

func (u *UnblockAccountUseCase) Execute(ctx context.Context, input GetAccountInput) (*account.Account, error) {
	a, err := u.get.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	a.Unblock()
	if err := u.get.repo.UpdateBlocked(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
