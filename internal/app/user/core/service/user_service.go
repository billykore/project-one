package service

import (
	"context"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
)

type userService struct {
	repo ports.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo ports.UserRepository) ports.UserService {
	return &userService{repo: repo}
}

func (s *userService) GetCurrentUser(ctx context.Context, id int) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}
