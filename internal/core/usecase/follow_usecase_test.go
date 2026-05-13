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

		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrUserNotFound)

		_, err := svc.Follow(ctx, followerID, followedID)

		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("Follow() error = %v, want %v", err, domain.ErrUserNotFound)
		}
	})

	t.Run("already following", func(t *testing.T) {
		followerID := 1
		followedID := 2

		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrAlreadyFollowing)

		_, err := svc.Follow(ctx, followerID, followedID)

		if !errors.Is(err, domain.ErrAlreadyFollowing) {
			t.Errorf("Follow() error = %v, want %v", err, domain.ErrAlreadyFollowing)
		}
	})
}

func TestFollowUseCase_GetFollowing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFollowRepo := mocks.NewMockFollowRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followerID := 1
		limit := 10
		offset := 0

		expectedFollowing := []*domain.Following{
			{ID: 2, FirstName: "John", LastName: "Doe", IsMutual: true},
		}

		mockFollowRepo.EXPECT().GetFollowing(ctx, followerID, limit, offset).Return(expectedFollowing, nil)

		results, err := svc.GetFollowing(ctx, followerID, limit, offset)

		if err != nil {
			t.Errorf("GetFollowing() unexpected error = %v", err)
		}
		if len(results) != len(expectedFollowing) {
			t.Errorf("GetFollowing() length = %d, want %d", len(results), len(expectedFollowing))
		}
	})

	t.Run("pagination defaults", func(t *testing.T) {
		followerID := 1
		// limit <= 0 should default to 10
		// offset < 0 should default to 0

		mockFollowRepo.EXPECT().GetFollowing(ctx, followerID, 10, 0).Return([]*domain.Following{}, nil)

		_, err := svc.GetFollowing(ctx, followerID, 0, -1)

		if err != nil {
			t.Errorf("GetFollowing() unexpected error = %v", err)
		}
	})
}
