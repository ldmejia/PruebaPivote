package account

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, a *Account) error
	GetByNumber(ctx context.Context, number string) (*Account, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*Account, error)
	UpdateBlocked(ctx context.Context, a *Account) error
	ListMovements(ctx context.Context, number string, from, to time.Time, limit, offset int) ([]*Movement, int, error)
}
