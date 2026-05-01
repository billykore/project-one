package service

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestUserService_GetCurrentUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokenRepo := mocks.NewMockUserTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserService(mockRepo, mockTokenRepo, mockHasher)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := 1
		expectedUser := &domain.User{ID: userID, Email: "test@example.com"}

		mockTokenRepo.EXPECT().GetTokenByUserID(ctx, userID).Return(&domain.UserToken{UserID: userID}, nil)
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(expectedUser, nil)
		user, err := svc.GetCurrentUser(ctx, userID)

		if err != nil {
			t.Errorf("GetCurrentUser() unexpected error = %v", err)
		}
		if user != expectedUser {
			t.Errorf("GetCurrentUser() user = %v, want %v", user, expectedUser)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		userID := 2

		mockTokenRepo.EXPECT().GetTokenByUserID(ctx, userID).Return(&domain.UserToken{UserID: userID}, nil)
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, domain.ErrUserNotFound)

		user, err := svc.GetCurrentUser(ctx, userID)

		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("GetCurrentUser() error = %v, want %v", err, domain.ErrUserNotFound)
		}
		if user != nil {
			t.Errorf("GetCurrentUser() user = %v, want nil", user)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		userID := 3

		mockTokenRepo.EXPECT().GetTokenByUserID(ctx, userID).Return(nil, nil)

		user, err := svc.GetCurrentUser(ctx, userID)

		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("GetCurrentUser() error = %v, want %v", err, domain.ErrUnauthorized)
		}
		if user != nil {
			t.Errorf("GetCurrentUser() user = %v, want nil", user)
		}
	})
}

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockTokenRepo := mocks.NewMockUserTokenRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	svc := NewUserService(mockRepo, mockTokenRepo, mockHasher)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "password123",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, domain.ErrUserNotFound)
		mockHasher.EXPECT().Hash(ctx, user.Password).Return("hashed_password", nil)
		mockRepo.EXPECT().CreateUser(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
			if u.Password != "hashed_password" {
				t.Errorf("expected hashed password, got %s", u.Password)
			}
			u.ID = 1
			return nil
		})

		err := svc.Register(ctx, user)

		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}
		if user.ID != 1 {
			t.Errorf("expected user ID 1, got %d", user.ID)
		}
	})

	t.Run("email already registered", func(t *testing.T) {
		user := &domain.User{
			Email: "exists@example.com",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(&domain.User{ID: 1}, nil)

		err := svc.Register(ctx, user)

		if !errors.Is(err, domain.ErrEmailAlreadyRegistered) {
			t.Errorf("Register() error = %v, want %v", err, domain.ErrEmailAlreadyRegistered)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		user := &domain.User{
			FirstName: "Jo", // too short
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "password123",
		}

		mockRepo.EXPECT().GetUserByEmail(ctx, user.Email).Return(nil, domain.ErrUserNotFound)

		err := svc.Register(ctx, user)

		if !errors.Is(err, domain.ErrValidationFailed) {
			t.Errorf("Register() error = %v, want %v", err, domain.ErrValidationFailed)
		}
	})
}
