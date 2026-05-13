package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"go.uber.org/mock/gomock"
)

func TestFollowUseCase_Follow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFollowRepo := mocks.NewMockFollowRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followerID := 1
		followedID := 2

		mockUserRepo.EXPECT().GetUserByID(ctx, followedID).Return(&domain.User{ID: followedID}, nil)
		mockFollowRepo.EXPECT().IsFollowing(ctx, followerID, followedID).Return(false, nil)
		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		follow, err := svc.Follow(ctx, followerID, followedID)

		if err != nil {
			t.Errorf("Follow() unexpected error = %v", err)
		}
		if follow.FollowerID != followerID || follow.FollowedID != followedID {
			t.Errorf("Follow() follow = %v, want follower %d, followed %d", follow, followerID, followedID)
		}
	})

	t.Run("cannot follow self", func(t *testing.T) {
		userID := 1
		_, err := svc.Follow(ctx, userID, userID)

		if !errors.Is(err, domain.ErrCannotFollowSelf) {
			t.Errorf("Follow() error = %v, want %v", err, domain.ErrCannotFollowSelf)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		followerID := 1
		followedID := 99

		mockUserRepo.EXPECT().GetUserByID(ctx, followedID).Return(nil, domain.ErrUserNotFound)

		_, err := svc.Follow(ctx, followerID, followedID)

		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("Follow() error = %v, want %v", err, domain.ErrUserNotFound)
		}
	})

	t.Run("already following", func(t *testing.T) {
		followerID := 1
		followedID := 2

		mockUserRepo.EXPECT().GetUserByID(ctx, followedID).Return(&domain.User{ID: followedID}, nil)
		mockFollowRepo.EXPECT().IsFollowing(ctx, followerID, followedID).Return(true, nil)

		_, err := svc.Follow(ctx, followerID, followedID)

		if !errors.Is(err, domain.ErrAlreadyFollowing) {
			t.Errorf("Follow() error = %v, want %v", err, domain.ErrAlreadyFollowing)
		}
	})
}
