package account

import "context"

type Repository interface {
	Create(ctx context.Context, a *Account) error
	GetByNumber(ctx context.Context, number string) (*Account, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*Account, error)
	UpdateBlocked(ctx context.Context, a *Account) error
}
