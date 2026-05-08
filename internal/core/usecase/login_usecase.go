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

func (s *loginUseCase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Error(ctx, "failed to get user by email", "email", email, "error", err)
		return "", domain.ErrInvalidCredentials
	}

	// 2. Compare passwords
	if err := s.hasher.Compare(ctx, password, user.Password); err != nil {
		s.log.Error(ctx, "password mismatch", "email", email, "error", err)
		return "", domain.ErrInvalidCredentials
	}

	// 3. Generate tokens
	accessToken, err := s.tokens.GenerateTokens(ctx, user)
	if err != nil {
		s.log.Error(ctx, "failed to generate tokens", "userID", user.ID, "error", err)
		return "", fmt.Errorf("generate tokens: %w", err)
	}

	// 4. Store access token
	err = s.tokenRepo.StoreToken(ctx, &domain.UserToken{
		UserID:    user.ID,
		Token:     accessToken.Token,
		ExpiresAt: accessToken.ExpiresAt,
	})
	if err != nil {
		s.log.Error(ctx, "failed to store user token", "userID", user.ID, "error", err)
		return "", fmt.Errorf("store user token: %w", err)
	}

	s.log.Info(ctx, "user logged in successfully", "userID", user.ID)
	return accessToken.Token, nil
}

func (s *loginUseCase) Logout(ctx context.Context, userID int) error {
	if userID == 0 {
		return fmt.Errorf("%w: userID cannot be zero", domain.ErrValidationFailed)
	}

	if err := s.tokenRepo.DeleteTokenByUserID(ctx, userID); err != nil {
		s.log.Error(ctx, "failed to delete user token on logout", "userID", userID, "error", err)
		return fmt.Errorf("delete user token: %w", err)
	}

	s.log.Info(ctx, "user logged out successfully")
	return nil
}
