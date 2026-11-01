package transfer

import "context"

type Repository interface {
	Execute(ctx context.Context, o Order) (*Transfer, error)
}
