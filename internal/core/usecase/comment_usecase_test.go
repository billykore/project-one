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
	content := "This is a comment"

	t.Run("success", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, int(postID)).
			Return(&domain.Post{ID: int(postID)}, nil)

		mockCommentRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, comment *domain.Comment) error {
				comment.ID = 100
				return nil
			})

		mockLog.EXPECT().Info(ctx, "comment created successfully", "commentID", 100, "postID", postID, "username", username)

		err := svc.AddComment(ctx, postID, username, content)
		assert.NoError(t, err)
	})

	t.Run("validation failure - empty content", func(t *testing.T) {
		err := svc.AddComment(ctx, postID, username, "")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidationFailed))
	})

	t.Run("validation failure - whitespace content", func(t *testing.T) {
		err := svc.AddComment(ctx, postID, username, "   ")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidationFailed))
	})

	t.Run("post not found", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, int(postID)).
			Return(nil, domain.ErrPostNotFound)

		err := svc.AddComment(ctx, postID, username, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
	})

	t.Run("repository error on create", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, int(postID)).
			Return(&domain.Post{ID: int(postID)}, nil)

		mockCommentRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("db error"))

		mockLog.EXPECT().Error(ctx, "failed to create comment", "postID", postID, "username", username, "error", gomock.Any())

		err := svc.AddComment(ctx, postID, username, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})
}

func TestCommentUseCase_GetCommentsByPostID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockLog)

	ctx := context.Background()
	postID := 1

	t.Run("success", func(t *testing.T) {
		expectedComments := []*domain.Comment{
			{ID: 1, PostID: postID, Username: "commenter1", Content: "First comment"},
			{ID: 2, PostID: postID, Username: "commenter2", Content: "Second comment"},
		}
		mockCommentRepo.EXPECT().GetByPostID(ctx, postID).Return(expectedComments, nil)

		comments, err := svc.GetCommentsByPostID(ctx, postID)
		assert.NoError(t, err)
		assert.Equal(t, expectedComments, comments)
	})

	t.Run("repository error", func(t *testing.T) {
		mockCommentRepo.EXPECT().GetByPostID(ctx, postID).Return(nil, errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to get comments for post", "postID", postID, "error", gomock.Any())

		comments, err := svc.GetCommentsByPostID(ctx, postID)
		assert.Error(t, err)
		assert.Nil(t, comments)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})
}

func TestCommentUseCase_EditComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockLog)

	ctx := context.Background()
	commentID := 1
	authorUsername := "author"
	nonAuthorUsername := "hacker"
	originalContent := "original content"
	newContent := "updated content"

	t.Run("success", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockCommentRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, comment *domain.Comment) error {
				assert.Equal(t, newContent, comment.Content)
				return nil
			})

		mockLog.EXPECT().Info(ctx, "comment updated successfully", "commentID", commentID, "username", authorUsername)

		err := svc.EditComment(ctx, commentID, authorUsername, newContent)
		assert.NoError(t, err)
	})

	t.Run("comment not found", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, domain.ErrCommentNotFound)

		err := svc.EditComment(ctx, commentID, authorUsername, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentNotFound))
	})

	t.Run("comment is nil", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, nil)

		err := svc.EditComment(ctx, commentID, authorUsername, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentNotFound))
	})

	t.Run("unauthorized", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockLog.EXPECT().Warn(ctx, "unauthorized attempt to edit comment", "commentID", commentID, "attemptedBy", nonAuthorUsername, "actualAuthor", authorUsername)

		err := svc.EditComment(ctx, commentID, nonAuthorUsername, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrUnauthorized))
	})

	t.Run("validation failure - empty content", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		err := svc.EditComment(ctx, commentID, authorUsername, "")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidationFailed))
	})

	t.Run("validation failure - whitespace content", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		err := svc.EditComment(ctx, commentID, authorUsername, "   ")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidationFailed))
	})

	t.Run("repository update error", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockCommentRepo.EXPECT().
			Update(ctx, gomock.Any()).
			Return(errors.New("db update error"))

		mockLog.EXPECT().Error(ctx, "failed to update comment in repository", "commentID", commentID, "error", gomock.Any())

		err := svc.EditComment(ctx, commentID, authorUsername, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})
}
