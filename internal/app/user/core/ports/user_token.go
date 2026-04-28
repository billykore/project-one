package ports

import (
	"context"

	"github.com/billykore/project-one/internal/app/user/core/domain"
)

// UserTokenRepository is a driven port for user token persistence.
type UserTokenRepository interface {
	// StoreToken saves a new user token.
	StoreToken(ctx context.Context, token *domain.UserToken) error
	// DeleteToken removes a user token by its string value.
	DeleteToken(ctx context.Context, token string) error
}
