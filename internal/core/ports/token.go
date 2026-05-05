package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// TokenRepository is a driven port for user token persistence.
type TokenRepository interface {
	// StoreToken saves a new user token.
	StoreToken(ctx context.Context, token *domain.UserToken) error
	// GetTokenByUserID retrieves a user token by the associated user ID.
	GetTokenByUserID(ctx context.Context, userID int) (*domain.UserToken, error)
	// DeleteToken removes a user token by its string value.
	DeleteTokenByUserID(ctx context.Context, userID int) error
}
