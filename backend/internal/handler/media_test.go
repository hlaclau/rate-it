package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/handler"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockMediaRepo struct {
	upsertFn func(m *domain.Media) error
}

func (m *mockMediaRepo) Upsert(media *domain.Media) error {
	if m.upsertFn != nil {
		return m.upsertFn(media)
	}
	return nil
}

type mockMediaFetcher struct {
	fetchMovieFn  func(id string) ([]byte, *domain.Media, error)
	fetchSeriesFn func(id string) ([]byte, *domain.Media, error)
	fetchMediaFn  func(params domain.MediaParams) ([]byte, error)
}

func (m *mockMediaFetcher) FetchMovie(id string) ([]byte, *domain.Media, error) {
	return m.fetchMovieFn(id)
}
func (m *mockMediaFetcher) FetchSeries(id string) ([]byte, *domain.Media, error) {
	return m.fetchSeriesFn(id)
}
func (m *mockMediaFetcher) FetchMedia(params domain.MediaParams) ([]byte, error) {
	return m.fetchMediaFn(params)
}

type mockMediaCache struct {
	getFn func(ctx context.Context, key string) ([]byte, error)
	setFn func(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

func (m *mockMediaCache) Get(ctx context.Context, key string) ([]byte, error) {
	return m.getFn(ctx, key)
}
func (m *mockMediaCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, key, value, ttl)
	}
	return nil
}

// ── Builders ──────────────────────────────────────────────────────────────────

func newMediaRouter(repo port.MediaRepository, fetcher port.MediaFetcher, cache port.MediaCache) http.Handler {
	uc := usecase.NewMediaUseCase(repo, fetcher, cache)
	h := handler.NewMediaHandler(uc)
	r := chi.NewRouter()
	h.Routes(r)
	return r
}

// alwaysMissCache returns ErrCacheMiss on Get and no error on Set.
func alwaysMissCache() port.MediaCache {
	return &mockMediaCache{
		getFn: func(_ context.Context, _ string) ([]byte, error) {
			return nil, port.ErrCacheMiss
		},
	}
}

func sampleMoviePayload() []byte { return []byte(`{"id":550,"title":"Fight Club"}`) }
func sampleMedia(id string) *domain.Media {
	return &domain.Media{ExternalID: id, Source: domain.SourceTMDB, Type: domain.TypeMovie, Title: "Fight Club"}
}

// ── GET /media/movie/{id} ─────────────────────────────────────────────────────

func TestGetMovie_CacheHit(t *testing.T) {
	cache := &mockMediaCache{
		getFn: func(_ context.Context, _ string) ([]byte, error) {
			return sampleMoviePayload(), nil
		},
	}
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, cache)

	req := httptest.NewRequest(http.MethodGet, "/media/movie/550", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != string(sampleMoviePayload()) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestGetMovie_CacheMiss_FetchSuccess(t *testing.T) {
	fetcher := &mockMediaFetcher{
		fetchMovieFn: func(id string) ([]byte, *domain.Media, error) {
			return sampleMoviePayload(), sampleMedia(id), nil
		},
	}
	router := newMediaRouter(&mockMediaRepo{}, fetcher, alwaysMissCache())

	req := httptest.NewRequest(http.MethodGet, "/media/movie/550", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetMovie_NotFound(t *testing.T) {
	fetcher := &mockMediaFetcher{
		fetchMovieFn: func(_ string) ([]byte, *domain.Media, error) {
			return nil, nil, port.ErrNotFound
		},
	}
	router := newMediaRouter(&mockMediaRepo{}, fetcher, alwaysMissCache())

	req := httptest.NewRequest(http.MethodGet, "/media/movie/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── GET /media/series/{id} ────────────────────────────────────────────────────

func TestGetSeries_Success(t *testing.T) {
	payload := []byte(`{"id":1399,"name":"Game of Thrones"}`)
	fetcher := &mockMediaFetcher{
		fetchSeriesFn: func(id string) ([]byte, *domain.Media, error) {
			return payload, sampleMedia(id), nil
		},
	}
	router := newMediaRouter(&mockMediaRepo{}, fetcher, alwaysMissCache())

	req := httptest.NewRequest(http.MethodGet, "/media/series/1399", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetSeries_NotFound(t *testing.T) {
	fetcher := &mockMediaFetcher{
		fetchSeriesFn: func(_ string) ([]byte, *domain.Media, error) {
			return nil, nil, port.ErrNotFound
		},
	}
	router := newMediaRouter(&mockMediaRepo{}, fetcher, alwaysMissCache())

	req := httptest.NewRequest(http.MethodGet, "/media/series/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── GET /media/search ─────────────────────────────────────────────────────────

func TestSearchMedia_Success(t *testing.T) {
	payload := []byte(`{"results":[]}`)
	fetcher := &mockMediaFetcher{
		fetchMediaFn: func(_ domain.MediaParams) ([]byte, error) {
			return payload, nil
		},
	}
	router := newMediaRouter(&mockMediaRepo{}, fetcher, alwaysMissCache())

	req := httptest.NewRequest(http.MethodGet, "/media/search?q=inception&type=movie", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSearchMedia_InvalidPage(t *testing.T) {
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, alwaysMissCache())

	for _, page := range []string{"abc", "0", "-1", "1001"} {
		req := httptest.NewRequest(http.MethodGet, "/media/search?page="+page, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("page=%s: expected 400, got %d", page, w.Code)
		}
	}
}

func TestSearchMedia_InvalidYearFrom(t *testing.T) {
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, alwaysMissCache())
	req := httptest.NewRequest(http.MethodGet, "/media/search?year_from=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSearchMedia_InvalidYearTo(t *testing.T) {
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, alwaysMissCache())
	req := httptest.NewRequest(http.MethodGet, "/media/search?year_to=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSearchMedia_InvalidVoteAverageMin(t *testing.T) {
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, alwaysMissCache())
	for _, v := range []string{"abc", "-1", "11"} {
		req := httptest.NewRequest(http.MethodGet, "/media/search?vote_average_min="+v, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("vote_average_min=%s: expected 400, got %d", v, w.Code)
		}
	}
}

func TestSearchMedia_InvalidVoteAverageMax(t *testing.T) {
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, alwaysMissCache())
	for _, v := range []string{"abc", "-1", "11"} {
		req := httptest.NewRequest(http.MethodGet, "/media/search?vote_average_max="+v, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("vote_average_max=%s: expected 400, got %d", v, w.Code)
		}
	}
}

func TestSearchMedia_InvalidSortBy(t *testing.T) {
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, alwaysMissCache())
	req := httptest.NewRequest(http.MethodGet, "/media/search?sort_by=invalid_sort", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSearchMedia_ValidSortBy(t *testing.T) {
	payload := []byte(`{"results":[]}`)
	fetcher := &mockMediaFetcher{
		fetchMediaFn: func(_ domain.MediaParams) ([]byte, error) { return payload, nil },
	}
	router := newMediaRouter(&mockMediaRepo{}, fetcher, alwaysMissCache())

	validSorts := []string{
		"popularity.desc", "popularity.asc",
		"vote_average.desc", "vote_average.asc",
		"release_date.desc", "release_date.asc",
	}
	for _, s := range validSorts {
		req := httptest.NewRequest(http.MethodGet, "/media/search?sort_by="+s, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("sort_by=%s: expected 200, got %d", s, w.Code)
		}
	}
}

func TestSearchMedia_InvalidVoteCountMin(t *testing.T) {
	router := newMediaRouter(&mockMediaRepo{}, &mockMediaFetcher{}, alwaysMissCache())
	for _, v := range []string{"abc", "-5"} {
		req := httptest.NewRequest(http.MethodGet, "/media/search?vote_count_min="+v, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("vote_count_min=%s: expected 400, got %d", v, w.Code)
		}
	}
}

func TestSearchMedia_DefaultPage(t *testing.T) {
	var gotPage int
	fetcher := &mockMediaFetcher{
		fetchMediaFn: func(p domain.MediaParams) ([]byte, error) {
			gotPage = p.Page
			return []byte(`{}`), nil
		},
	}
	router := newMediaRouter(&mockMediaRepo{}, fetcher, alwaysMissCache())
	req := httptest.NewRequest(http.MethodGet, "/media/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if gotPage != 1 {
		t.Errorf("expected default page=1, got %d", gotPage)
	}
}
