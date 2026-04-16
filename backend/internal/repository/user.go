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
		// Check for unique constraint violation (PostgreSQL)
		// Usually we'd check the error code, but for simplicity:
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
