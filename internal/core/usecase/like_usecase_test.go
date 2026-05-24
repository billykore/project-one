package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLikeUseCase_ToggleLike(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)

	svc := NewLikeUseCase(mockLikeRepo, mockPostRepo, mockLog)

	ctx := context.Background()
	postID := 1
	username := "testuser"
	expectedPost := &domain.Post{ID: postID, LikeCount: 1}

	t.Run("success - like (post not yet liked)", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		mockPostRepo.EXPECT().
			IncrementLikeCount(ctx, postID, 1).
			Return(nil)

		mockLog.EXPECT().
			Info(ctx, "post liked successfully", "postID", postID, "username", username)

		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(expectedPost, nil)

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.NoError(t, err)
		assert.True(t, liked)
		assert.Equal(t, 1, likeCount)
	})

	t.Run("success - unlike (post already liked)", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(domain.ErrAlreadyLiked)

		mockLikeRepo.EXPECT().
			Delete(ctx, postID, username).
			Return(nil)

		mockPostRepo.EXPECT().
			IncrementLikeCount(ctx, postID, -1).
			Return(nil)

		mockLog.EXPECT().
			Info(ctx, "post unliked successfully", "postID", postID, "username", username)

		expectedPostUnlike := &domain.Post{ID: postID, LikeCount: 0}
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(expectedPostUnlike, nil)

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.NoError(t, err)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("post not found", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(domain.ErrPostNotFound)

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrPostNotFound)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("invalid post ID (zero)", func(t *testing.T) {
		liked, likeCount, err := svc.ToggleLike(ctx, 0, username)
		assert.ErrorIs(t, err, domain.ErrInvalidPost)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("empty username", func(t *testing.T) {
		liked, likeCount, err := svc.ToggleLike(ctx, postID, "")
		assert.ErrorIs(t, err, domain.ErrValidationFailed)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("create fails with internal error", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("db error"))

		mockLog.EXPECT().
			Error(ctx, "failed to create like", "postID", postID, "username", username, "error", gomock.Any())

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("delete fails during unlike", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(domain.ErrAlreadyLiked)

		mockLikeRepo.EXPECT().
			Delete(ctx, postID, username).
			Return(errors.New("delete error"))

		mockLog.EXPECT().
			Error(ctx, "failed to delete like", "postID", postID, "username", username, "error", gomock.Any())

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("delete returns ErrNotLiked during unlike", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(domain.ErrAlreadyLiked)

		mockLikeRepo.EXPECT().
			Delete(ctx, postID, username).
			Return(domain.ErrNotLiked)

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrNotLiked)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("increment like count fails", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		mockPostRepo.EXPECT().
			IncrementLikeCount(ctx, postID, 1).
			Return(errors.New("increment error"))

		mockLog.EXPECT().
			Error(ctx, "failed to increment like count", "postID", postID, "error", gomock.Any())

		mockLog.EXPECT().
			Info(ctx, "post liked successfully", "postID", postID, "username", username)

		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(expectedPost, nil)

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.NoError(t, err)
		assert.True(t, liked)
		assert.Equal(t, 1, likeCount)
	})

	t.Run("get post by ID fails", func(t *testing.T) {
		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		mockPostRepo.EXPECT().
			IncrementLikeCount(ctx, postID, 1).
			Return(nil)

		mockLog.EXPECT().
			Info(ctx, "post liked successfully", "postID", postID, "username", username)

		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(nil, errors.New("db error"))

		mockLog.EXPECT().
			Error(ctx, "failed to get post for like count", "postID", postID, "error", gomock.Any())

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked) // Even if it liked successfully, we return internal server error according to logic
		assert.Equal(t, 0, likeCount)
	})
}

func TestLikeUseCase_GetLikeStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)

	svc := NewLikeUseCase(mockLikeRepo, mockPostRepo, mockLog)

	ctx := context.Background()
	postID := 1
	username := "testuser"
	expectedPost := &domain.Post{ID: postID, LikeCount: 5}

	t.Run("success - user has liked", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(expectedPost, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(true, nil)

		liked, likeCount, err := svc.GetLikeStatus(ctx, postID, username)
		assert.NoError(t, err)
		assert.True(t, liked)
		assert.Equal(t, 5, likeCount)
	})

	t.Run("success - user has not liked", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(expectedPost, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(false, nil)

		liked, likeCount, err := svc.GetLikeStatus(ctx, postID, username)
		assert.NoError(t, err)
		assert.False(t, liked)
		assert.Equal(t, 5, likeCount)
	})

	t.Run("post not found", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(nil, domain.ErrPostNotFound)

		liked, likeCount, err := svc.GetLikeStatus(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrPostNotFound)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("invalid post ID", func(t *testing.T) {
		liked, likeCount, err := svc.GetLikeStatus(ctx, 0, username)
		assert.ErrorIs(t, err, domain.ErrInvalidPost)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("empty username", func(t *testing.T) {
		liked, likeCount, err := svc.GetLikeStatus(ctx, postID, "")
		assert.ErrorIs(t, err, domain.ErrValidationFailed)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("exists check fails", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(expectedPost, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(false, errors.New("exists check error"))

		mockLog.EXPECT().
			Error(ctx, "failed to check like existence", "postID", postID, "username", username, "error", gomock.Any())

		liked, likeCount, err := svc.GetLikeStatus(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})
}
