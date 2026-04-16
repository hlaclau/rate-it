package port

import (
	"context"
	"errors"
	"time"

	"github.com/hlaclau/rate-it-api/internal/domain"
)

var ErrUserMediaNotFound = errors.New("list entry not found")

// ListEntry is a read-projection joining user_media with media.
type ListEntry struct {
	MediaID     string                 `db:"media_id"    json:"media_id"`
	ExternalID  string                 `db:"external_id" json:"external_id"`
	Source      domain.MediaSource     `db:"source"      json:"source"`
	Type        domain.MediaType       `db:"type"        json:"type"`
	Title       string                 `db:"title"       json:"title"`
	PosterPath  *string                `db:"poster_path" json:"poster_path"`
	ReleaseYear *int16                 `db:"release_year" json:"release_year"`
	Status      domain.UserMediaStatus `db:"status"      json:"status"`
	Rating      *int16                 `db:"rating"      json:"rating"`
	Review      *string                `db:"review"      json:"review"`
	AddedAt     time.Time              `db:"added_at"    json:"added_at"`
}

// AddOrUpdateParams carries the data needed to add or update a list entry.
type AddOrUpdateParams struct {
	ExternalID string
	Source     domain.MediaSource
	Type       domain.MediaType
	Status     domain.UserMediaStatus
	Rating     *int16
	Review     *string
}

// UserMediaRepository persists user list entries.
type UserMediaRepository interface {
	Upsert(ctx context.Context, um *domain.UserMedia) error
	GetByUserAndMedia(ctx context.Context, userID, mediaID string) (*domain.UserMedia, error)
	ListByUser(ctx context.Context, userID string) ([]ListEntry, error)
	Remove(ctx context.Context, userID, mediaID string) error
	GetByUserAndExternalID(ctx context.Context, userID, externalID, source string) (*ListEntry, error)
}

// ListUseCase manages the user's media list.
type ListUseCase interface {
	AddOrUpdate(ctx context.Context, userID string, params AddOrUpdateParams) error
	Update(ctx context.Context, userID, mediaID string, status domain.UserMediaStatus, rating *int16, review *string) error
	GetList(ctx context.Context, userID string) ([]ListEntry, error)
	Remove(ctx context.Context, userID, mediaID string) error
	GetStatus(ctx context.Context, userID, externalID, source string) (*ListEntry, error)
}
