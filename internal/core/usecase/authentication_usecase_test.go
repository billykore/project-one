package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthenticationUseCase_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tokens := mocks.NewMockTokenService(ctrl)
	sessions := mocks.NewMockTokenRepository(ctrl)
	users := mocks.NewMockUserRepository(ctrl)
	uc := NewAuthenticationUseCase(tokens, sessions, users)
	ctx := context.Background()

	t.Run("returns a current minimal principal", func(t *testing.T) {
		tokens.EXPECT().ValidateToken(ctx, "valid").Return(&domain.User{ID: 11, Username: "old-name"}, nil)
		sessions.EXPECT().IsActive(ctx, "valid", 11).Return(true, nil)
		users.EXPECT().GetUserByID(ctx, 11).Return(&domain.User{ID: 11, Username: "new-name", Password: "never-expose"}, nil)

		user, err := uc.Authenticate(ctx, "valid")

		assert.NoError(t, err)
		assert.Equal(t, &domain.User{ID: 11, Username: "new-name"}, user)
	})

	t.Run("rejects a revoked session", func(t *testing.T) {
		tokens.EXPECT().ValidateToken(ctx, "revoked").Return(&domain.User{ID: 11}, nil)
		sessions.EXPECT().IsActive(ctx, "revoked", 11).Return(false, nil)

		user, err := uc.Authenticate(ctx, "revoked")

		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUntrustedToken)
	})

	t.Run("does not expose verifier or repository failures", func(t *testing.T) {
		tokens.EXPECT().ValidateToken(ctx, "broken").Return(nil, errors.New("invalid signature"))

		user, err := uc.Authenticate(ctx, "broken")

		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUntrustedToken)
	})
}
