package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// TokenRepository is a driven port for user token persistence.
type TokenRepository interface {
	// StoreToken saves a new user token.
	StoreToken(ctx context.Context, token *domain.UserToken) error
	// IsActive reports whether this exact, unexpired token is an active session
	// for the authenticated user. It is used after JWT signature validation so
	// logout can revoke a token before its JWT expiry.
	IsActive(ctx context.Context, token string, userID int) (bool, error)
	// GetTokenByUsername retrieves a user token by the associated username.
	GetTokenByUsername(ctx context.Context, username string) (*domain.UserToken, error)
	// DeleteToken removes a user token by its string value.
	DeleteTokenByUsername(ctx context.Context, username string) error
}
