package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// UserUseCase is a driving port for user-related application logic.
type UserUseCase interface {
	// GetCurrentUser retrieves the user with the given ID.
	GetCurrentUser(ctx context.Context, id int) (*domain.User, error)
	// Register creates a new user account.
	Register(ctx context.Context, user *domain.User) error
}
