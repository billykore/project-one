package usecase

import (
	"context"

	"github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/identity/ports"
	"github.com/billykore/project-one/internal/platform/problem"
)

type authenticationUseCase struct {
	tokens   ports.TokenService
	sessions ports.TokenRepository
	users    ports.UserRepository
}

// NewAuthenticationUseCase creates the application service responsible for
// translating an inbound credential into a current authenticated principal.
func NewAuthenticationUseCase(tokens ports.TokenService, sessions ports.TokenRepository, users ports.UserRepository) ports.Authenticator {
	if tokens == nil || sessions == nil || users == nil {
		panic("NewAuthenticationUseCase: dependencies must not be nil")
	}
	return &authenticationUseCase{tokens: tokens, sessions: sessions, users: users}
}

func (uc *authenticationUseCase) Authenticate(ctx context.Context, rawToken string) (*domain.User, error) {
	claimsUser, err := uc.tokens.ValidateToken(ctx, rawToken)
	if err != nil || claimsUser == nil || claimsUser.ID <= 0 {
		return nil, problem.ErrUntrustedToken
	}

	active, err := uc.sessions.IsActive(ctx, rawToken, claimsUser.ID)
	if err != nil || !active {
		return nil, problem.ErrUntrustedToken
	}

	currentUser, err := uc.users.GetUserByID(ctx, claimsUser.ID)
	if err != nil || currentUser == nil {
		return nil, problem.ErrUntrustedToken
	}

	// Do not place credential or profile data loaded from persistence in the
	// HTTP context. Downstream handlers only need the authenticated identity.
	return &domain.User{ID: currentUser.ID, Username: currentUser.Username}, nil
}
