package repository

import (
	"context"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/billykore/project-one/internal/notifications/ports"
)

// identityUsers is the slice of identity's user repository this adapter reads.
type identityUsers interface {
	GetUserByID(ctx context.Context, id int) (*identitydomain.User, error)
}

// actorLookup translates identity users into notifications's own Actor model so the
// identity aggregate does not leak past this adapter.
type actorLookup struct {
	users identityUsers
}

var _ ports.ActorLookup = (*actorLookup)(nil)

// NewActorLookup adapts identity's user repository to the ActorLookup port.
func NewActorLookup(users identityUsers) ports.ActorLookup {
	return &actorLookup{users: users}
}

// GetUserByID returns the notifications projection of the identity user.
func (l *actorLookup) GetUserByID(ctx context.Context, id int) (*domain.Actor, error) {
	user, err := l.users.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Actor{ID: user.ID, Username: user.Username}, nil
}
