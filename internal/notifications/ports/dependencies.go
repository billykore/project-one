package ports

import (
	"context"

	"github.com/billykore/project-one/internal/notifications/domain"
)

// ActorLookup is notifications' narrow identity projection dependency.
type ActorLookup interface {
	GetUserByID(ctx context.Context, id int) (*domain.Actor, error)
}
