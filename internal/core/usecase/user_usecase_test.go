package usecase

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_GetUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokenRepo := mocks.NewMockTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserUseCase(mockRepo, mockTokenRepo, mockHasher)

	ctx := context.Background()
	username := "testuser"

	t.Run("success", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:        1,
			FirstName: "Test",
			LastName:  "User",
			Username:  username,
			Email:     "test@example.com",
		}
		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(expectedUser, nil)

		user, err := svc.GetUser(ctx, username)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, domain.ErrUserNotFound)

		user, err := svc.GetUser(ctx, username)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestUserUseCase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokenRepo := mocks.NewMockTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserUseCase(mockRepo, mockTokenRepo, mockHasher)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
			Email:     "john@example.com",
			Password:  "password123",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, domain.ErrUserNotFound)
		mockRepo.EXPECT().GetUserByUsername(ctx, user.Username).Return(nil, domain.ErrUserNotFound)
		mockHasher.EXPECT().Hash(ctx, user.Password).Return("hashed_password", nil)
		mockRepo.EXPECT().CreateUser(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
			assert.Equal(t, "hashed_password", u.Password)
			u.ID = 1
			return nil
		})

		err := svc.Register(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
	})

	t.Run("email already registered", func(t *testing.T) {
		user := &domain.User{
			Email: "exists@example.com",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(&domain.User{ID: 1}, nil)

		err := svc.Register(ctx, user)

		assert.ErrorIs(t, err, domain.ErrEmailAlreadyRegistered)
	})

	t.Run("username already taken", func(t *testing.T) {
		user := &domain.User{
			Username: "johndoe",
			Email:    "john@example.com",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, domain.ErrUserNotFound)
		mockRepo.EXPECT().GetUserByUsername(ctx, user.Username).Return(&domain.User{ID: 1}, nil)

		err := svc.Register(ctx, user)

		assert.ErrorIs(t, err, domain.ErrUsernameAlreadyTaken)
	})

	t.Run("validation failure", func(t *testing.T) {
		user := &domain.User{
			FirstName: "Jo", // too short
			LastName:  "Doe",
			Username:  "johndoe",
			Email:     "john@example.com",
			Password:  "password123",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, domain.ErrUserNotFound)
		mockRepo.EXPECT().GetUserByUsername(ctx, user.Username).Return(nil, domain.ErrUserNotFound)

		err := svc.Register(ctx, user)

		assert.ErrorIs(t, err, domain.ErrValidationFailed)
	})
}
