package usecase

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFollowUseCase_Follow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFollowRepo := mocks.NewMockFollowRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo, mockPublisher, mockLogger)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1, Username: "user1"}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 2, Username: "user2"}, nil)
		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockPublisher.EXPECT().Publish(ctx, gomock.Any()).Return(nil)

		follow, err := svc.Follow(ctx, followerUsername, followedUsername)

		assert.NoError(t, err)
		assert.NotNil(t, follow)
		assert.Equal(t, "user1", follow.FollowerUsername)
		assert.Equal(t, "user2", follow.FollowedUsername)
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

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{Username: "user1"}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(nil, domain.ErrUserNotFound)

		follow, err := svc.Follow(ctx, followerUsername, followedUsername)

		assert.Error(t, err)
		assert.Nil(t, follow)
	})

	t.Run("already following", func(t *testing.T) {
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{Username: "user1"}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{Username: "user2"}, nil)
		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrAlreadyFollowing)

		follow, err := svc.Follow(ctx, followerUsername, followedUsername)

		assert.Error(t, err)
		assert.Nil(t, follow)
	})

	t.Run("validation failure does not publish", func(t *testing.T) {
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1, Username: "user1"}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 0, Username: "user2"}, nil)
		mockFollowRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockPublisher.EXPECT().Publish(ctx, gomock.Any()).Return(nil)

		follow, err := svc.Follow(ctx, followerUsername, followedUsername)

		assert.NoError(t, err)
		assert.NotNil(t, follow)
	})
}

func TestFollowUseCase_Unfollow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFollowRepo := mocks.NewMockFollowRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo, mockPublisher, mockLogger)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followerUsername := "user1"
		followedUsername := "user2"

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1, Username: followerUsername}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 2, Username: followedUsername}, nil)
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

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1, Username: followerUsername}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 2, Username: followedUsername}, nil)
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
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo, mockPublisher, mockLogger)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followerUsername := "user1"
		limit := 10

		expectedFollowing := []domain.Following{
			{Username: "user2", FirstName: "John", LastName: "Doe", IsMutual: true},
		}

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1, Username: "user1"}, nil)
		mockFollowRepo.EXPECT().GetFollowing(ctx, 1, nil, limit+1).Return(expectedFollowing, nil)

		results, err := svc.GetFollowing(ctx, followerUsername, nil, limit)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowing, results.Data)
	})

	t.Run("pagination defaults", func(t *testing.T) {
		followerUsername := "user1"
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(&domain.User{ID: 1, Username: "user1"}, nil)
		mockFollowRepo.EXPECT().GetFollowing(ctx, 1, nil, 11).Return([]domain.Following{}, nil)

		_, err := svc.GetFollowing(ctx, followerUsername, nil, 0)

		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		followerUsername := "user1"
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followerUsername).Return(nil, nil)

		results, err := svc.GetFollowing(ctx, followerUsername, nil, 10)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, results)
	})
}

func TestFollowUseCase_GetFollowers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFollowRepo := mocks.NewMockFollowRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	svc := NewFollowUseCase(mockFollowRepo, mockUserRepo, mockPublisher, mockLogger)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		followedUsername := "user1"
		limit := 10

		expectedFollowers := []domain.Follower{
			{Username: "user1", FirstName: "John", LastName: "Doe", IsMutual: true},
		}

		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 1, Username: "user1"}, nil)
		mockFollowRepo.EXPECT().GetFollowers(ctx, 1, nil, limit+1).Return(expectedFollowers, nil)

		results, err := svc.GetFollowers(ctx, followedUsername, nil, limit)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowers, results.Data)
	})

	t.Run("pagination defaults", func(t *testing.T) {
		followedUsername := "user1"
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(&domain.User{ID: 1, Username: "user1"}, nil)
		mockFollowRepo.EXPECT().GetFollowers(ctx, 1, nil, 11).Return([]domain.Follower{}, nil)

		_, err := svc.GetFollowers(ctx, followedUsername, nil, 0)

		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		followedUsername := "user1"
		mockUserRepo.EXPECT().GetUserByUsername(ctx, followedUsername).Return(nil, nil)

		results, err := svc.GetFollowers(ctx, followedUsername, nil, 10)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, results)
	})
}
