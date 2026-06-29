package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type userUseCase struct {
	userRepo ports.UserRepository
	hasher   ports.Hasher
}

// NewUserUseCase creates a new instance of ports.UserUseCase.
func NewUserUseCase(userRepo ports.UserRepository, hasher ports.Hasher) ports.UserUseCase {
	if userRepo == nil {
		panic("NewUserUseCase: userRepo is required")
	}
	if hasher == nil {
		panic("NewUserUseCase: hasher is required")
	}
	return &userUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (s *userUseCase) GetUser(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (s *userUseCase) Register(ctx context.Context, user *domain.User) error {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return domain.ErrEmailAlreadyRegistered
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("get user by email: %w", err)
	}

	existingUsername, err := s.userRepo.GetUserByUsername(ctx, user.Username)
	if err == nil && existingUsername != nil {
		return domain.ErrUsernameAlreadyTaken
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("get user by username: %w", err)
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

func (s *userUseCase) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	// 1. Retrieve user
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("get user by username: %w", err)
	}

	// 2. Validate current password
	if err := s.hasher.Compare(ctx, oldPassword, user.Password); err != nil {
		return domain.ErrInvalidCredentials
	}

	// 3. Validate new password length (minimum 8 characters)
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrValidationFailed)
	}

	// 4. Hash new password
	hashedPassword, err := s.hasher.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.Password = hashedPassword

	// 5. Save to repository
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}
