package usecase

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_GetUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher)

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
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher)

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

		assert.ErrorIs(t, err, domain.ErrInvalidUser)
	})
}

func TestUserUseCase_ChangePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher)

	ctx := context.Background()
	username := "testuser"

	t.Run("success", func(t *testing.T) {
		existingUser := &domain.User{
			ID:       1,
			Username: username,
			Password: "hashed_old_password",
		}
		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(existingUser, nil)
		mockHasher.EXPECT().Compare(ctx, "old_password", "hashed_old_password").Return(nil)
		mockHasher.EXPECT().Hash(ctx, "new_password_123").Return("hashed_new_password", nil)
		mockRepo.EXPECT().UpdateUser(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
			assert.Equal(t, "hashed_new_password", u.Password)
			return nil
		})

		err := svc.ChangePassword(ctx, username, "old_password", "new_password_123")
		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, domain.ErrUserNotFound)

		err := svc.ChangePassword(ctx, username, "old_password", "new_password_123")
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		existingUser := &domain.User{
			ID:       1,
			Username: username,
			Password: "hashed_old_password",
		}
		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(existingUser, nil)
		mockHasher.EXPECT().Compare(ctx, "wrong_old_password", "hashed_old_password").Return(domain.ErrInvalidCredentials)

		err := svc.ChangePassword(ctx, username, "wrong_old_password", "new_password_123")
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("validation failed - too short", func(t *testing.T) {
		existingUser := &domain.User{
			ID:       1,
			Username: username,
			Password: "hashed_old_password",
		}
		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(existingUser, nil)
		mockHasher.EXPECT().Compare(ctx, "old_password", "hashed_old_password").Return(nil)

		err := svc.ChangePassword(ctx, username, "old_password", "short")
		assert.ErrorIs(t, err, domain.ErrPasswordTooShort)
	})
}

func TestUserUseCase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher)

	ctx := context.Background()
	oldUsername := "olduser"

	t.Run("success - username unchanged", func(t *testing.T) {
		currentUser := &domain.User{
			ID:        1,
			Username:  oldUsername,
			FirstName: "Old",
			LastName:  "User",
		}
		updatedUser := &domain.User{
			FirstName: "New",
			LastName:  "User",
			Username:  oldUsername,
		}

		mockRepo.EXPECT().GetUserByUsername(ctx, oldUsername).Return(currentUser, nil)
		mockRepo.EXPECT().UpdateProfile(ctx, oldUsername, gomock.Any()).DoAndReturn(
			func(_ context.Context, old string, u *domain.User) error {
				assert.Equal(t, "New", u.FirstName)
				assert.Equal(t, "User", u.LastName)
				assert.Equal(t, oldUsername, u.Username)
				return nil
			},
		)

		err := svc.UpdateProfile(ctx, oldUsername, updatedUser)
		assert.NoError(t, err)
	})

	t.Run("success - username changed", func(t *testing.T) {
		currentUser := &domain.User{
			ID:        1,
			Username:  oldUsername,
			FirstName: "Old",
			LastName:  "User",
		}
		newUsername := "newuser"
		updatedUser := &domain.User{
			FirstName: "New",
			LastName:  "User",
			Username:  newUsername,
		}

		mockRepo.EXPECT().GetUserByUsername(ctx, oldUsername).Return(currentUser, nil)
		mockRepo.EXPECT().GetUserByUsername(ctx, newUsername).Return(nil, domain.ErrUserNotFound)
		// The oldUsername parameter should be the original username (before update),
		// while the domain.User should carry the new username.
		mockRepo.EXPECT().UpdateProfile(ctx, oldUsername, gomock.Any()).DoAndReturn(
			func(_ context.Context, old string, u *domain.User) error {
				assert.Equal(t, newUsername, u.Username)
				assert.Equal(t, oldUsername, old)
				return nil
			},
		)

		err := svc.UpdateProfile(ctx, oldUsername, updatedUser)
		assert.NoError(t, err)
	})

	t.Run("username already taken", func(t *testing.T) {
		currentUser := &domain.User{
			ID:        1,
			Username:  oldUsername,
			FirstName: "Old",
			LastName:  "User",
		}
		newUsername := "takenuser"
		updatedUser := &domain.User{
			FirstName: "New",
			LastName:  "User",
			Username:  newUsername,
		}

		mockRepo.EXPECT().GetUserByUsername(ctx, oldUsername).Return(currentUser, nil)
		mockRepo.EXPECT().GetUserByUsername(ctx, newUsername).Return(&domain.User{ID: 2, Username: newUsername}, nil)

		err := svc.UpdateProfile(ctx, oldUsername, updatedUser)
		assert.ErrorIs(t, err, domain.ErrUsernameAlreadyTaken)
	})

	t.Run("validation failure - empty first name", func(t *testing.T) {
		currentUser := &domain.User{
			ID:        1,
			Username:  oldUsername,
			FirstName: "Old",
			LastName:  "User",
		}
		updatedUser := &domain.User{
			FirstName: "",
			LastName:  "User",
			Username:  oldUsername,
		}

		mockRepo.EXPECT().GetUserByUsername(ctx, oldUsername).Return(currentUser, nil)

		err := svc.UpdateProfile(ctx, oldUsername, updatedUser)
		assert.ErrorIs(t, err, domain.ErrInvalidUser)
	})

	t.Run("user not found", func(t *testing.T) {
		updatedUser := &domain.User{
			FirstName: "New",
			LastName:  "User",
			Username:  "newuser",
		}

		mockRepo.EXPECT().GetUserByUsername(ctx, "nonexistent").Return(nil, domain.ErrUserNotFound)

		err := svc.UpdateProfile(ctx, "nonexistent", updatedUser)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}
