package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCommentUseCase_AddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockPublisher)

	ctx := context.Background()
	postID := 1
	username := "testuser"
	author := &identitydomain.User{ID: 1, Username: username}
	content := "This is a comment"

	t.Run("success", func(t *testing.T) {
		mockPostRepo.EXPECT().
			Load(ctx, int(postID)).
			Return(&domain.Post{ID: int(postID), UserID: 2, Username: "postowner"}, nil)

		mockCommentRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, comment *domain.Comment) error {
				comment.ID = 100
				assert.Equal(t, author.ID, comment.UserID)
				return nil
			})

		mockUserRepo.EXPECT().GetUserByID(ctx, 2).Return(&identitydomain.User{ID: 2, Username: "postowner"}, nil)
		mockPublisher.EXPECT().Publish(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, event platformports.Event) error {
			assert.Equal(t, "user:2", event.Key)
			return nil
		})

		err := svc.AddComment(ctx, postID, author, content)
		assert.NoError(t, err)
	})

	t.Run("validation failure - empty content", func(t *testing.T) {
		err := svc.AddComment(ctx, postID, author, "")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentTooShort))
	})

	t.Run("validation failure - whitespace content", func(t *testing.T) {
		err := svc.AddComment(ctx, postID, author, "   ")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentTooShort))
	})

	t.Run("post not found", func(t *testing.T) {
		mockPostRepo.EXPECT().
			Load(ctx, int(postID)).
			Return(nil, problem.ErrPostNotFound)

		err := svc.AddComment(ctx, postID, author, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrPostNotFound))
	})

	t.Run("repository error on create", func(t *testing.T) {
		mockPostRepo.EXPECT().
			Load(ctx, int(postID)).
			Return(&domain.Post{ID: int(postID)}, nil)

		mockCommentRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(fmt.Errorf("%w: %vd", problem.ErrRepositoryFailure, "db error"))

		err := svc.AddComment(ctx, postID, author, content)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})
}

func TestCommentUseCase_GetCommentsByPostID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostCommandRepository(ctrl)
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
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})
}

func TestCommentUseCase_EditComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockPublisher)

	ctx := context.Background()
	commentID := 1
	authorUsername := "author"
	authorID := 1
	nonAuthorID := 2
	originalContent := "original content"
	newContent := "updated content"

	t.Run("success", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
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

		err := svc.EditComment(ctx, commentID, authorID, newContent)
		assert.NoError(t, err)
	})

	t.Run("comment not found", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, problem.ErrCommentNotFound)

		err := svc.EditComment(ctx, commentID, authorID, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentNotFound))
	})

	t.Run("comment is nil", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, nil)

		err := svc.EditComment(ctx, commentID, authorID, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentNotFound))
	})

	t.Run("unauthorized", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		err := svc.EditComment(ctx, commentID, nonAuthorID, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentNotOwned))
	})

	t.Run("validation failure - empty content", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		err := svc.EditComment(ctx, commentID, authorID, "")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrInvalidComment))
	})

	t.Run("validation failure - whitespace content", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		err := svc.EditComment(ctx, commentID, authorID, "   ")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrInvalidComment))
	})

	t.Run("repository update error", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
			Username: authorUsername,
			Content:  originalContent,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockCommentRepo.EXPECT().
			Update(ctx, gomock.Any()).
			Return(errors.New("db update error"))

		err := svc.EditComment(ctx, commentID, authorID, newContent)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})
}

func TestCommentUseCase_DeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)

	svc := NewCommentUseCase(mockCommentRepo, mockPostRepo, mockUserRepo, mockPublisher)

	ctx := context.Background()
	commentID := 1
	authorUsername := "author"
	authorID := 1
	nonAuthorID := 2

	t.Run("success", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
			Username: authorUsername,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockCommentRepo.EXPECT().
			Delete(ctx, commentID).
			Return(nil)

		err := svc.DeleteComment(ctx, commentID, authorID)
		assert.NoError(t, err)
	})

	t.Run("comment not found - error", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, problem.ErrCommentNotFound)

		err := svc.DeleteComment(ctx, commentID, authorID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentNotFound))
	})

	t.Run("comment not found - other repo error", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, errors.New("db error"))

		err := svc.DeleteComment(ctx, commentID, authorID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})

	t.Run("comment is nil", func(t *testing.T) {
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(nil, nil)

		err := svc.DeleteComment(ctx, commentID, authorID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentNotFound))
	})

	t.Run("unauthorized", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
			Username: authorUsername,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		err := svc.DeleteComment(ctx, commentID, nonAuthorID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrCommentNotOwned))
	})

	t.Run("repository delete error", func(t *testing.T) {
		existingComment := &domain.Comment{
			ID:       commentID,
			UserID:   authorID,
			Username: authorUsername,
		}
		mockCommentRepo.EXPECT().
			GetByID(ctx, commentID).
			Return(existingComment, nil)

		mockCommentRepo.EXPECT().
			Delete(ctx, commentID).
			Return(errors.New("db delete error"))

		err := svc.DeleteComment(ctx, commentID, authorID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})
}
