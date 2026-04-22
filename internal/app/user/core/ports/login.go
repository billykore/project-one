package ports

import (
	"context"

	"github.com/billykore/project-one/internal/app/user/core/domain"
)

// UserRepository is a driven port for user persistence.
type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
}

// TokenService is a driven port for token management.
type TokenService interface {
	GenerateTokens(ctx context.Context, user *domain.User) (accessToken, refreshToken string, err error)
	ValidateToken(ctx context.Context, token string) (userID int, err error)
}

// LoginService is a driving port for login-related application logic.
type LoginService interface {
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, token string) error
}
