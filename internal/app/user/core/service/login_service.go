package service

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
)

type loginService struct {
	repo   ports.UserRepository
	tokens ports.TokenService
	hasher ports.Hasher
	log    ports.Logger
}

// NewLoginService creates a new instance of LoginService.
func NewLoginService(
	repo ports.UserRepository,
	tokens ports.TokenService,
	hasher ports.Hasher,
	log ports.Logger,
) ports.LoginService {
	return &loginService{
		repo:   repo,
		tokens: tokens,
		hasher: hasher,
		log:    log,
	}
}

func (s *loginService) Login(ctx context.Context, email, password string) (string, string, error) {
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
	accessToken, refreshToken, err := s.tokens.GenerateTokens(ctx, user)
	if err != nil {
		s.log.Error(ctx, "failed to generate tokens", "userID", user.ID, "error", err)
		return "", "", fmt.Errorf("generate tokens: %w", err)
	}

	s.log.Info(ctx, "user logged in successfully", "userID", user.ID)
	return accessToken, refreshToken, nil
}

func (s *loginService) Logout(ctx context.Context, _ string) error {
	// For now, logout is just a placeholder as tokens are stateless.
	// In the future, we could implement token blacklisting here.
	s.log.Info(ctx, "user logged out successfully")
	return nil
}
