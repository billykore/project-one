package ports

import (
	"context"

	"github.com/billykore/project-one/internal/app/user/core/domain"
)

// UserService is a driving port for user-related application logic.
type UserService interface {
	// GetCurrentUser retrieves the user with the given ID.
	GetCurrentUser(ctx context.Context, id int) (*domain.User, error)
}
