package usecase

import (
	"context"
	"testing"
	"time"

	vo "github.com/billykore/project-one/internal/platform/pagination"
	"github.com/billykore/project-one/internal/platform/problem"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFeedUseCase_GetFeed_ReturnsPostsForUserAndFollowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostQueryRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, logger)

	t.Run("returns posts from self and followed users", func(t *testing.T) {
		ctx := context.Background()
		userID := 1
		now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

		followRepo.EXPECT().GetFollowedUserIDs(ctx, 1).Return([]int{2, 3}, nil)

		postRepo.EXPECT().GetFeed(ctx, []int{1, 2, 3}, (*vo.Cursor)(nil), 11).
			Return([]*publishingdomain.Post{
				{ID: 3, Username: "charlie", Title: "Third", Content: "Content 3", CreatedAt: now, UpdatedAt: now},
				{ID: 2, Username: "bob", Title: "Second", Content: "Content 2", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
			}, nil)

		result, err := uc.GetFeed(ctx, userID, nil, 10)
		assert.NoError(t, err)
		assert.Len(t, result.Posts, 2)
		assert.False(t, result.HasMore)
		assert.Nil(t, result.NextCursor)
	})
}

func TestFeedUseCase_GetFeed_DetectsHasMore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostQueryRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, logger)
	ctx := context.Background()
	userID := 1
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	followRepo.EXPECT().GetFollowedUserIDs(ctx, 1).Return([]int{}, nil)

	posts := make([]*publishingdomain.Post, 11)
	for i := 0; i < 11; i++ {
		posts[i] = &publishingdomain.Post{
			ID:        i + 1,
			Username:  "alice",
			Title:     "Post",
			Content:   "Content",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			UpdatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	postRepo.EXPECT().GetFeed(ctx, []int{1}, (*vo.Cursor)(nil), 11).Return(posts, nil)

	result, err := uc.GetFeed(ctx, userID, (*vo.Cursor)(nil), 10)
	assert.NoError(t, err)
	assert.Len(t, result.Posts, 10)
	assert.True(t, result.HasMore)
	assert.NotNil(t, result.NextCursor)
}

func TestFeedUseCase_GetFeed_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostQueryRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, logger)
	ctx := context.Background()

	_, err := uc.GetFeed(ctx, 0, nil, 10)
	assert.ErrorIs(t, err, problem.ErrInvalidUser)
}

func TestFeedUseCase_GetFeed_EmptyFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostQueryRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, logger)
	ctx := context.Background()

	followRepo.EXPECT().GetFollowedUserIDs(ctx, 1).Return([]int{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []int{1}, (*vo.Cursor)(nil), 11).Return([]*publishingdomain.Post{}, nil)

	result, err := uc.GetFeed(ctx, 1, nil, 10)
	assert.NoError(t, err)
	assert.Len(t, result.Posts, 0)
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)
}

func TestFeedUseCase_GetFeed_WithCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostQueryRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, logger)
	ctx := context.Background()
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	cursor := &vo.Cursor{
		CreatedAt: now.Add(-2 * time.Hour),
		ID:        5,
	}

	followRepo.EXPECT().GetFollowedUserIDs(ctx, 1).Return([]int{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []int{1}, cursor, 11).
		Return([]*publishingdomain.Post{
			{ID: 3, Username: "alice", Title: "Older", Content: "Content", CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: now},
		}, nil)

	result, err := uc.GetFeed(ctx, 1, cursor, 10)
	assert.NoError(t, err)
	assert.Len(t, result.Posts, 1)
	assert.False(t, result.HasMore)
}

func TestFeedUseCase_GetFeed_ClampsLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := mocks.NewMockPostQueryRepository(ctrl)
	followRepo := mocks.NewMockFollowRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc := NewFeedUseCase(postRepo, followRepo, logger)
	ctx := context.Background()

	// limit=0 clamps to 10, so dbLimit=11
	followRepo.EXPECT().GetFollowedUserIDs(ctx, 1).Return([]int{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []int{1}, (*vo.Cursor)(nil), 11).Return([]*publishingdomain.Post{}, nil)

	_, err := uc.GetFeed(ctx, 1, nil, 0)
	assert.NoError(t, err)

	// limit=100 clamps to 50, so dbLimit=51
	followRepo.EXPECT().GetFollowedUserIDs(ctx, 1).Return([]int{}, nil)
	postRepo.EXPECT().GetFeed(ctx, []int{1}, (*vo.Cursor)(nil), 51).Return([]*publishingdomain.Post{}, nil)

	_, err = uc.GetFeed(ctx, 1, nil, 100)
	assert.NoError(t, err)
}
