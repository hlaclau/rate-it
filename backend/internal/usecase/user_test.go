package usecase_test

import (
	"context"
	"testing"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
)

func TestUserUseCase_SearchUsers_ReturnsResults(t *testing.T) {
	repo := &mockUserRepository{
		searchByUsernameFn: func(_ context.Context, query string, limit int) ([]*domain.User, error) {
			if query != "ali" {
				t.Errorf("unexpected query: %q", query)
			}
			if limit != 20 {
				t.Errorf("unexpected limit: %d", limit)
			}
			return []*domain.User{{ID: "u1", Username: "alice"}}, nil
		},
	}

	uc := usecase.NewUserUseCase(repo)
	users, err := uc.SearchUsers(context.Background(), "ali")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 || users[0].Username != "alice" {
		t.Errorf("unexpected users: %+v", users)
	}
}

func TestUserUseCase_GetByUsername_Found(t *testing.T) {
	repo := &mockUserRepository{
		getByUsernameFn: func(_ context.Context, username string) (*domain.User, error) {
			if username != "alice" {
				t.Errorf("unexpected username: %q", username)
			}
			return &domain.User{ID: "u1", Username: "alice"}, nil
		},
	}

	uc := usecase.NewUserUseCase(repo)
	user, err := uc.GetByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("unexpected user ID: %s", user.ID)
	}
}

func TestUserUseCase_GetByUsername_NotFound(t *testing.T) {
	repo := &mockUserRepository{
		getByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, port.ErrUserNotFound
		},
	}

	uc := usecase.NewUserUseCase(repo)
	_, err := uc.GetByUsername(context.Background(), "ghost")
	if err != port.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
