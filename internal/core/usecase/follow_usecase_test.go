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
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 2}, nil)
		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		follow, err := svc.Follow(ctx, followerUsername, followedUsername)

		assert.NoError(t, err)
		assert.NotNil(t, follow)
		assert.Equal(t, 1, follow.FollowerID)
		assert.Equal(t, 2, follow.FollowedID)
	})

	t.Run("cannot follow self", func(t *testing.T) {
		username := "user1"
		follow, err := svc.Follow(ctx, username, username)

		assert.ErrorIs(t, err, domain.ErrCannotFollowSelf)
		assert.Nil(t, follow)
	})

	t.Run("user not found", func(t *testing.T) {
		followerUsername := "user1"
		followedUsername := "notfound"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(nil, domain.ErrUserNotFound)

		follow, err := svc.Follow(ctx, followerUsername, followedUsername)

		assert.Error(t, err)
		assert.Nil(t, follow)
	})

	t.Run("already following", func(t *testing.T) {
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 2}, nil)
		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrAlreadyFollowing)

		follow, err := svc.Follow(ctx, followerUsername, followedUsername)

		assert.Error(t, err)
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
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 2}, nil)
		mockFollowRepo.EXPECT().Delete(ctx, 1, 2).Return(nil)

		err := svc.Unfollow(ctx, followerUsername, followedUsername)

		assert.NoError(t, err)
	})

	t.Run("cannot unfollow self", func(t *testing.T) {
		username := "user1"
		err := svc.Unfollow(ctx, username, username)

		assert.ErrorIs(t, err, domain.ErrCannotUnfollowSelf)
	})

	t.Run("not following", func(t *testing.T) {
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 2}, nil)
		mockFollowRepo.EXPECT().Delete(ctx, 1, 2).Return(domain.ErrNotFollowing)

		err := svc.Unfollow(ctx, followerUsername, followedUsername)

		assert.Error(t, err)
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
		followerUsername := "user1"
		limit := 10
		offset := 0

		expectedFollowing := []domain.Following{
			{ID: 2, FirstName: "John", LastName: "Doe", IsMutual: true},
		}

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1}, nil)
		mockFollowRepo.EXPECT().GetFollowing(ctx, 1, limit, offset).Return(expectedFollowing, nil)

		results, err := svc.GetFollowing(ctx, followerUsername, limit, offset)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowing, results)
	})

	t.Run("pagination defaults", func(t *testing.T) {
		followerUsername := "user1"
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1}, nil)
		mockFollowRepo.EXPECT().GetFollowing(ctx, 1, 10, 0).Return([]domain.Following{}, nil)

		_, err := svc.GetFollowing(ctx, followerUsername, 0, -1)

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
		followedUsername := "user1"
		limit := 10
		offset := 0

		expectedFollowers := []domain.Follower{
			{ID: 2, FirstName: "John", LastName: "Doe", IsMutual: true},
		}

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 1}, nil)
		mockFollowRepo.EXPECT().GetFollowers(ctx, 1, limit, offset).Return(expectedFollowers, nil)

		results, err := svc.GetFollowers(ctx, followedUsername, limit, offset)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowers, results)
	})

	t.Run("pagination defaults", func(t *testing.T) {
		followedUsername := "user1"
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 1}, nil)
		mockFollowRepo.EXPECT().GetFollowers(ctx, 1, 10, 0).Return([]domain.Follower{}, nil)

		_, err := svc.GetFollowers(ctx, followedUsername, 0, -1)

		assert.NoError(t, err)
	})
}
