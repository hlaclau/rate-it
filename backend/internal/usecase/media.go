package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
)

const cacheTTL = 7 * 24 * time.Hour
const searchCacheTTL = 24 * time.Hour

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

func (uc *MediaUseCase) SearchMedia(ctx context.Context, params domain.MediaParams) ([]byte, error) {
	params.Query = strings.ToLower(strings.TrimSpace(params.Query))

	page := params.Page
	if page < 1 {
		page = 1
	}

	var key string
	if params.Query != "" {
		adult := 0
		if params.IncludeAdult {
			adult = 1
		}
		key = fmt.Sprintf("search:tmdb:%s:%s:p%d:lang:%s:adult:%d",
			params.Type, params.Query, page, strings.ToLower(params.Language), adult)
	} else {
		key = fmt.Sprintf("discover:tmdb:%s:%s:yf%d:yt%d:vmin%.1f:vmax%.1f:vc%d:lang:%s:genres:%s:wp:%s:%s:p%d",
			params.Type, params.SortBy,
			params.YearFrom, params.YearTo,
			params.VoteAverageMin, params.VoteAverageMax,
			params.VoteCountMin,
			strings.ToLower(params.Language),
			params.WithGenres,
			params.WatchProviders, strings.ToUpper(params.WatchRegion),
			page)
	}

	raw, err := uc.cache.Get(ctx, key)
	if err == nil {
		return raw, nil
	}
	if !errors.Is(err, port.ErrCacheMiss) {
		return nil, err
	}

	raw, err = uc.fetcher.FetchMedia(params)
	if err != nil {
		return nil, err
	}

	return raw, uc.cache.Set(ctx, key, raw, searchCacheTTL)
}

func (uc *MediaUseCase) get(
	ctx context.Context,
	id string,
	mediaType domain.MediaType,
	fetch func(string) ([]byte, *domain.Media, error),
) ([]byte, error) {
	key := fmt.Sprintf("media:tmdb:%s:%s:v2", mediaType, id)

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
