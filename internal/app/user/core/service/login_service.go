package service

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
)

type loginService struct {
	repo       ports.UserRepository
	tokens     ports.TokenService
	userTokens ports.UserTokenRepository
	hasher     ports.Hasher
	log        ports.Logger
}

// NewLoginService creates a new instance of LoginService.
func NewLoginService(
	repo ports.UserRepository,
	tokens ports.TokenService,
	userTokens ports.UserTokenRepository,
	hasher ports.Hasher,
	log ports.Logger,
) ports.LoginService {
	return &loginService{
		repo:       repo,
		tokens:     tokens,
		userTokens: userTokens,
		hasher:     hasher,
		log:        log,
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

	// 4. Store access token
	accessExp, err := s.tokens.GetTokenExpiry(ctx, accessToken)
	if err != nil {
		s.log.Error(ctx, "failed to get access token expiry", "userID", user.ID, "error", err)
		return "", "", fmt.Errorf("get access token expiry: %w", err)
	}

	err = s.userTokens.StoreToken(ctx, &domain.UserToken{
		UserID:    user.ID,
		Token:     accessToken,
		ExpiresAt: accessExp,
	})
	if err != nil {
		s.log.Error(ctx, "failed to store user token", "userID", user.ID, "error", err)
		return "", "", fmt.Errorf("store user token: %w", err)
	}

	s.log.Info(ctx, "user logged in successfully", "userID", user.ID)
	return accessToken, refreshToken, nil
}

func (s *loginService) Logout(ctx context.Context, token string) error {
	if err := s.userTokens.DeleteToken(ctx, token); err != nil {
		s.log.Error(ctx, "failed to delete user token on logout", "error", err)
		return fmt.Errorf("delete user token: %w", err)
	}

	s.log.Info(ctx, "user logged out successfully")
	return nil
}
