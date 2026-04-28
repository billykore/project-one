package service

import (
	"context"
	"errors"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
)

type userService struct {
	repo   ports.UserRepository
	hasher ports.Hasher
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo ports.UserRepository, hasher ports.Hasher) ports.UserService {
	return &userService{
		repo:   repo,
		hasher: hasher,
	}
}

func (s *userService) GetCurrentUser(ctx context.Context, id int) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *userService) Register(ctx context.Context, user *domain.User) error {
	existingUser, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return domain.ErrEmailAlreadyRegistered
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}

	if err := user.Validate(); err != nil {
		return err
	}

	hashedPassword, err := s.hasher.Hash(ctx, user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return s.repo.CreateUser(ctx, user)
}
