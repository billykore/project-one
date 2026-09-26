package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// Authenticator is the driving port for authenticating an inbound request.
// Implementations validate the credential, enforce session revocation, and
// return a current, safe user principal for delivery adapters to attach to the
// request context.
type Authenticator interface {
	Authenticate(ctx context.Context, rawToken string) (*domain.User, error)
}
