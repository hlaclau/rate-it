package usecase_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockUserRepository struct {
	createFn     func(ctx context.Context, user *domain.User) error
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn    func(ctx context.Context, id string) (*domain.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}
func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, port.ErrUserNotFound
}
func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, port.ErrUserNotFound
}

type mockSessionRepository struct {
	setFn    func(ctx context.Context, userID, refreshToken string, ttl time.Duration) error
	getFn    func(ctx context.Context, userID string) (string, error)
	deleteFn func(ctx context.Context, userID string) error
}

func (m *mockSessionRepository) Set(ctx context.Context, userID, refreshToken string, ttl time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, userID, refreshToken, ttl)
	}
	return nil
}
func (m *mockSessionRepository) Get(ctx context.Context, userID string) (string, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID)
	}
	return "", port.ErrUnauthorized
}
func (m *mockSessionRepository) Delete(ctx context.Context, userID string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID)
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func setupAuthUseCase() (*usecase.AuthUseCase, *mockUserRepository, *mockSessionRepository) {
	_ = os.Setenv("JWT_SECRET", "test-secret")
	_ = os.Setenv("JWT_ACCESS_EXPIRATION", "5m")
	_ = os.Setenv("JWT_REFRESH_EXPIRATION", "10m")

	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	uc := usecase.NewAuthUseCase(userRepo, sessionRepo)
	return uc, userRepo, sessionRepo
}

func hashPassword(t *testing.T, pwd string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestAuthUseCase_Register_Success(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase()

	userRepo.createFn = func(ctx context.Context, user *domain.User) error {
		user.ID = "u1"
		return nil
	}

	user, err := uc.Register(context.Background(), "alice", "alice@test.com", "Password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Username != "alice" || user.Email != "alice@test.com" {
		t.Errorf("unexpected user details: %+v", user)
	}

	// Verify password was hashed
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("Password123")); err != nil {
		t.Errorf("password was not hashed correctly: %v", err)
	}
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	uc, userRepo, sessionRepo := setupAuthUseCase()

	pwdHash := hashPassword(t, "MySecret")
	userRepo.getByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return &domain.User{ID: "u1", Email: email, PasswordHash: pwdHash}, nil
	}

	sessionRepo.setFn = func(ctx context.Context, userID, refreshToken string, ttl time.Duration) error {
		if userID != "u1" {
			t.Errorf("unexpected userID in session: %s", userID)
		}
		return nil
	}

	access, refresh, user, err := uc.Login(context.Background(), "alice@test.com", "MySecret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("expected non-empty tokens")
	}
	if user.ID != "u1" {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestAuthUseCase_Login_InvalidPassword(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase()

	pwdHash := hashPassword(t, "MySecret")
	userRepo.getByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return &domain.User{ID: "u1", Email: email, PasswordHash: pwdHash}, nil
	}

	_, _, _, err := uc.Login(context.Background(), "alice@test.com", "WrongPassword")
	if err != port.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUseCase_Login_UserNotFound(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase()

	userRepo.getByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return nil, port.ErrUserNotFound
	}

	_, _, _, err := uc.Login(context.Background(), "unknown@test.com", "MySecret")
	if err != port.ErrInvalidCredentials { // AuthUseCase obscures ErrUserNotFound
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUseCase_Logout(t *testing.T) {
	uc, _, sessionRepo := setupAuthUseCase()

	called := false
	sessionRepo.deleteFn = func(ctx context.Context, userID string) error {
		called = true
		return nil
	}

	err := uc.Logout(context.Background(), "u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected sessionRepo.Delete to be called")
	}
}

func TestAuthUseCase_Refresh_Success(t *testing.T) {
	uc, userRepo, sessionRepo := setupAuthUseCase()

	// Generate a valid refresh token using the UseCase
	validRefresh, err := uc.GenerateRefreshToken("u1")
	if err != nil {
		t.Fatalf("failed to generate test refresh token: %v", err)
	}

	sessionRepo.getFn = func(ctx context.Context, userID string) (string, error) {
		return validRefresh, nil // Match the token in cache
	}
	userRepo.getByIDFn = func(ctx context.Context, id string) (*domain.User, error) {
		return &domain.User{ID: "u1", Email: "alice@test.com"}, nil
	}

	newAccess, userID, err := uc.Refresh(context.Background(), validRefresh)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if newAccess == "" {
		t.Error("expected new access token")
	}
	if userID != "u1" {
		t.Errorf("unexpected userID returned: %s", userID)
	}
}

func TestAuthUseCase_Refresh_InvalidToken(t *testing.T) {
	uc, _, _ := setupAuthUseCase()

	_, _, err := uc.Refresh(context.Background(), "invalid-token-string")
	if err != port.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthUseCase_Refresh_TokenMismatch(t *testing.T) {
	uc, _, sessionRepo := setupAuthUseCase()

	validRefresh, _ := uc.GenerateRefreshToken("u1")

	sessionRepo.getFn = func(ctx context.Context, userID string) (string, error) {
		return "different-token-in-redis", nil
	}

	_, _, err := uc.Refresh(context.Background(), validRefresh)
	if err != port.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthUseCase_ValidateAccessToken_Success(t *testing.T) {
	uc, _, _ := setupAuthUseCase()

	// Simulate login to get an access token
	os.Setenv("JWT_SECRET", "test-secret")
	// Since generateAccessToken is private, we'll forge one using jwt directly
	claims := jwt.MapClaims{
		"sub": "u1",
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	userID, err := uc.ValidateAccessToken(tokenString)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != "u1" {
		t.Errorf("expected u1, got %s", userID)
	}
}

func TestAuthUseCase_ValidateAccessToken_Expired(t *testing.T) {
	uc, _, _ := setupAuthUseCase()

	claims := jwt.MapClaims{
		"sub": "u1",
		"exp": time.Now().Add(-5 * time.Minute).Unix(), // Expired 5 minutes ago
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	_, err := uc.ValidateAccessToken(tokenString)
	if err != port.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
