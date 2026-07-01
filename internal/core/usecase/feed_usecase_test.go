package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/billykore/project-one/internal/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFeedUseCase_GetFeed_ReturnsPostsForUserAndFollowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, userRepo, logger)

	t.Run("returns posts from self and followed users", func(t *testing.T) {
		ctx := context.Background()
		username := "alice"
		now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

		userRepo.EXPECT().GetUserByUsername(ctx, username).Return(&domain.User{
			ID: 1, Username: "alice",
		}, nil)

		followRepo.EXPECT().GetFollowedUsernames(ctx, username).Return([]string{"bob", "charlie"}, nil)

		postRepo.EXPECT().GetFeed(ctx, []string{"alice", "bob", "charlie"}, time.Time{}, 0, 11).
			Return([]*domain.Post{
				{ID: 3, Username: "charlie", Title: "Third", Content: "Content 3", CreatedAt: now, UpdatedAt: now},
				{ID: 2, Username: "bob", Title: "Second", Content: "Content 2", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
			}, nil)

		result, err := uc.GetFeed(ctx, username, nil, 10)
		assert.NoError(t, err)
		assert.Len(t, result.Posts, 2)
		assert.False(t, result.HasMore)
		assert.NotNil(t, result.NextCursor)
		assert.Equal(t, 2, result.NextCursor.ID)
	})
}

func TestFeedUseCase_GetFeed_DetectsHasMore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, userRepo, logger)
	ctx := context.Background()
	username := "alice"
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	userRepo.EXPECT().GetUserByUsername(ctx, username).Return(&domain.User{
		ID: 1, Username: "alice",
	}, nil)

	followRepo.EXPECT().GetFollowedUsernames(ctx, username).Return([]string{}, nil)

	posts := make([]*domain.Post, 11)
	for i := 0; i < 11; i++ {
		posts[i] = &domain.Post{
			ID:        i + 1,
			Username:  "alice",
			Title:     "Post",
			Content:   "Content",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			UpdatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	postRepo.EXPECT().GetFeed(ctx, []string{"alice"}, time.Time{}, 0, 11).Return(posts, nil)

	result, err := uc.GetFeed(ctx, username, nil, 10)
	assert.NoError(t, err)
	assert.Len(t, result.Posts, 10)
	assert.True(t, result.HasMore)
	assert.NotNil(t, result.NextCursor)
}

func TestFeedUseCase_GetFeed_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, userRepo, logger)
	ctx := context.Background()

	userRepo.EXPECT().GetUserByUsername(ctx, "ghost").Return(nil, errors.New("not found"))

	_, err := uc.GetFeed(ctx, "ghost", nil, 10)
	assert.Error(t, err)
}

func TestFeedUseCase_GetFeed_EmptyFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, userRepo, logger)
	ctx := context.Background()

	userRepo.EXPECT().GetUserByUsername(ctx, "alice").Return(&domain.User{
		ID: 1, Username: "alice",
	}, nil)
	followRepo.EXPECT().GetFollowedUsernames(ctx, "alice").Return([]string{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []string{"alice"}, time.Time{}, 0, 11).Return([]*domain.Post{}, nil)

	result, err := uc.GetFeed(ctx, "alice", nil, 10)
	assert.NoError(t, err)
	assert.Len(t, result.Posts, 0)
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)
}

func TestFeedUseCase_GetFeed_WithCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, userRepo, logger)
	ctx := context.Background()
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	cursor := &pagination.Cursor{
		CreatedAt: now.Add(-2 * time.Hour),
		ID:        5,
	}

	userRepo.EXPECT().GetUserByUsername(ctx, "alice").Return(&domain.User{
		ID: 1, Username: "alice",
	}, nil)
	followRepo.EXPECT().GetFollowedUsernames(ctx, "alice").Return([]string{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []string{"alice"}, cursor.CreatedAt, cursor.ID, 11).
		Return([]*domain.Post{
			{ID: 3, Username: "alice", Title: "Older", Content: "Content", CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: now},
		}, nil)

	result, err := uc.GetFeed(ctx, "alice", cursor, 10)
	assert.NoError(t, err)
	assert.Len(t, result.Posts, 1)
	assert.False(t, result.HasMore)
}

func TestFeedUseCase_GetFeed_ClampsLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, userRepo, logger)
	ctx := context.Background()

	// limit=0 clamps to 10, so dbLimit=11
	userRepo.EXPECT().GetUserByUsername(ctx, "alice").Return(&domain.User{
		ID: 1, Username: "alice",
	}, nil)
	followRepo.EXPECT().GetFollowedUsernames(ctx, "alice").Return([]string{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []string{"alice"}, time.Time{}, 0, 11).Return([]*domain.Post{}, nil)

	_, err := uc.GetFeed(ctx, "alice", nil, 0)
	assert.NoError(t, err)

	// limit=100 clamps to 50, so dbLimit=51
	userRepo.EXPECT().GetUserByUsername(ctx, "alice").Return(&domain.User{
		ID: 1, Username: "alice",
	}, nil)
	followRepo.EXPECT().GetFollowedUsernames(ctx, "alice").Return([]string{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []string{"alice"}, time.Time{}, 0, 51).Return([]*domain.Post{}, nil)

	_, err = uc.GetFeed(ctx, "alice", nil, 100)
	assert.NoError(t, err)
}
