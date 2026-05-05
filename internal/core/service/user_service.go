package service

import (
	"context"
	"errors"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type userService struct {
	userRepo  ports.UserRepository
	tokenRepo ports.TokenRepository
	hasher    ports.Hasher
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo ports.UserRepository, tokenRepo ports.TokenRepository, hasher ports.Hasher) ports.UserService {
	return &userService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		hasher:    hasher,
	}
}

func (s *userService) GetCurrentUser(ctx context.Context, id int) (*domain.User, error) {
	token, err := s.tokenRepo.GetTokenByUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, domain.ErrUnauthorized
	}
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *userService) Register(ctx context.Context, user *domain.User) error {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
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

	return s.userRepo.CreateUser(ctx, user)
}
