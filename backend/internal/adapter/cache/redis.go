package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/redis/go-redis/v9"
)

// RedisCache implements port.MediaCache backed by Redis.
type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		slog.Debug("cache miss", "key", key)
		return nil, port.ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}
	slog.Debug("cache hit", "key", key, "bytes", len(val))
	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	slog.Debug("cache set", "key", key, "bytes", len(value), "ttl", ttl)
	return nil
}
