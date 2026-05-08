package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type userService struct {
	userRepo  ports.UserRepository
	tokenRepo ports.TokenRepository
	hasher    ports.Hasher
}

// NewUserService creates a new instance of ports.UserService.
func NewUserService(userRepo ports.UserRepository, tokenRepo ports.TokenRepository, hasher ports.Hasher) ports.UserService {
	if userRepo == nil {
		panic("NewUserService: userRepo is required")
	}
	if tokenRepo == nil {
		panic("NewUserService: tokenRepo is required")
	}
	if hasher == nil {
		panic("NewUserService: hasher is required")
	}
	return &userService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		hasher:    hasher,
	}
}

func (s *userService) GetCurrentUser(ctx context.Context, id int) (*domain.User, error) {
	token, err := s.tokenRepo.GetTokenByUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get token by user id: %w", err)
	}
	if token == nil {
		return nil, domain.ErrUnauthorized
	}
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (s *userService) Register(ctx context.Context, user *domain.User) error {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return domain.ErrEmailAlreadyRegistered
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("get user by email: %w", err)
	}

	if err := user.Validate(); err != nil {
		return fmt.Errorf("validate user: %w", err)
	}

	hashedPassword, err := s.hasher.Hash(ctx, user.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.Password = hashedPassword

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
