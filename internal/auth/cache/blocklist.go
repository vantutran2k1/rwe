package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Blocklist interface {
	AddToBlocklist(ctx context.Context, tokenID string, duration time.Duration) error
	IsBlocklisted(ctx context.Context, tokenID string) (bool, error)
}

type RedisBlocklist struct {
	client *redis.Client
}

func NewRedisBlocklist(client *redis.Client) *RedisBlocklist {
	return &RedisBlocklist{client: client}
}

func (r *RedisBlocklist) AddToBlocklist(ctx context.Context, tokenID string, duration time.Duration) error {
	key := fmt.Sprintf("blocklist:%s", tokenID)
	return r.client.Set(ctx, key, "1", duration).Err()
}

func (r *RedisBlocklist) IsBlocklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blocklist:%s", tokenID)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
