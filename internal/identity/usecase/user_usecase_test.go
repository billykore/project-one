package usecase

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/identity/domain"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_GetUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockSearchRepo := mocks.NewMockUserSearchRepository(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher, mockSearchRepo)

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
		mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, problem.ErrUserNotFound)

		user, err := svc.GetUser(ctx, username)
		assert.ErrorIs(t, err, problem.ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestUserUseCase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockSearchRepo := mocks.NewMockUserSearchRepository(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher, mockSearchRepo)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
			Email:     "john@example.com",
			Password:  "password123",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, problem.ErrUserNotFound)
		mockRepo.EXPECT().GetUserByUsername(ctx, user.Username).Return(nil, problem.ErrUserNotFound)
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

		assert.ErrorIs(t, err, problem.ErrEmailAlreadyRegistered)
	})

	t.Run("username already taken", func(t *testing.T) {
		user := &domain.User{
			Username: "johndoe",
			Email:    "john@example.com",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, problem.ErrUserNotFound)
		mockRepo.EXPECT().GetUserByUsername(ctx, user.Username).Return(&domain.User{ID: 1}, nil)

		err := svc.Register(ctx, user)

		assert.ErrorIs(t, err, problem.ErrUsernameAlreadyTaken)
	})

	t.Run("validation failure", func(t *testing.T) {
		user := &domain.User{
			FirstName: "Jo", // too short
			LastName:  "Doe",
			Username:  "johndoe",
			Email:     "john@example.com",
			Password:  "password123",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, problem.ErrUserNotFound)
		mockRepo.EXPECT().GetUserByUsername(ctx, user.Username).Return(nil, problem.ErrUserNotFound)

		err := svc.Register(ctx, user)

		assert.ErrorIs(t, err, problem.ErrInvalidUser)
	})
}

func TestUserUseCase_ChangePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockSearchRepo := mocks.NewMockUserSearchRepository(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher, mockSearchRepo)

	ctx := context.Background()
	username := "testuser"
	userID := 1

	t.Run("success", func(t *testing.T) {
		existingUser := &domain.User{
			ID:       1,
			Username: username,
			Password: "hashed_old_password",
		}
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(existingUser, nil)
		mockHasher.EXPECT().Compare(ctx, "old_password", "hashed_old_password").Return(nil)
		mockHasher.EXPECT().Hash(ctx, "new_password_123").Return("hashed_new_password", nil)
		mockRepo.EXPECT().UpdateUser(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
			assert.Equal(t, "hashed_new_password", u.Password)
			return nil
		})

		err := svc.ChangePassword(ctx, userID, "old_password", "new_password_123")
		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, problem.ErrUserNotFound)

		err := svc.ChangePassword(ctx, userID, "old_password", "new_password_123")
		assert.ErrorIs(t, err, problem.ErrUserNotFound)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		existingUser := &domain.User{
			ID:       1,
			Username: username,
			Password: "hashed_old_password",
		}
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(existingUser, nil)
		mockHasher.EXPECT().Compare(ctx, "wrong_old_password", "hashed_old_password").Return(problem.ErrInvalidCredentials)

		err := svc.ChangePassword(ctx, userID, "wrong_old_password", "new_password_123")
		assert.ErrorIs(t, err, problem.ErrInvalidCredentials)
	})

	t.Run("validation failed - too short", func(t *testing.T) {
		existingUser := &domain.User{
			ID:       1,
			Username: username,
			Password: "hashed_old_password",
		}
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(existingUser, nil)
		mockHasher.EXPECT().Compare(ctx, "old_password", "hashed_old_password").Return(nil)

		err := svc.ChangePassword(ctx, userID, "old_password", "short")
		assert.ErrorIs(t, err, problem.ErrPasswordTooShort)
	})

	t.Run("invalid user id", func(t *testing.T) {
		err := svc.ChangePassword(ctx, 0, "old_password", "new_password_123")
		assert.ErrorIs(t, err, problem.ErrInvalidUser)
	})
}

func TestUserUseCase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockSearchRepo := mocks.NewMockUserSearchRepository(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher, mockSearchRepo)

	ctx := context.Background()
	userID := 1
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

		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(currentUser, nil)
		mockRepo.EXPECT().UpdateProfile(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, u *domain.User) error {
				assert.Equal(t, "New", u.FirstName)
				assert.Equal(t, "User", u.LastName)
				assert.Equal(t, oldUsername, u.Username)
				return nil
			},
		)

		err := svc.UpdateProfile(ctx, userID, updatedUser)
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

		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(currentUser, nil)
		mockRepo.EXPECT().GetUserByUsername(ctx, newUsername).Return(nil, problem.ErrUserNotFound)
		mockRepo.EXPECT().UpdateProfile(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, u *domain.User) error {
				assert.Equal(t, newUsername, u.Username)
				return nil
			},
		)

		err := svc.UpdateProfile(ctx, userID, updatedUser)
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

		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(currentUser, nil)
		mockRepo.EXPECT().GetUserByUsername(ctx, newUsername).Return(&domain.User{ID: 2, Username: newUsername}, nil)

		err := svc.UpdateProfile(ctx, userID, updatedUser)
		assert.ErrorIs(t, err, problem.ErrUsernameAlreadyTaken)
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

		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(currentUser, nil)

		err := svc.UpdateProfile(ctx, userID, updatedUser)
		assert.ErrorIs(t, err, problem.ErrInvalidUser)
	})

	t.Run("user not found", func(t *testing.T) {
		updatedUser := &domain.User{
			FirstName: "New",
			LastName:  "User",
			Username:  "newuser",
		}

		mockRepo.EXPECT().GetUserByID(ctx, 2).Return(nil, problem.ErrUserNotFound)

		err := svc.UpdateProfile(ctx, 2, updatedUser)
		assert.ErrorIs(t, err, problem.ErrUserNotFound)
	})
}

func TestUserUseCase_SearchUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockSearchRepo := mocks.NewMockUserSearchRepository(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher, mockSearchRepo)

	ctx := context.Background()

	t.Run("success - returns search results", func(t *testing.T) {
		expected := []domain.SearchResult{
			{Username: "billy", FirstName: "Billy", LastName: "Kore"},
			{Username: "billie", FirstName: "Billie", LastName: "Eilish"},
		}
		mockSearchRepo.EXPECT().Search(ctx, "bil", nil, 10).Return(expected, nil, false, nil)

		results, nextCursor, hasMore, err := svc.SearchUsers(ctx, "bil", nil, 10)
		assert.NoError(t, err)
		assert.Equal(t, expected, results)
		assert.Nil(t, nextCursor)
		assert.False(t, hasMore)
	})

	t.Run("success - empty results", func(t *testing.T) {
		mockSearchRepo.EXPECT().Search(ctx, "xyz", nil, 10).Return([]domain.SearchResult{}, nil, false, nil)

		results, _, _, err := svc.SearchUsers(ctx, "xyz", nil, 10)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("success - with cursor and has_more", func(t *testing.T) {
		expected := []domain.SearchResult{
			{Username: "bilbo", FirstName: "Bilbo", LastName: "Baggins"},
		}
		cursor := &vo.Cursor{ID: 10}
		nextCursor := &vo.Cursor{ID: 15}
		mockSearchRepo.EXPECT().Search(ctx, "bil", cursor, 5).Return(expected, nextCursor, true, nil)

		results, nc, hm, err := svc.SearchUsers(ctx, "bil", cursor, 5)
		assert.NoError(t, err)
		assert.Equal(t, expected, results)
		assert.Equal(t, nextCursor, nc)
		assert.True(t, hm)
	})

	t.Run("query too short - 2 chars", func(t *testing.T) {
		results, _, _, err := svc.SearchUsers(ctx, "ab", nil, 10)
		assert.ErrorIs(t, err, problem.ErrSearchQueryTooShort)
		assert.Nil(t, results)
	})

	t.Run("query too short - empty", func(t *testing.T) {
		results, _, _, err := svc.SearchUsers(ctx, "", nil, 10)
		assert.ErrorIs(t, err, problem.ErrSearchQueryTooShort)
		assert.Nil(t, results)
	})

	t.Run("repository error propagated", func(t *testing.T) {
		mockSearchRepo.EXPECT().Search(ctx, "bil", nil, 10).Return(nil, nil, false, problem.ErrRepositoryFailure)

		results, _, _, err := svc.SearchUsers(ctx, "bil", nil, 10)
		assert.ErrorIs(t, err, problem.ErrRepositoryFailure)
		assert.Nil(t, results)
	})
}
