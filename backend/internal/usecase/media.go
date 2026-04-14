package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
)

const cacheTTL = 7 * 24 * time.Hour

type MediaUseCase struct {
	repo    port.MediaRepository
	fetcher port.MediaFetcher
	cache   port.MediaCache
}

func NewMediaUseCase(repo port.MediaRepository, fetcher port.MediaFetcher, cache port.MediaCache) *MediaUseCase {
	return &MediaUseCase{repo: repo, fetcher: fetcher, cache: cache}
}

func (uc *MediaUseCase) GetMovie(ctx context.Context, id string) ([]byte, error) {
	return uc.get(ctx, id, domain.TypeMovie, uc.fetcher.FetchMovie)
}

func (uc *MediaUseCase) GetSeries(ctx context.Context, id string) ([]byte, error) {
	return uc.get(ctx, id, domain.TypeSeries, uc.fetcher.FetchSeries)
}

func (uc *MediaUseCase) get(
	ctx context.Context,
	id string,
	mediaType domain.MediaType,
	fetch func(string) ([]byte, *domain.Media, error),
) ([]byte, error) {
	key := fmt.Sprintf("media:tmdb:%s:%s", mediaType, id)

	raw, err := uc.cache.Get(ctx, key)
	if err == nil {
		return raw, nil
	}
	if !errors.Is(err, port.ErrCacheMiss) {
		return nil, err
	}

	raw, media, err := fetch(id)
	if err != nil {
		return nil, err
	}

	if err = uc.cache.Set(ctx, key, raw, cacheTTL); err != nil {
		return nil, err
	}

	if err = uc.repo.Upsert(media); err != nil {
		return nil, err
	}

	return raw, nil
}
