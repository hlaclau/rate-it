package domain

import "time"

type User struct {
	ID           string    `db:"id"`
	Username     string    `db:"username"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Bio          *string   `db:"bio"`
	AvatarURL    *string   `db:"avatar_url"`
	CreatedAt    time.Time `db:"created_at"`
}
