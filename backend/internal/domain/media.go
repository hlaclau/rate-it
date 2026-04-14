package domain

import "time"

type MediaSource string
type MediaType string

const (
	SourceTMDB        MediaSource = "tmdb"
)

const (
	TypeMovie  MediaType = "movie"
	TypeSeries MediaType = "series"
	TypeBook   MediaType = "book"
)

type Media struct {
	ID          string      `db:"id"`
	ExternalID  string      `db:"external_id"`
	Source      MediaSource `db:"source"`
	Type        MediaType   `db:"type"`
	Title       string      `db:"title"`
	PosterPath  *string     `db:"poster_path"`
	ReleaseYear *int16      `db:"release_year"`
	CachedAt    time.Time   `db:"cached_at"`
}
