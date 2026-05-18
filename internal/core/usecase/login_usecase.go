package usecase

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type loginUseCase struct {
	repo      ports.UserRepository
	tokens    ports.TokenService
	tokenRepo ports.TokenRepository
	hasher    ports.Hasher
	log       ports.Logger
}

// NewLoginUseCase creates a new instance of ports.LoginUseCase.
func NewLoginUseCase(
	repo ports.UserRepository,
	tokens ports.TokenService,
	tokenRepo ports.TokenRepository,
	hasher ports.Hasher,
	log ports.Logger,
) ports.LoginUseCase {
	if repo == nil {
		panic("NewLoginUseCase: repo is required")
	}
	if tokens == nil {
		panic("NewLoginUseCase: tokens is required")
	}
	if tokenRepo == nil {
		panic("NewLoginUseCase: tokenRepo is required")
	}
	if hasher == nil {
		panic("NewLoginUseCase: hasher is required")
	}
	if log == nil {
		panic("NewLoginUseCase: log is required")
	}
	return &loginUseCase{
		repo:      repo,
		tokens:    tokens,
		tokenRepo: tokenRepo,
		hasher:    hasher,
		log:       log,
	}
}

func (s *loginUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	// 1. Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Error(ctx, "failed to get user by email", "email", email, "error", err)
		return "", "", domain.ErrInvalidCredentials
	}

	// 2. Compare passwords
	if err := s.hasher.Compare(ctx, password, user.Password); err != nil {
		s.log.Error(ctx, "password mismatch", "email", email, "error", err)
		return "", "", domain.ErrInvalidCredentials
	}

	// 3. Generate tokens
	accessToken, err := s.tokens.GenerateTokens(ctx, user)
	if err != nil {
		s.log.Error(ctx, "failed to generate tokens", "username", user.Username, "error", err)
		return "", "", fmt.Errorf("generate tokens: %w", err)
	}

	// 4. Store access token
	err = s.tokenRepo.StoreToken(ctx, &domain.UserToken{
		Username:  user.Username,
		Token:     accessToken.Token,
		ExpiresAt: accessToken.ExpiresAt,
	})
	if err != nil {
		s.log.Error(ctx, "failed to store user token", "username", user.Username, "error", err)
		return "", "", fmt.Errorf("store user token: %w", err)
	}

	s.log.Info(ctx, "user logged in successfully", "username", user.Username)
	return accessToken.Token, user.Username, nil
}

func (s *loginUseCase) Logout(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", domain.ErrValidationFailed)
	}
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		s.log.Error(ctx, "failed to get user by username on logout", "username", username, "error", err)
		return fmt.Errorf("get user by username: %w", err)
	}

	if err := s.tokenRepo.DeleteTokenByUsername(ctx, user.Username); err != nil {
		s.log.Error(ctx, "failed to delete user token on logout", "username", username, "error", err)
		return fmt.Errorf("delete user token: %w", err)
	}

	s.log.Info(ctx, "user logged out successfully", "username", username)
	return nil
}
