package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// TokenRepository is a driven port for user token persistence.
type TokenRepository interface {
	// StoreToken saves a new user token.
	StoreToken(ctx context.Context, token *domain.UserToken) error
	// GetTokenByUsername retrieves a user token by the associated username.
	GetTokenByUsername(ctx context.Context, username string) (*domain.UserToken, error)
	// DeleteToken removes a user token by its string value.
	DeleteTokenByUsername(ctx context.Context, username string) error
}
