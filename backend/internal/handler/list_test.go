package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/handler"
	"github.com/hlaclau/rate-it-api/internal/port"
)

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockListUseCase struct {
	addOrUpdateFn func(ctx context.Context, userID string, params port.AddOrUpdateParams) error
	updateFn      func(ctx context.Context, userID, mediaID string, status domain.UserMediaStatus, rating *int16, review *string) error
	getListFn     func(ctx context.Context, userID string) ([]port.ListEntry, error)
	removeFn      func(ctx context.Context, userID, mediaID string) error
	getStatusFn   func(ctx context.Context, userID, externalID, source string) (*port.ListEntry, error)
}

func (m *mockListUseCase) AddOrUpdate(ctx context.Context, userID string, params port.AddOrUpdateParams) error {
	return m.addOrUpdateFn(ctx, userID, params)
}
func (m *mockListUseCase) Update(ctx context.Context, userID, mediaID string, status domain.UserMediaStatus, rating *int16, review *string) error {
	return m.updateFn(ctx, userID, mediaID, status, rating, review)
}
func (m *mockListUseCase) GetList(ctx context.Context, userID string) ([]port.ListEntry, error) {
	return m.getListFn(ctx, userID)
}
func (m *mockListUseCase) Remove(ctx context.Context, userID, mediaID string) error {
	return m.removeFn(ctx, userID, mediaID)
}
func (m *mockListUseCase) GetStatus(ctx context.Context, userID, externalID, source string) (*port.ListEntry, error) {
	return m.getStatusFn(ctx, userID, externalID, source)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// newListRouter builds a router with RequireAuth applied, injecting userID into context.
func newListRouter(uc port.ListUseCase, userID string) http.Handler {
	r := chi.NewRouter()
	// Inject the user into context to simulate a valid session.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userID != "" {
				ctx := context.WithValue(r.Context(), handler.UserIDKey, userID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	})
	h := handler.NewListHandler(uc)
	h.Routes(r)
	return r
}

// newListRouterNoAuth builds a router WITHOUT injecting a user (unauthenticated).
func newListRouterNoAuth(uc port.ListUseCase) http.Handler {
	r := chi.NewRouter()
	h := handler.NewListHandler(uc)
	h.Routes(r)
	return r
}

func sampleListEntry() port.ListEntry {
	title := "Inception"
	return port.ListEntry{
		MediaID:    "m1",
		ExternalID: "550",
		Source:     domain.SourceTMDB,
		Type:       domain.TypeMovie,
		Title:      title,
		Status:     domain.StatusWatched,
		AddedAt:    time.Now(),
	}
}

// ── RequireAuth middleware ────────────────────────────────────────────────────

func TestListRoutes_Unauthenticated(t *testing.T) {
	mock := &mockListUseCase{}
	router := newListRouterNoAuth(mock)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/list"},
		{http.MethodPost, "/list"},
		{http.MethodPut, "/list/m1"},
		{http.MethodDelete, "/list/m1"},
		{http.MethodGet, "/list/status/ext1"},
	}

	for _, tc := range routes {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tc.method, tc.path, w.Code)
		}
	}
}

// ── GET /list ─────────────────────────────────────────────────────────────────

func TestGetList_Success(t *testing.T) {
	entries := []port.ListEntry{sampleListEntry()}
	mock := &mockListUseCase{
		getListFn: func(_ context.Context, _ string) ([]port.ListEntry, error) {
			return entries, nil
		},
	}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodGet, "/list", nil)
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

func TestGetList_Empty(t *testing.T) {
	mock := &mockListUseCase{
		getListFn: func(_ context.Context, _ string) ([]port.ListEntry, error) {
			return []port.ListEntry{}, nil
		},
	}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ── POST /list ────────────────────────────────────────────────────────────────

func TestAddOrUpdate_Success(t *testing.T) {
	mock := &mockListUseCase{
		addOrUpdateFn: func(_ context.Context, _ string, _ port.AddOrUpdateParams) error {
			return nil
		},
	}
	router := newListRouter(mock, "u1")

	body, _ := json.Marshal(map[string]any{
		"external_id": "550",
		"source":      "tmdb",
		"type":        "movie",
		"status":      "watched",
	})
	req := httptest.NewRequest(http.MethodPost, "/list", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestAddOrUpdate_MissingFields(t *testing.T) {
	mock := &mockListUseCase{}
	router := newListRouter(mock, "u1")

	cases := []map[string]any{
		{"source": "tmdb", "type": "movie", "status": "watched"},      // missing external_id
		{"external_id": "550", "type": "movie", "status": "watched"},  // missing source
		{"external_id": "550", "source": "tmdb", "status": "watched"}, // missing type
		{"external_id": "550", "source": "tmdb", "type": "movie"},     // missing status
	}

	for _, c := range cases {
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/list", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("payload %v: expected 400, got %d", c, w.Code)
		}
	}
}

func TestAddOrUpdate_InvalidBody(t *testing.T) {
	mock := &mockListUseCase{}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodPost, "/list", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAddOrUpdate_MediaNotFound(t *testing.T) {
	mock := &mockListUseCase{
		addOrUpdateFn: func(_ context.Context, _ string, _ port.AddOrUpdateParams) error {
			return port.ErrNotFound
		},
	}
	router := newListRouter(mock, "u1")

	body, _ := json.Marshal(map[string]any{
		"external_id": "9999",
		"source":      "tmdb",
		"type":        "movie",
		"status":      "watched",
	})
	req := httptest.NewRequest(http.MethodPost, "/list", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── PUT /list/{mediaID} ───────────────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	mock := &mockListUseCase{
		updateFn: func(_ context.Context, _, _ string, _ domain.UserMediaStatus, _ *int16, _ *string) error {
			return nil
		},
	}
	router := newListRouter(mock, "u1")

	body, _ := json.Marshal(map[string]any{"status": "watched"})
	req := httptest.NewRequest(http.MethodPut, "/list/m1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestUpdate_MissingStatus(t *testing.T) {
	mock := &mockListUseCase{}
	router := newListRouter(mock, "u1")

	body, _ := json.Marshal(map[string]any{"rating": 8})
	req := httptest.NewRequest(http.MethodPut, "/list/m1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	mock := &mockListUseCase{
		updateFn: func(_ context.Context, _, _ string, _ domain.UserMediaStatus, _ *int16, _ *string) error {
			return port.ErrUserMediaNotFound
		},
	}
	router := newListRouter(mock, "u1")

	body, _ := json.Marshal(map[string]any{"status": "watched"})
	req := httptest.NewRequest(http.MethodPut, "/list/unknown", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_InvalidBody(t *testing.T) {
	mock := &mockListUseCase{}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodPut, "/list/m1", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ── DELETE /list/{mediaID} ────────────────────────────────────────────────────

func TestRemove_Success(t *testing.T) {
	mock := &mockListUseCase{
		removeFn: func(_ context.Context, _, _ string) error { return nil },
	}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodDelete, "/list/m1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestRemove_NotFound(t *testing.T) {
	mock := &mockListUseCase{
		removeFn: func(_ context.Context, _, _ string) error { return port.ErrUserMediaNotFound },
	}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodDelete, "/list/ghost", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── GET /list/status/{externalID} ────────────────────────────────────────────

func TestGetStatus_Success(t *testing.T) {
	entry := sampleListEntry()
	mock := &mockListUseCase{
		getStatusFn: func(_ context.Context, _, _, _ string) (*port.ListEntry, error) {
			return &entry, nil
		},
	}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodGet, "/list/status/550?source=tmdb", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp port.ListEntry
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ExternalID != "550" {
		t.Errorf("unexpected external_id: %s", resp.ExternalID)
	}
}

func TestGetStatus_DefaultSource(t *testing.T) {
	var gotSource string
	mock := &mockListUseCase{
		getStatusFn: func(_ context.Context, _, _, source string) (*port.ListEntry, error) {
			gotSource = source
			entry := sampleListEntry()
			return &entry, nil
		},
	}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodGet, "/list/status/550", nil) // no ?source= param
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if gotSource != "tmdb" {
		t.Errorf("expected default source 'tmdb', got %q", gotSource)
	}
}

func TestGetStatus_NotFound(t *testing.T) {
	mock := &mockListUseCase{
		getStatusFn: func(_ context.Context, _, _, _ string) (*port.ListEntry, error) {
			return nil, port.ErrUserMediaNotFound
		},
	}
	router := newListRouter(mock, "u1")

	req := httptest.NewRequest(http.MethodGet, "/list/status/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
