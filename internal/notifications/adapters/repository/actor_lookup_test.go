package repository

import (
	"context"
	"testing"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeIdentityUsers struct {
	user *identitydomain.User
}

func (f fakeIdentityUsers) find(any) (*identitydomain.User, error) {
	if f.user == nil {
		return nil, problem.ErrUserNotFound
	}
	return f.user, nil
}

func (f fakeIdentityUsers) GetUserByID(_ context.Context, id int) (*identitydomain.User, error) {
	return f.find(id)
}

func TestNewActorLookup_ProjectsOnlyIDAndUsername(t *testing.T) {
	lookup := NewActorLookup(fakeIdentityUsers{user: &identitydomain.User{ID: 42, Username: "alice", Email: "alice@example.test", Password: "hash"}})

	account, err := lookup.GetUserByID(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, &domain.Actor{ID: 42, Username: "alice"}, account)
}

func TestNewActorLookup_PassesThroughNotFound(t *testing.T) {
	_, err := NewActorLookup(fakeIdentityUsers{}).GetUserByID(context.Background(), 7)
	assert.ErrorIs(t, err, problem.ErrUserNotFound)
}
