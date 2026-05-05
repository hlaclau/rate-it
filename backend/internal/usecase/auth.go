package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo    port.UserRepository
	sessionRepo port.SessionRepository
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewAuthUseCase(userRepo port.UserRepository, sessionRepo port.SessionRepository) *AuthUseCase {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-me"
	}

	accessDur, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRATION"))
	if accessDur == 0 {
		accessDur = 15 * time.Minute
	}

	refreshDur, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_EXPIRATION"))
	if refreshDur == 0 {
		refreshDur = 7 * 24 * time.Hour
	}

	return &AuthUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		secret:      []byte(secret),
		accessTTL:   accessDur,
		refreshTTL:  refreshDur,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}

	if err = uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (string, string, *domain.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", nil, port.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, port.ErrInvalidCredentials
	}

	accessToken, err := uc.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := uc.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	if err = uc.sessionRepo.Set(ctx, user.ID, refreshToken, uc.refreshTTL); err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, userID string) error {
	return uc.sessionRepo.Delete(ctx, userID)
}

func (uc *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	// In a more complex setup, we'd use a JWT for refresh tokens too to store userID inside.
	// Since we are enforcing single-device, we need the userID.
	// We'll trust the refresh_token value but we need a way to know WHICH user it belongs to
	// if we don't send the userID.
	// Actually, let's keep the refresh token as a JWT with userID claim to avoid extra roundtrips or complex keys.
	// NO - user requested a stable School project style.

	// Let's modify the plan slightly: if RefreshToken is just a UUID, we need to know for WHICH user.
	// Usually, we'd send the userID as well or the refresh token would be a JWT.
	// I'll use a JWT for the refresh token to keep it simple (it contains the userID).

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return uc.secret, nil
	})

	if err != nil || !token.Valid {
		return "", "", port.ErrUnauthorized
	}

	userID := claims.Subject
	storedToken, err := uc.sessionRepo.Get(ctx, userID)
	if err != nil || storedToken != refreshToken {
		return "", "", port.ErrUnauthorized
	}

	// Get user to get the email
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", port.ErrUnauthorized
	}

	newAccessToken, err := uc.generateAccessToken(user.ID, user.Email)
	return newAccessToken, user.ID, err
}

func (uc *AuthUseCase) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return uc.secret, nil
	})

	if err != nil || !token.Valid {
		return "", port.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", port.ErrUnauthorized
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", port.ErrUnauthorized
	}

	return userID, nil
}

func (uc *AuthUseCase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *AuthUseCase) generateAccessToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(uc.accessTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.secret)
}

// GenerateRefreshToken is called during login
func (uc *AuthUseCase) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.refreshTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.secret)
}
