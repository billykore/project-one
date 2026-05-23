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

func TestCommentUseCase_AddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockLog)

	ctx := context.Background()
	postID := 1
	username := "testuser"
	userID := 42
	content := "This is a comment"

	t.Run("success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, username).
			Return(&domain.User{ID: userID, Username: username}, nil)

		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockCommentRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, comment *domain.Comment) error {
				comment.ID = 100
				return nil
			})

		mockLog.EXPECT().Info(ctx, "comment created successfully", "commentID", 100, "postID", postID, "userID", userID)

		err := svc.AddComment(ctx, postID, username, content)
		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, username).
			Return(nil, domain.ErrUserNotFound)

		err := svc.AddComment(ctx, postID, username, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrUserNotFound))
	})

	t.Run("validation failure - empty content", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, username).
			Return(&domain.User{ID: userID, Username: username}, nil)

		err := svc.AddComment(ctx, postID, username, "")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidationFailed))
	})

	t.Run("post not found", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, username).
			Return(&domain.User{ID: userID, Username: username}, nil)

		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(nil, domain.ErrPostNotFound)

		err := svc.AddComment(ctx, postID, username, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
	})

	t.Run("repository error on create", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, username).
			Return(&domain.User{ID: userID, Username: username}, nil)

		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, postID).
			Return(&domain.Post{ID: postID}, nil)

		mockCommentRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("db error"))

		mockLog.EXPECT().Error(ctx, "failed to create comment", "postID", postID, "userID", userID, "error", gomock.Any())

		err := svc.AddComment(ctx, postID, username, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})
}
