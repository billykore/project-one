package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// UserRepository is a driven port for user persistence.
type UserRepository interface {
	// GetUserByEmail retrieves a user by their email address.
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	// GetUserByUsername retrieves a user by their username.
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	// GetUserByID retrieves a user by their ID.
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	// CreateUser saves a new user to the repository.
	CreateUser(ctx context.Context, user *domain.User) error
	// UpdateUser updates user details, including password if changed.
	UpdateUser(ctx context.Context, user *domain.User) error
	// UpdateProfile updates the mutable profile fields (first_name, last_name, username)
	// and cascades the username change to all denormalized columns within a transaction.
	// oldUsername is the user's current username before the update.
	UpdateProfile(ctx context.Context, oldUsername string, user *domain.User) error
}

// UserUseCase is a driving port for user-related application logic.
type UserUseCase interface {
	// Register creates a new user account.
	Register(ctx context.Context, user *domain.User) error
	// GetUser retrieves a user by their username.
	GetUser(ctx context.Context, username string) (*domain.User, error)
	// ChangePassword verifies the old password and sets the new password.
	ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error
	// UpdateProfile updates the authenticated user's profile fields (first_name, last_name, username).
	UpdateProfile(ctx context.Context, username string, user *domain.User) error
}
