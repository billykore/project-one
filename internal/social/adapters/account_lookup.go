package repository

import (
	"context"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/social/domain"
	"github.com/billykore/project-one/internal/social/ports"
)

// identityUsers is the slice of identity's user repository this adapter reads.
type identityUsers interface {
	GetUserByID(ctx context.Context, id int) (*identitydomain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*identitydomain.User, error)
}

// accountLookup translates identity users into social's own Account model so the
// identity aggregate does not leak past this adapter.
type accountLookup struct {
	users identityUsers
}

var _ ports.AccountLookup = (*accountLookup)(nil)

// NewAccountLookup adapts identity's user repository to the AccountLookup port.
func NewAccountLookup(users identityUsers) ports.AccountLookup {
	return &accountLookup{users: users}
}

// GetUserByID returns the social projection of the identity user.
func (l *accountLookup) GetUserByID(ctx context.Context, id int) (*domain.Account, error) {
	user, err := l.users.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Account{ID: user.ID, Username: user.Username}, nil
}

// GetUserByUsername returns the social projection of the identity user.
func (l *accountLookup) GetUserByUsername(ctx context.Context, username string) (*domain.Account, error) {
	user, err := l.users.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &domain.Account{ID: user.ID, Username: user.Username}, nil
}
