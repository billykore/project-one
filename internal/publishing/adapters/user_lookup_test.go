package repository

import (
	"context"
	"testing"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
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

func (f fakeIdentityUsers) GetUserByUsername(_ context.Context, username string) (*identitydomain.User, error) {
	return f.find(username)
}

func TestNewUserLookup_ProjectsOnlyIDAndUsername(t *testing.T) {
	lookup := NewUserLookup(fakeIdentityUsers{user: &identitydomain.User{ID: 42, Username: "alice", Email: "alice@example.test", Password: "hash"}})

	account, err := lookup.GetUserByID(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, &domain.Author{ID: 42, Username: "alice"}, account)

	account, err = lookup.GetUserByUsername(context.Background(), "alice")
	require.NoError(t, err)
	assert.Equal(t, &domain.Author{ID: 42, Username: "alice"}, account)
}

func TestNewUserLookup_PassesThroughNotFound(t *testing.T) {
	_, err := NewUserLookup(fakeIdentityUsers{}).GetUserByID(context.Background(), 7)
	assert.ErrorIs(t, err, problem.ErrUserNotFound)
}
