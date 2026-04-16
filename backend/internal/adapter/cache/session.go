package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionCache struct {
	client *redis.Client
}

func NewSessionCache(client *redis.Client) *SessionCache {
	return &SessionCache{client: client}
}

func (c *SessionCache) Set(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)
	return c.client.Set(ctx, key, token, ttl).Err()
}

func (c *SessionCache) Get(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("session:%s", userID)
	return c.client.Get(ctx, key).Result()
}

func (c *SessionCache) Delete(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return c.client.Del(ctx, key).Err()
}
