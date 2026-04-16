package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
)

var ErrInvalidRating = errors.New("rating only allowed for watched status")

type ListUseCase struct {
	listRepo  port.UserMediaRepository
	mediaRepo port.MediaRepository
	fetcher   port.MediaFetcher
}

func NewListUseCase(
	listRepo port.UserMediaRepository,
	mediaRepo port.MediaRepository,
	fetcher port.MediaFetcher,
) *ListUseCase {
	return &ListUseCase{
		listRepo:  listRepo,
		mediaRepo: mediaRepo,
		fetcher:   fetcher,
	}
}

func (uc *ListUseCase) AddOrUpdate(ctx context.Context, userID string, params port.AddOrUpdateParams) error {
	if params.Rating != nil && params.Status != domain.StatusWatched {
		return ErrInvalidRating
	}

	var (
		raw   []byte
		media *domain.Media
		err   error
	)

	switch params.Type {
	case domain.TypeMovie:
		raw, media, err = uc.fetcher.FetchMovie(params.ExternalID)
	case domain.TypeSeries:
		raw, media, err = uc.fetcher.FetchSeries(params.ExternalID)
	default:
		return fmt.Errorf("unsupported media type: %s", params.Type)
	}
	_ = raw

	if err != nil {
		return fmt.Errorf("fetch media: %w", err)
	}

	if err = uc.mediaRepo.Upsert(media); err != nil {
		return fmt.Errorf("upsert media: %w", err)
	}

	um := domain.UserMedia{
		UserID:  userID,
		MediaID: media.ID,
		Status:  params.Status,
		Rating:  params.Rating,
		Review:  params.Review,
		AddedAt: time.Now(),
	}
	return uc.listRepo.Upsert(ctx, &um)
}

func (uc *ListUseCase) Update(ctx context.Context, userID, mediaID string, status domain.UserMediaStatus, rating *int16, review *string) error {
	if rating != nil && status != domain.StatusWatched {
		return ErrInvalidRating
	}

	um := domain.UserMedia{
		UserID:  userID,
		MediaID: mediaID,
		Status:  status,
		Rating:  rating,
		Review:  review,
		AddedAt: time.Now(),
	}
	return uc.listRepo.Upsert(ctx, &um)
}

func (uc *ListUseCase) GetList(ctx context.Context, userID string) ([]port.ListEntry, error) {
	return uc.listRepo.ListByUser(ctx, userID)
}

func (uc *ListUseCase) Remove(ctx context.Context, userID, mediaID string) error {
	return uc.listRepo.Remove(ctx, userID, mediaID)
}

func (uc *ListUseCase) GetStatus(ctx context.Context, userID, externalID, source string) (*port.ListEntry, error) {
	return uc.listRepo.GetByUserAndExternalID(ctx, userID, externalID, source)
}
