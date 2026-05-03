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

func (s *loginService) Login(ctx context.Context, email, password string) (string, error) {
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
	err = s.userTokens.StoreToken(ctx, &domain.UserToken{
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

func (s *loginService) Logout(ctx context.Context, userID int) error {
	if userID == 0 {
		return fmt.Errorf("userID cannot be zero")
	}

	if err := s.userTokens.DeleteTokenByUserID(ctx, userID); err != nil {
		s.log.Error(ctx, "failed to delete user token on logout", "userID", userID, "error", err)
		return fmt.Errorf("delete user token: %w", err)
	}

	s.log.Info(ctx, "user logged out successfully")
	return nil
}
