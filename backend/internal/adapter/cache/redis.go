package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		log.Printf("cache miss: %s", key)
		return nil, port.ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}
	log.Printf("cache hit: %s (%d bytes)", key, len(val))
	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	log.Printf("cache set: %s (%d bytes, ttl=%s)", key, len(value), ttl)
	return nil
}
