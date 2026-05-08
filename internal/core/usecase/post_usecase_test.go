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

func TestPostUseCase_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLog)

	ctx := context.Background()
	userID := 1
	title := "Test Title"
	content := "Test Content"
	tags := []string{"tag1", "tag2"}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, post *domain.Post) error {
				post.ID = 1
				return nil
			})
		mockLog.EXPECT().Info(ctx, "post created successfully", "postID", gomock.Any(), "userID", gomock.Any())

		post, err := svc.CreatePost(ctx, userID, title, content, tags)

		assert.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, 1, post.ID)
		assert.Equal(t, title, post.Title)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to create post", "userID", gomock.Any(), "error", gomock.Any())

		post, err := svc.CreatePost(ctx, userID, title, content, tags)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})
}

func TestPostUseCase_GetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLog)

	ctx := context.Background()
	userID := 1
	postID := 1

	t.Run("success", func(t *testing.T) {
		expectedPost := &domain.Post{ID: postID, UserID: userID, Title: "Test Title"}
		mockRepo.EXPECT().GetByID(ctx, postID).Return(expectedPost, nil)

		post, err := svc.GetPostByID(ctx, userID, postID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPost, post)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, postID).Return(nil, domain.ErrPostNotFound)

		post, err := svc.GetPostByID(ctx, userID, postID)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, postID).Return(nil, errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to get post by id", "postID", postID, "error", gomock.Any())

		post, err := svc.GetPostByID(ctx, userID, postID)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})

	t.Run("invalid id", func(t *testing.T) {
		post, err := svc.GetPostByID(ctx, userID, 0)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInvalidPost))
	})

	t.Run("unauthorized access", func(t *testing.T) {
		expectedPost := &domain.Post{ID: postID, UserID: 2, Title: "Test Title"} // Belongs to a different user
		mockRepo.EXPECT().GetByID(ctx, postID).Return(expectedPost, nil)
		mockLog.EXPECT().Error(ctx, "unauthorized access to post", "postID", postID, "userID", userID)

		post, err := svc.GetPostByID(ctx, userID, postID)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrUnauthorized))
	})
}

func TestPostUseCase_UpdatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLog)

	ctx := context.Background()
	userID := 1
	postID := 1
	initialTitle := "Old Title"
	initialContent := "Old Content"

	t.Run("success full update", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, UserID: userID, Title: initialTitle, Content: initialContent}
		newTitle := "New Title"
		newContent := "New Content"

		mockRepo.EXPECT().GetByID(ctx, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, post *domain.Post) error {
			assert.Equal(t, newTitle, post.Title)
			assert.Equal(t, newContent, post.Content)
			return nil
		})
		mockLog.EXPECT().Info(ctx, "post updated successfully", "postID", postID, "userID", userID)

		post, err := svc.UpdatePost(ctx, userID, postID, newTitle, newContent)

		assert.NoError(t, err)
		assert.Equal(t, newTitle, post.Title)
		assert.Equal(t, newContent, post.Content)
	})

	t.Run("success partial update - title only", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, UserID: userID, Title: initialTitle, Content: initialContent}
		newTitle := "New Title"

		mockRepo.EXPECT().GetByID(ctx, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, post *domain.Post) error {
			assert.Equal(t, newTitle, post.Title)
			assert.Equal(t, initialContent, post.Content)
			return nil
		})
		mockLog.EXPECT().Info(ctx, "post updated successfully", "postID", postID, "userID", userID)

		post, err := svc.UpdatePost(ctx, userID, postID, newTitle, "")

		assert.NoError(t, err)
		assert.Equal(t, newTitle, post.Title)
		assert.Equal(t, initialContent, post.Content)
	})

	t.Run("success partial update - content only", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, UserID: userID, Title: initialTitle, Content: initialContent}
		newContent := "New Content"

		mockRepo.EXPECT().GetByID(ctx, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, post *domain.Post) error {
			assert.Equal(t, initialTitle, post.Title)
			assert.Equal(t, newContent, post.Content)
			return nil
		})
		mockLog.EXPECT().Info(ctx, "post updated successfully", "postID", postID, "userID", userID)

		post, err := svc.UpdatePost(ctx, userID, postID, "", newContent)

		assert.NoError(t, err)
		assert.Equal(t, initialTitle, post.Title)
		assert.Equal(t, newContent, post.Content)
	})

	t.Run("unauthorized update", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, UserID: 2, Title: initialTitle, Content: initialContent}
		mockRepo.EXPECT().GetByID(ctx, postID).Return(existingPost, nil)
		mockLog.EXPECT().Error(ctx, "unauthorized update attempt", "postID", postID, "userID", userID)

		post, err := svc.UpdatePost(ctx, userID, postID, "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrUnauthorized))
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, postID).Return(nil, domain.ErrPostNotFound)

		post, err := svc.UpdatePost(ctx, userID, postID, "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
	})

	t.Run("invalid id", func(t *testing.T) {
		post, err := svc.UpdatePost(ctx, userID, 0, "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInvalidPost))
	})
}
