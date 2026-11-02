package redis

import (
	"context"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const CoreValidationFlag = "feature:core_validation"

type Flags struct {
	client *goredis.Client
}

func NewFlags(client *goredis.Client) *Flags {
	return &Flags{client: client}
}

func (f *Flags) CoreValidationEnabled(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	v, err := f.client.Get(ctx, CoreValidationFlag).Result()
	if err != nil {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "off", "false", "0":
		return false
	}
	return true
}
