package transferapp

import (
	"context"
	"time"
)

type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (string, bool, error)
	Unlock(ctx context.Context, key, token string) error
}
