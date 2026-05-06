package usecase

import (
	"context"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
)

type UserUseCase struct {
	userRepo port.UserRepository
}

func NewUserUseCase(userRepo port.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

func (uc *UserUseCase) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
	return uc.userRepo.SearchByUsername(ctx, query, 20)
}

func (uc *UserUseCase) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return uc.userRepo.GetByUsername(ctx, username)
}
