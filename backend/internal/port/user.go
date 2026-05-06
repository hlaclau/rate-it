package port

import (
	"context"
	"errors"

	"github.com/hlaclau/rate-it-api/internal/domain"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrEmailAlreadyExists    = errors.New("email already in use")
	ErrUsernameAlreadyExists = errors.New("username already taken")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUnauthorized          = errors.New("unauthorized")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	SearchByUsername(ctx context.Context, query string, limit int) ([]*domain.User, error)
}

type AuthUseCase interface {
	Register(ctx context.Context, username, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, user *domain.User, err error)
	Refresh(ctx context.Context, refreshToken string) (accessToken string, userID string, err error)
	ValidateAccessToken(tokenString string) (userID string, err error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	Logout(ctx context.Context, userID string) error
}

type UserUseCase interface {
	SearchUsers(ctx context.Context, query string) ([]*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}
