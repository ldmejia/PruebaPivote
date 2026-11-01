package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var unlockScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0`)

type Locker struct {
	client *goredis.Client
}

func NewLocker(client *goredis.Client) *Locker {
	return &Locker{client: client}
}

func (l *Locker) Lock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	return token, ok, err
}

func (l *Locker) Unlock(ctx context.Context, key, token string) error {
	return unlockScript.Run(ctx, l.client, []string{key}, token).Err()
}
