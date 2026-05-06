package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, q, u.Username, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "users_email_key":
				return port.ErrEmailAlreadyExists
			case "users_username_key":
				return port.ErrUsernameAlreadyExists
			default:
				return port.ErrUserAlreadyExists
			}
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `SELECT id, username, email, password_hash, bio, avatar_url, created_at FROM users WHERE email = $1`

	var u domain.User
	err := r.db.GetContext(ctx, &u, q, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, port.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `SELECT id, username, email, password_hash, bio, avatar_url, created_at FROM users WHERE id = $1`

	var u domain.User
	err := r.db.GetContext(ctx, &u, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, port.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	const q = `SELECT id, username, email, password_hash, bio, avatar_url, created_at FROM users WHERE username = $1`

	var u domain.User
	err := r.db.GetContext(ctx, &u, q, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, port.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) SearchByUsername(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	const q = `
		SELECT id, username, email, password_hash, bio, avatar_url, created_at
		FROM users
		WHERE username ILIKE '%' || $1 || '%'
		ORDER BY username
		LIMIT $2`

	var users []*domain.User
	err := r.db.SelectContext(ctx, &users, q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search users by username: %w", err)
	}

	return users, nil
}
