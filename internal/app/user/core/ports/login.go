package ports

import (
	"context"

	"github.com/billykore/project-one/internal/app/user/core/domain"
)

// UserRepository is a driven port for user persistence.
type UserRepository interface {
	// GetUserByEmail retrieves a user by their email address.
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	// GetUserByID retrieves a user by their ID.
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	// CreateUser saves a new user to the repository.
	CreateUser(ctx context.Context, user *domain.User) error
}

// TokenService is a driven port for token management.
type TokenService interface {
	// GenerateTokens creates new access and refresh tokens for the given user.
	GenerateTokens(ctx context.Context, user *domain.User) (accessToken *domain.TokenDetails, err error)
	// ValidateToken verifies the token and returns the user ID.
	ValidateToken(ctx context.Context, token string) (userID int, err error)
}

// LoginService is a driving port for login-related application logic.
type LoginService interface {
	// Login authenticates a user and returns tokens.
	Login(ctx context.Context, email, password string) (accessToken string, err error)
	// Logout invalidates the given token.
	Logout(ctx context.Context, userID int) error
}
