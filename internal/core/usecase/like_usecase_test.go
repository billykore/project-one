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

	t.Run("success - like (post not yet liked)", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(false, nil)

		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		mockLog.EXPECT().
			Info(ctx, "post liked successfully", "postID", postID, "username", username)

		mockLikeRepo.EXPECT().
			CountByPostID(ctx, postID).
			Return(1, nil)

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.NoError(t, err)
		assert.True(t, liked)
		assert.Equal(t, 1, likeCount)
	})

	t.Run("success - unlike (post already liked)", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(true, nil)

		mockLikeRepo.EXPECT().
			Delete(ctx, postID, username).
			Return(nil)

		mockLog.EXPECT().
			Info(ctx, "post unliked successfully", "postID", postID, "username", username)

		mockLikeRepo.EXPECT().
			CountByPostID(ctx, postID).
			Return(0, nil)

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.NoError(t, err)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("post not found", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(nil, domain.ErrPostNotFound)

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

	t.Run("invalid post ID (negative)", func(t *testing.T) {
		liked, likeCount, err := svc.ToggleLike(ctx, -1, username)
		assert.ErrorIs(t, err, domain.ErrInvalidPost)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("post repo error (non-NotFound)", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(nil, errors.New("db error"))

		mockLog.EXPECT().
			Error(ctx, "failed to verify post existence for like", "postID", postID, "error", gomock.Any())

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("exists check fails", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(false, errors.New("exists check error"))

		mockLog.EXPECT().
			Error(ctx, "failed to check like existence", "postID", postID, "username", username, "error", gomock.Any())

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("create fails", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(false, nil)

		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("insert error"))

		mockLog.EXPECT().
			Error(ctx, "failed to create like", "postID", postID, "username", username, "error", gomock.Any())

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})

	t.Run("delete fails", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(true, nil)

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

	t.Run("CountByPostID fails after like", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(false, nil)

		mockLikeRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		mockLog.EXPECT().
			Info(ctx, "post liked successfully", "postID", postID, "username", username)

		mockLikeRepo.EXPECT().
			CountByPostID(ctx, postID).
			Return(0, errors.New("count error"))

		mockLog.EXPECT().
			Error(ctx, "failed to count likes", "postID", postID, "error", gomock.Any())

		liked, likeCount, err := svc.ToggleLike(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
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

	t.Run("success - user has liked", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(true, nil)

		mockLikeRepo.EXPECT().
			CountByPostID(ctx, postID).
			Return(5, nil)

		liked, likeCount, err := svc.GetLikeStatus(ctx, postID, username)
		assert.NoError(t, err)
		assert.True(t, liked)
		assert.Equal(t, 5, likeCount)
	})

	t.Run("success - user has not liked", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(false, nil)

		mockLikeRepo.EXPECT().
			CountByPostID(ctx, postID).
			Return(5, nil)

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

	t.Run("exists check fails", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

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

	t.Run("CountByPostID fails", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockLikeRepo.EXPECT().
			Exists(ctx, postID, username).
			Return(true, nil)

		mockLikeRepo.EXPECT().
			CountByPostID(ctx, postID).
			Return(0, errors.New("count error"))

		mockLog.EXPECT().
			Error(ctx, "failed to count likes", "postID", postID, "error", gomock.Any())

		liked, likeCount, err := svc.GetLikeStatus(ctx, postID, username)
		assert.ErrorIs(t, err, domain.ErrInternalServer)
		assert.False(t, liked)
		assert.Equal(t, 0, likeCount)
	})
}
