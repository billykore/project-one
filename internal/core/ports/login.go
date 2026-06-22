package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// TokenService is a driven port for token management.
type TokenService interface {
	// GenerateTokens creates new access and refresh tokens for the given user.
	GenerateTokens(ctx context.Context, user *domain.User) (accessToken *domain.UserToken, err error)
	// ValidateToken verifies the token and returns the username.
	ValidateToken(ctx context.Context, token string) (username string, err error)
}

// LoginUseCase is a driving port for login-related application logic.
type LoginUseCase interface {
	// Login authenticates a user and returns tokens.
	Login(ctx context.Context, email, password string) (*domain.UserToken, error)
	// Logout invalidates the given token.
	Logout(ctx context.Context, username string) error
}
