package domain

import "time"

type UserMediaStatus string

const (
	StatusWatched     UserMediaStatus = "watched"
	StatusPlanToWatch UserMediaStatus = "plan_to_watch"
)

type UserMedia struct {
	UserID  string          `db:"user_id"`
	MediaID string          `db:"media_id"`
	Status  UserMediaStatus `db:"status"`
	Rating  *int16          `db:"rating"` // 1–10, nil if not rated
	AddedAt time.Time       `db:"added_at"`
}
