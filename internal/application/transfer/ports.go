package transferapp

import (
	"context"
	"time"

	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (string, bool, error)
	Unlock(ctx context.Context, key, token string) error
}

type CoreValidator interface {
	ValidateBeneficiary(ctx context.Context, t *transfer.Transfer) error
}

type Flags interface {
	CoreValidationEnabled(ctx context.Context) bool
}
