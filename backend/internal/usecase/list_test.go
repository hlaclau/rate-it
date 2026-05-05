package usecase_test

import (
	"context"
	"testing"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockListRepo struct {
	upsertFn                 func(ctx context.Context, um *domain.UserMedia) error
	getByUserAndMediaFn      func(ctx context.Context, userID, mediaID string) (*domain.UserMedia, error)
	listByUserFn             func(ctx context.Context, userID string) ([]port.ListEntry, error)
	removeFn                 func(ctx context.Context, userID, mediaID string) error
	getByUserAndExternalIDFn func(ctx context.Context, userID, externalID, source string) (*port.ListEntry, error)
}

func (m *mockListRepo) Upsert(ctx context.Context, um *domain.UserMedia) error {
	return m.upsertFn(ctx, um)
}
func (m *mockListRepo) GetByUserAndMedia(ctx context.Context, userID, mediaID string) (*domain.UserMedia, error) {
	if m.getByUserAndMediaFn != nil {
		return m.getByUserAndMediaFn(ctx, userID, mediaID)
	}
	return nil, port.ErrUserMediaNotFound
}
func (m *mockListRepo) ListByUser(ctx context.Context, userID string) ([]port.ListEntry, error) {
	return m.listByUserFn(ctx, userID)
}
func (m *mockListRepo) Remove(ctx context.Context, userID, mediaID string) error {
	return m.removeFn(ctx, userID, mediaID)
}
func (m *mockListRepo) GetByUserAndExternalID(ctx context.Context, userID, externalID, source string) (*port.ListEntry, error) {
	return m.getByUserAndExternalIDFn(ctx, userID, externalID, source)
}

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
	if m.fetchMovieFn != nil {
		return m.fetchMovieFn(id)
	}
	return nil, nil, nil
}
func (m *mockMediaFetcher) FetchSeries(id string) ([]byte, *domain.Media, error) {
	if m.fetchSeriesFn != nil {
		return m.fetchSeriesFn(id)
	}
	return nil, nil, nil
}
func (m *mockMediaFetcher) FetchMedia(params domain.MediaParams) ([]byte, error) {
	if m.fetchMediaFn != nil {
		return m.fetchMediaFn(params)
	}
	return nil, nil
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestListUseCase_AddOrUpdate_InvalidRating(t *testing.T) {
	uc := usecase.NewListUseCase(&mockListRepo{}, &mockMediaRepo{}, &mockMediaFetcher{})
	rating := int16(8)
	err := uc.AddOrUpdate(context.Background(), "u1", port.AddOrUpdateParams{
		Status: domain.StatusPlanToWatch, // Rating not allowed for plan_to_watch
		Rating: &rating,
	})
	if err != usecase.ErrInvalidRating {
		t.Fatalf("expected ErrInvalidRating, got %v", err)
	}
}

func TestListUseCase_AddOrUpdate_Success(t *testing.T) {
	ctx := context.Background()
	fetcher := &mockMediaFetcher{
		fetchMovieFn: func(id string) ([]byte, *domain.Media, error) {
			return nil, &domain.Media{ID: "m1", Type: domain.TypeMovie}, nil
		},
	}
	mediaRepo := &mockMediaRepo{
		upsertFn: func(m *domain.Media) error { return nil },
	}
	listRepo := &mockListRepo{
		upsertFn: func(ctx context.Context, um *domain.UserMedia) error {
			if um.UserID != "u1" || um.MediaID != "m1" || um.Status != domain.StatusWatched {
				t.Errorf("unexpected user media values: %+v", um)
			}
			return nil
		},
	}

	uc := usecase.NewListUseCase(listRepo, mediaRepo, fetcher)

	err := uc.AddOrUpdate(ctx, "u1", port.AddOrUpdateParams{
		ExternalID: "550",
		Source:     domain.SourceTMDB,
		Type:       domain.TypeMovie,
		Status:     domain.StatusWatched,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListUseCase_AddOrUpdate_UnsupportedType(t *testing.T) {
	uc := usecase.NewListUseCase(&mockListRepo{}, &mockMediaRepo{}, &mockMediaFetcher{})
	err := uc.AddOrUpdate(context.Background(), "u1", port.AddOrUpdateParams{
		Type: "unknown_type",
	})
	if err == nil || err.Error() != "unsupported media type: unknown_type" {
		t.Fatalf("expected unsupported media type error, got %v", err)
	}
}

func TestListUseCase_Update_Success(t *testing.T) {
	listRepo := &mockListRepo{
		upsertFn: func(ctx context.Context, um *domain.UserMedia) error { return nil },
	}
	uc := usecase.NewListUseCase(listRepo, nil, nil)

	rating := int16(9)
	err := uc.Update(context.Background(), "u1", "m1", domain.StatusWatched, &rating, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListUseCase_Update_InvalidRating(t *testing.T) {
	uc := usecase.NewListUseCase(&mockListRepo{}, &mockMediaRepo{}, &mockMediaFetcher{})
	rating := int16(9)
	err := uc.Update(context.Background(), "u1", "m1", domain.StatusPlanToWatch, &rating, nil)
	if err != usecase.ErrInvalidRating {
		t.Fatalf("expected ErrInvalidRating, got %v", err)
	}
}

func TestListUseCase_GetList(t *testing.T) {
	listRepo := &mockListRepo{
		listByUserFn: func(ctx context.Context, userID string) ([]port.ListEntry, error) {
			return []port.ListEntry{{MediaID: "m1"}}, nil
		},
	}
	uc := usecase.NewListUseCase(listRepo, nil, nil)

	entries, err := uc.GetList(context.Background(), "u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(entries) != 1 || entries[0].MediaID != "m1" {
		t.Errorf("unexpected entries returned: %+v", entries)
	}
}

func TestListUseCase_Remove(t *testing.T) {
	listRepo := &mockListRepo{
		removeFn: func(ctx context.Context, userID, mediaID string) error { return nil },
	}
	uc := usecase.NewListUseCase(listRepo, nil, nil)
	if err := uc.Remove(context.Background(), "u1", "m1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListUseCase_GetStatus(t *testing.T) {
	listRepo := &mockListRepo{
		getByUserAndExternalIDFn: func(ctx context.Context, userID, externalID, source string) (*port.ListEntry, error) {
			return &port.ListEntry{ExternalID: "550"}, nil
		},
	}
	uc := usecase.NewListUseCase(listRepo, nil, nil)
	entry, err := uc.GetStatus(context.Background(), "u1", "550", string(domain.SourceTMDB))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.ExternalID != "550" {
		t.Errorf("unexpected external ID: %s", entry.ExternalID)
	}
}
