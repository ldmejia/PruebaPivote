package transfer

import "context"

type Repository interface {
	FindByIdempotencyKey(ctx context.Context, ownerID, key string) (*Transfer, string, error)
	Execute(ctx context.Context, o Order) (*Transfer, error)
}
