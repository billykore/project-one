package repository

import (
	"context"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/publishing/ports"
)

// identityUsers is the slice of identity's user repository this adapter reads.
type identityUsers interface {
	GetUserByID(ctx context.Context, id int) (*identitydomain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*identitydomain.User, error)
}

// authorLookup translates identity users into publishing's own Author model so the
// identity aggregate does not leak past this adapter.
type authorLookup struct {
	users identityUsers
}

var _ ports.UserLookup = (*authorLookup)(nil)

// NewUserLookup adapts identity's user repository to the UserLookup port.
func NewUserLookup(users identityUsers) ports.UserLookup {
	return &authorLookup{users: users}
}

// GetUserByID returns the publishing projection of the identity user.
func (l *authorLookup) GetUserByID(ctx context.Context, id int) (*domain.Author, error) {
	user, err := l.users.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Author{ID: user.ID, Username: user.Username}, nil
}

// GetUserByUsername returns the publishing projection of the identity user.
func (l *authorLookup) GetUserByUsername(ctx context.Context, username string) (*domain.Author, error) {
	user, err := l.users.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &domain.Author{ID: user.ID, Username: user.Username}, nil
}
