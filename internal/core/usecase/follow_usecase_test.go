package usecase

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		assert.NoError(t, err)
		assert.NotNil(t, follow)
		assert.Equal(t, followerID, follow.FollowerID)
		assert.Equal(t, followedID, follow.FollowedID)
	})

	t.Run("cannot follow self", func(t *testing.T) {
		userID := 1
		follow, err := svc.Follow(ctx, userID, userID)

		assert.ErrorIs(t, err, domain.ErrCannotFollowSelf)
		assert.Nil(t, follow)
	})

	t.Run("user not found", func(t *testing.T) {
		followerID := 1
		followedID := 99

		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrUserNotFound)

		follow, err := svc.Follow(ctx, followerID, followedID)

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, follow)
	})

	t.Run("already following", func(t *testing.T) {
		followerID := 1
		followedID := 2

		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrAlreadyFollowing)

		follow, err := svc.Follow(ctx, followerID, followedID)

		assert.ErrorIs(t, err, domain.ErrAlreadyFollowing)
		assert.Nil(t, follow)
	})
}

func TestFollowUseCase_Unfollow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFollowRepo := mocks.NewMockFollowRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followerID := 1
		followedID := 2

		mockFollowRepo.EXPECT().Delete(ctx, followerID, followedID).Return(nil)

		err := svc.Unfollow(ctx, followerID, followedID)

		assert.NoError(t, err)
	})

	t.Run("cannot unfollow self", func(t *testing.T) {
		userID := 1
		err := svc.Unfollow(ctx, userID, userID)

		assert.ErrorIs(t, err, domain.ErrCannotUnfollowSelf)
	})

	t.Run("not following", func(t *testing.T) {
		followerID := 1
		followedID := 2

		mockFollowRepo.EXPECT().Delete(ctx, followerID, followedID).Return(domain.ErrNotFollowing)

		err := svc.Unfollow(ctx, followerID, followedID)

		assert.ErrorIs(t, err, domain.ErrNotFollowing)
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

		expectedFollowing := []domain.Following{
			{ID: 2, FirstName: "John", LastName: "Doe", IsMutual: true},
		}

		mockFollowRepo.EXPECT().GetFollowing(ctx, followerID, limit, offset).Return(expectedFollowing, nil)

		results, err := svc.GetFollowing(ctx, followerID, limit, offset)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowing, results)
	})

	t.Run("pagination defaults", func(t *testing.T) {
		followerID := 1
		// limit <= 0 should default to 10
		// offset < 0 should default to 0

		mockFollowRepo.EXPECT().GetFollowing(ctx, followerID, 10, 0).Return([]domain.Following{}, nil)

		_, err := svc.GetFollowing(ctx, followerID, 0, -1)

		assert.NoError(t, err)
	})
}

func TestFollowUseCase_GetFollowers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFollowRepo := mocks.NewMockFollowRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followedID := 1
		limit := 10
		offset := 0

		expectedFollowers := []domain.Follower{
			{ID: 2, FirstName: "John", LastName: "Doe", IsMutual: true},
		}

		mockFollowRepo.EXPECT().GetFollowers(ctx, followedID, limit, offset).Return(expectedFollowers, nil)

		results, err := svc.GetFollowers(ctx, followedID, limit, offset)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowers, results)
	})

	t.Run("pagination defaults", func(t *testing.T) {
		followedID := 1
		// limit <= 0 should default to 10
		// offset < 0 should default to 0

		mockFollowRepo.EXPECT().GetFollowers(ctx, followedID, 10, 0).Return([]domain.Follower{}, nil)

		_, err := svc.GetFollowers(ctx, followedID, 0, -1)

		assert.NoError(t, err)
	})
}
