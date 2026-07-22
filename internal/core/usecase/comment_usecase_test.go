package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCommentUseCase_AddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockPublisher)

	ctx := context.Background()
	postID := 1
	username := "testuser"
	content := "This is a comment"

	t.Run("success", func(t *testing.T) {
		mockPostRepo.EXPECT().
			GetByIDOnly(ctx, int(postID)).
			Return(&domain.Post{ID: int(postID), Username: "postowner"}, nil)

		mockCommentRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, comment *domain.Comment) error {
				comment.ID = 100
				return nil
			})

		mockUserRepo.EXPECT().GetUserByUsername(ctx, "postowner").Return(&domain.User{ID: 2, Username: "postowner"}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(&domain.User{ID: 1, Username: username}, nil)
		mockPublisher.EXPECT().Publish(ctx, gomock.Any()).Return(nil)

		err := svc.AddComment(ctx, postID, username, content)
		assert.NoError(t, err)
	})

	t.Run("validation failure - empty content", func(t *testing.T) {
		err := svc.AddComment(ctx, postID, username, "")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentTooShort))
	})

	t.Run("validation failure - whitespace content", func(t *testing.T) {
		err := svc.AddComment(ctx, postID, username, "   ")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentTooShort))
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
			Return(fmt.Errorf("%w: %vd", domain.ErrRepositoryFailure, "db error"))

		err := svc.AddComment(ctx, postID, username, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrRepositoryFailure))
	})
}

func TestCommentUseCase_GetCommentsByPostID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockPublisher)

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

		comments, err := svc.GetCommentsByPostID(ctx, postID)
		assert.Error(t, err)
		assert.Nil(t, comments)
		assert.True(t, errors.Is(err, domain.ErrRepositoryFailure))
	})
}

func TestCommentUseCase_EditComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockPublisher)

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

		err := svc.EditComment(ctx, commentID, nonAuthorUsername, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentNotOwned))
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
		assert.True(t, errors.Is(err, domain.ErrInvalidComment))
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
		assert.True(t, errors.Is(err, domain.ErrInvalidComment))
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

		err := svc.EditComment(ctx, commentID, authorUsername, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrRepositoryFailure))
	})
}

func TestCommentUseCase_DeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockPublisher)

	ctx := context.Background()
	commentID := 1
	authorUsername := "author"
	nonAuthorUsername := "hacker"

	t.Run("success", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockCommentRepo.EXPECT().
			Delete(ctx, commentID).
			Return(nil)

		err := svc.DeleteComment(ctx, commentID, authorUsername)
		assert.NoError(t, err)
	})

	t.Run("comment not found - error", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, domain.ErrCommentNotFound)

		err := svc.DeleteComment(ctx, commentID, authorUsername)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentNotFound))
	})

	t.Run("comment not found - other repo error", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, errors.New("db error"))

		err := svc.DeleteComment(ctx, commentID, authorUsername)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrRepositoryFailure))
	})

	t.Run("comment is nil", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, nil)

		err := svc.DeleteComment(ctx, commentID, authorUsername)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentNotFound))
	})

	t.Run("unauthorized", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		err := svc.DeleteComment(ctx, commentID, nonAuthorUsername)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrCommentNotOwned))
	})

	t.Run("repository delete error", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			Username: authorUsername,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockCommentRepo.EXPECT().
			Delete(ctx, commentID).
			Return(errors.New("db delete error"))

		err := svc.DeleteComment(ctx, commentID, authorUsername)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrRepositoryFailure))
	})
}
