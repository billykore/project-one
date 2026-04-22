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
	svc := NewUserService(mockRepo)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := 1
		expectedUser := &domain.User{ID: userID, Email: "test@example.com"}

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

		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, domain.ErrUserNotFound)

		user, err := svc.GetCurrentUser(ctx, userID)

		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("GetCurrentUser() error = %v, want %v", err, domain.ErrUserNotFound)
		}
		if user != nil {
			t.Errorf("GetCurrentUser() user = %v, want nil", user)
		}
	})
}
