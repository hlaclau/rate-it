package handler_test

import (
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

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockUserUseCase struct {
	searchUsersFn   func(ctx context.Context, query string) ([]*domain.User, error)
	getByUsernameFn func(ctx context.Context, username string) (*domain.User, error)
}

func (m *mockUserUseCase) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
	return m.searchUsersFn(ctx, query)
}
func (m *mockUserUseCase) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return m.getByUsernameFn(ctx, username)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newUserRouter(userUC port.UserUseCase, listUC port.ListUseCase) http.Handler {
	r := chi.NewRouter()
	h := handler.NewUserHandler(userUC, listUC)
	h.Routes(r)
	return r
}

// ── GET /users/search ─────────────────────────────────────────────────────────

func TestUserSearch_ShortQuery_ReturnsEmpty(t *testing.T) {
	userUC := &mockUserUseCase{}
	router := newUserRouter(userUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/search?q=a", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty array, got %d items", len(resp))
	}
}

func TestUserSearch_ValidQuery_ReturnsUsers(t *testing.T) {
	userUC := &mockUserUseCase{
		searchUsersFn: func(_ context.Context, query string) ([]*domain.User, error) {
			return []*domain.User{{ID: "u1", Username: "alice"}}, nil
		},
	}
	router := newUserRouter(userUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/search?q=ali", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp))
	}
	if resp[0]["username"] != "alice" {
		t.Errorf("unexpected username: %v", resp[0]["username"])
	}
	// email must not be present in response
	if _, ok := resp[0]["email"]; ok {
		t.Error("email should not be in search response")
	}
}

func TestUserSearch_MissingQuery_ReturnsEmpty(t *testing.T) {
	router := newUserRouter(&mockUserUseCase{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty array, got %d items", len(resp))
	}
}

// ── GET /users/{username}/list ────────────────────────────────────────────────

func TestGetUserList_UserNotFound(t *testing.T) {
	userUC := &mockUserUseCase{
		getByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, port.ErrUserNotFound
		},
	}
	router := newUserRouter(userUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/ghost/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetUserList_Success(t *testing.T) {
	userUC := &mockUserUseCase{
		getByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "u1", Username: "alice"}, nil
		},
	}
	listUC := &mockListUseCase{
		getListFn: func(_ context.Context, userID string) ([]port.ListEntry, error) {
			if userID != "u1" {
				return nil, nil
			}
			return []port.ListEntry{sampleListEntry()}, nil
		},
	}
	router := newUserRouter(userUC, listUC)

	req := httptest.NewRequest(http.MethodGet, "/users/alice/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp []port.ListEntry
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp))
	}
}
