package port

import (
	"context"
	"time"
)

type SessionRepository interface {
	Set(ctx context.Context, userID, refreshToken string, ttl time.Duration) error
	Get(ctx context.Context, userID string) (string, error)
	Delete(ctx context.Context, userID string) error
}
