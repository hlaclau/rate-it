package port

import (
	"context"
	"errors"
	"time"

	"github.com/hlaclau/rate-it-api/internal/domain"
)

// ErrNotFound is returned when the requested media does not exist in the external source.
var ErrNotFound = errors.New("not found")

// ErrCacheMiss is returned by MediaCache.Get when the key is absent.
var ErrCacheMiss = errors.New("cache miss")

// MediaRepository persists base media records used for FK integrity and card rendering.
type MediaRepository interface {
	Upsert(m *domain.Media) error
}

// MediaFetcher retrieves full media details from an external source (e.g. TMDB).
// It returns the raw JSON payload and a domain.Media populated with base fields.
type MediaFetcher interface {
	FetchMovie(id string) (raw []byte, m *domain.Media, err error)
	FetchSeries(id string) (raw []byte, m *domain.Media, err error)
	SearchMedia(query string) ([]byte, error)
}

// MediaCache is a key/value store for raw media payloads.
type MediaCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}
