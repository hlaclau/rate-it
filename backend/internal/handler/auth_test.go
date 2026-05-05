package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/handler"
	"github.com/hlaclau/rate-it-api/internal/port"
)

// ── Mock ────────────────────────────────────────────────────────────────────

type mockAuthUseCase struct {
	registerFn           func(ctx context.Context, username, email, password string) (*domain.User, error)
	loginFn              func(ctx context.Context, email, password string) (string, string, *domain.User, error)
	refreshFn            func(ctx context.Context, refreshToken string) (string, string, error)
	validateAccessTokenFn func(tokenString string) (string, error)
	getUserByIDFn        func(ctx context.Context, id string) (*domain.User, error)
	logoutFn             func(ctx context.Context, userID string) error
}

func (m *mockAuthUseCase) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	return m.registerFn(ctx, username, email, password)
}
func (m *mockAuthUseCase) Login(ctx context.Context, email, password string) (string, string, *domain.User, error) {
	return m.loginFn(ctx, email, password)
}
func (m *mockAuthUseCase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	return m.refreshFn(ctx, refreshToken)
}
func (m *mockAuthUseCase) ValidateAccessToken(tokenString string) (string, error) {
	return m.validateAccessTokenFn(tokenString)
}
func (m *mockAuthUseCase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return m.getUserByIDFn(ctx, id)
}
func (m *mockAuthUseCase) Logout(ctx context.Context, userID string) error {
	return m.logoutFn(ctx, userID)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func newAuthRouter(uc port.AuthUseCase) http.Handler {
	r := chi.NewRouter()
	h := handler.NewAuthHandler(uc)
	h.Routes(r)
	return r
}

func postJSON(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func sampleUser() *domain.User {
	return &domain.User{ID: "u1", Username: "alice", Email: "alice@example.com", PasswordHash: "hashed"}
}

// ── POST /register ────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	mock := &mockAuthUseCase{
		registerFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
			return sampleUser(), nil
		},
	}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "Passw0rd",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp domain.User
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.PasswordHash != "" {
		t.Error("password_hash must be stripped from response")
	}
}

func TestRegister_MissingUsername(t *testing.T) {
	mock := &mockAuthUseCase{}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"email":    "alice@example.com",
		"password": "Passw0rd",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	mock := &mockAuthUseCase{}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"username": "alice",
		"email":    "not-an-email",
		"password": "Passw0rd",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegister_WeakPassword_TooShort(t *testing.T) {
	mock := &mockAuthUseCase{}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "Ab1",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegister_WeakPassword_NoUppercase(t *testing.T) {
	mock := &mockAuthUseCase{}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "passw0rd",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegister_WeakPassword_NoDigit(t *testing.T) {
	mock := &mockAuthUseCase{}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "Password",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mock := &mockAuthUseCase{
		registerFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
			return nil, port.ErrEmailAlreadyExists
		},
	}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "Passw0rd",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestRegister_UsernameAlreadyExists(t *testing.T) {
	mock := &mockAuthUseCase{
		registerFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
			return nil, port.ErrUsernameAlreadyExists
		},
	}
	w := postJSON(t, newAuthRouter(mock), "/register", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "Passw0rd",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestRegister_InvalidBody(t *testing.T) {
	mock := &mockAuthUseCase{}
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ── POST /login ───────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	mock := &mockAuthUseCase{
		loginFn: func(_ context.Context, _, _ string) (string, string, *domain.User, error) {
			return "access_tok", "refresh_tok", sampleUser(), nil
		},
	}
	w := postJSON(t, newAuthRouter(mock), "/login", map[string]string{
		"email":    "alice@example.com",
		"password": "Passw0rd",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	var hasAccess, hasRefresh bool
	for _, c := range cookies {
		if c.Name == "access_token" {
			hasAccess = true
		}
		if c.Name == "refresh_token" {
			hasRefresh = true
		}
	}
	if !hasAccess || !hasRefresh {
		t.Error("expected both access_token and refresh_token cookies to be set")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mock := &mockAuthUseCase{
		loginFn: func(_ context.Context, _, _ string) (string, string, *domain.User, error) {
			return "", "", nil, port.ErrInvalidCredentials
		},
	}
	w := postJSON(t, newAuthRouter(mock), "/login", map[string]string{
		"email":    "alice@example.com",
		"password": "WrongPass1",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_MissingEmail(t *testing.T) {
	mock := &mockAuthUseCase{}
	w := postJSON(t, newAuthRouter(mock), "/login", map[string]string{
		"password": "Passw0rd",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	mock := &mockAuthUseCase{}
	w := postJSON(t, newAuthRouter(mock), "/login", map[string]string{
		"email": "alice@example.com",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ── POST /refresh ─────────────────────────────────────────────────────────────

func TestRefresh_Success(t *testing.T) {
	mock := &mockAuthUseCase{
		refreshFn: func(_ context.Context, _ string) (string, string, error) {
			return "new_access", "u1", nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid_refresh"})
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRefresh_NoCookie(t *testing.T) {
	mock := &mockAuthUseCase{}
	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	mock := &mockAuthUseCase{
		refreshFn: func(_ context.Context, _ string) (string, string, error) {
			return "", "", port.ErrUnauthorized
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad_token"})
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// ── POST /logout ──────────────────────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	called := false
	mock := &mockAuthUseCase{
		logoutFn: func(_ context.Context, _ string) error {
			called = true
			return nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	ctx := context.WithValue(req.Context(), handler.UserIDKey, "u1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if !called {
		t.Error("expected logout usecase to be called")
	}
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if (c.Name == "access_token" || c.Name == "refresh_token") && c.MaxAge != -1 {
			t.Errorf("cookie %s should be expired (MaxAge=-1)", c.Name)
		}
	}
}

func TestLogout_WithoutSession(t *testing.T) {
	// Logout without a user in context should still clear cookies and return 204.
	mock := &mockAuthUseCase{}
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

// ── GET /me ───────────────────────────────────────────────────────────────────

func TestMe_Success(t *testing.T) {
	mock := &mockAuthUseCase{
		getUserByIDFn: func(_ context.Context, id string) (*domain.User, error) {
			return sampleUser(), nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx := context.WithValue(req.Context(), handler.UserIDKey, "u1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp domain.User
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.PasswordHash != "" {
		t.Error("password_hash must be stripped from /me response")
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	mock := &mockAuthUseCase{}
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMe_UserNotFound(t *testing.T) {
	mock := &mockAuthUseCase{
		getUserByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, port.ErrUserNotFound
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx := context.WithValue(req.Context(), handler.UserIDKey, "ghost")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	newAuthRouter(mock).ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
