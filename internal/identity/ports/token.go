package ports

import (
	"context"

	"github.com/billykore/project-one/internal/identity/domain"
)

// TokenRepository is a driven port for user token persistence.
type TokenRepository interface {
	// StoreToken saves a new user token.
	StoreToken(ctx context.Context, token *domain.UserToken) error
	// IsActive reports whether this exact, unexpired token is an active session
	// for the authenticated user. It is used after JWT signature validation so
	// logout can revoke a token before its JWT expiry.
	IsActive(ctx context.Context, token string, userID int) (bool, error)
	// DeleteTokensByUserID revokes every active session for a user.
	DeleteTokensByUserID(ctx context.Context, userID int) error
}
