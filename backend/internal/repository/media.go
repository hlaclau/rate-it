package repository

import (
	"fmt"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/jmoiron/sqlx"
)

// MediaRepository implements port.MediaRepository.
type MediaRepository struct {
	db *sqlx.DB
}

func NewMediaRepository(db *sqlx.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Upsert(m *domain.Media) error {
	const q = `
		INSERT INTO media (external_id, source, type, title, poster_path, release_year)
		VALUES (:external_id, :source, :type, :title, :poster_path, :release_year)
		ON CONFLICT (external_id, source)
		DO UPDATE SET
			title        = EXCLUDED.title,
			poster_path  = EXCLUDED.poster_path,
			release_year = EXCLUDED.release_year,
			cached_at    = NOW()
		RETURNING id, cached_at`

	rows, err := r.db.NamedQuery(q, m)
	if err != nil {
		return fmt.Errorf("upsert media: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&m.ID, &m.CachedAt); err != nil {
			return fmt.Errorf("scan upsert result: %w", err)
		}
	}
	return rows.Err()
}
