package ports

import (
	"context"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
)

// ActorLookup is notifications' narrow identity projection dependency.
type ActorLookup interface {
	GetUserByID(ctx context.Context, id int) (*identitydomain.User, error)
}
