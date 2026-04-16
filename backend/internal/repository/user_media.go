package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/jmoiron/sqlx"
)

// UserMediaRepository implements port.UserMediaRepository.
type UserMediaRepository struct {
	db *sqlx.DB
}

func NewUserMediaRepository(db *sqlx.DB) *UserMediaRepository {
	return &UserMediaRepository{db: db}
}

func (r *UserMediaRepository) Upsert(ctx context.Context, um *domain.UserMedia) error {
	const q = `
		INSERT INTO user_media (user_id, media_id, status, rating, review)
		VALUES (:user_id, :media_id, :status, :rating, :review)
		ON CONFLICT (user_id, media_id) DO UPDATE SET
			status = EXCLUDED.status,
			rating = EXCLUDED.rating,
			review = EXCLUDED.review
		RETURNING added_at`

	rows, err := r.db.NamedQueryContext(ctx, q, um)
	if err != nil {
		return fmt.Errorf("upsert user_media: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&um.AddedAt); err != nil {
			return fmt.Errorf("scan upsert result: %w", err)
		}
	}
	return rows.Err()
}

func (r *UserMediaRepository) GetByUserAndMedia(ctx context.Context, userID, mediaID string) (*domain.UserMedia, error) {
	const q = `
		SELECT user_id, media_id, status, rating, review, added_at
		FROM user_media
		WHERE user_id = $1 AND media_id = $2`

	var um domain.UserMedia
	if err := r.db.GetContext(ctx, &um, q, userID, mediaID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, port.ErrUserMediaNotFound
		}
		return nil, fmt.Errorf("get user_media: %w", err)
	}
	return &um, nil
}

func (r *UserMediaRepository) ListByUser(ctx context.Context, userID string) ([]port.ListEntry, error) {
	const q = `
		SELECT
			m.id           AS media_id,
			m.external_id,
			m.source,
			m.type,
			m.title,
			m.poster_path,
			m.release_year,
			um.status,
			um.rating,
			um.review,
			um.added_at
		FROM user_media um
		JOIN media m ON m.id = um.media_id
		WHERE um.user_id = $1
		ORDER BY um.added_at DESC`

	var entries []port.ListEntry
	if err := r.db.SelectContext(ctx, &entries, q, userID); err != nil {
		return nil, fmt.Errorf("list user_media: %w", err)
	}
	return entries, nil
}

func (r *UserMediaRepository) Remove(ctx context.Context, userID, mediaID string) error {
	const q = `DELETE FROM user_media WHERE user_id = $1 AND media_id = $2`

	res, err := r.db.ExecContext(ctx, q, userID, mediaID)
	if err != nil {
		return fmt.Errorf("remove user_media: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return port.ErrUserMediaNotFound
	}
	return nil
}

func (r *UserMediaRepository) GetByUserAndExternalID(ctx context.Context, userID, externalID, source string) (*port.ListEntry, error) {
	const q = `
		SELECT
			m.id           AS media_id,
			m.external_id,
			m.source,
			m.type,
			m.title,
			m.poster_path,
			m.release_year,
			um.status,
			um.rating,
			um.review,
			um.added_at
		FROM user_media um
		JOIN media m ON m.id = um.media_id
		WHERE um.user_id = $1 AND m.external_id = $2 AND m.source = $3`

	var entry port.ListEntry
	if err := r.db.GetContext(ctx, &entry, q, userID, externalID, source); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, port.ErrUserMediaNotFound
		}
		return nil, fmt.Errorf("get user_media by external id: %w", err)
	}
	return &entry, nil
}
