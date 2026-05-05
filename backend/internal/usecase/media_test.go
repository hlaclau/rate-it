package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
)

type mockMediaCache struct {
	getFn func(ctx context.Context, key string) ([]byte, error)
	setFn func(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

func (m *mockMediaCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.getFn != nil {
		return m.getFn(ctx, key)
	}
	return nil, port.ErrCacheMiss
}
func (m *mockMediaCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, key, value, ttl)
	}
	return nil
}

func TestMediaUseCase_GetMovie_CacheHit(t *testing.T) {
	cache := &mockMediaCache{
		getFn: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"title":"Cached Movie"}`), nil
		},
	}
	uc := usecase.NewMediaUseCase(&mockMediaRepo{}, &mockMediaFetcher{}, cache)

	raw, err := uc.GetMovie(context.Background(), "123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(raw) != `{"title":"Cached Movie"}` {
		t.Errorf("unexpected cached value: %s", string(raw))
	}
}

func TestMediaUseCase_GetMovie_CacheMiss(t *testing.T) {
	fetcher := &mockMediaFetcher{
		fetchMovieFn: func(id string) ([]byte, *domain.Media, error) {
			return []byte(`{"title":"Fetched Movie"}`), &domain.Media{ID: "m1"}, nil
		},
	}
	repo := &mockMediaRepo{
		upsertFn: func(m *domain.Media) error { return nil },
	}
	cache := &mockMediaCache{} // Returns ErrCacheMiss by default

	uc := usecase.NewMediaUseCase(repo, fetcher, cache)
	raw, err := uc.GetMovie(context.Background(), "123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(raw) != `{"title":"Fetched Movie"}` {
		t.Errorf("unexpected fetched value: %s", string(raw))
	}
}

func TestMediaUseCase_SearchMedia_CacheMiss(t *testing.T) {
	fetcher := &mockMediaFetcher{
		fetchMediaFn: func(params domain.MediaParams) ([]byte, error) {
			return []byte(`{"results":[]}`), nil
		},
	}
	cache := &mockMediaCache{} // Returns ErrCacheMiss by default
	uc := usecase.NewMediaUseCase(&mockMediaRepo{}, fetcher, cache)

	params := domain.MediaParams{
		Type:  string(domain.TypeMovie),
		Query: "Inception",
	}
	raw, err := uc.SearchMedia(context.Background(), params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(raw) != `{"results":[]}` {
		t.Errorf("unexpected fetched search value: %s", string(raw))
	}
}
