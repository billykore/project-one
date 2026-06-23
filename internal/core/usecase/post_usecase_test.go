package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostUseCase_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog)

	ctx := context.Background()
	username := "testuser"
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
		mockLog.EXPECT().Info(ctx, "post created successfully", "postID", gomock.Any(), "username", username)

		post, err := svc.CreatePost(ctx, username, title, content, tags)

		assert.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, 1, post.ID)
		assert.Equal(t, title, post.Title)
		assert.Equal(t, username, post.Username)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to create post", "username", username, "error", gomock.Any())

		post, err := svc.CreatePost(ctx, username, title, content, tags)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})
}

func TestPostUseCase_GetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog)

	ctx := context.Background()
	username := "testuser"
	postID := 1

	t.Run("success", func(t *testing.T) {
		expectedPost := &domain.Post{ID: postID, Username: username, Title: "Test Title"}
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(expectedPost, nil)

		post, err := svc.GetPostByID(ctx, postID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPost, post)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(nil, domain.ErrPostNotFound)

		post, err := svc.GetPostByID(ctx, postID)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(nil, errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to get post by id", "postID", postID, "error", gomock.Any())

		post, err := svc.GetPostByID(ctx, postID)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})

	t.Run("invalid id", func(t *testing.T) {
		post, err := svc.GetPostByID(ctx, 0)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInvalidPost))
	})
}

func TestPostUseCase_GetPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog)

	ctx := context.Background()
	username := "testuser"
	limit := 10
	offset := 0

	t.Run("success", func(t *testing.T) {
		expectedPosts := []*domain.Post{
			{ID: 1, Username: username, Title: "Post 1"},
			{ID: 2, Username: username, Title: "Post 2"},
		}
		mockRepo.EXPECT().GetUserPosts(ctx, username, limit, offset).Return(expectedPosts, nil)

		posts, err := svc.GetPosts(ctx, username, limit, offset)

		assert.NoError(t, err)
		assert.Equal(t, expectedPosts, posts)
	})

	t.Run("empty results", func(t *testing.T) {
		mockRepo.EXPECT().GetUserPosts(ctx, username, limit, offset).Return([]*domain.Post{}, nil)

		posts, err := svc.GetPosts(ctx, username, limit, offset)

		assert.NoError(t, err)
		assert.Empty(t, posts)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetUserPosts(ctx, username, limit, offset).Return(nil, errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to get posts for user", "username", username, "error", gomock.Any())

		posts, err := svc.GetPosts(ctx, username, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})

	t.Run("pagination defaults", func(t *testing.T) {
		mockRepo.EXPECT().GetUserPosts(ctx, username, 10, 0).Return([]*domain.Post{}, nil)

		_, err := svc.GetPosts(ctx, username, 0, -1)

		assert.NoError(t, err)
	})
}

func TestPostUseCase_UpdatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog)

	ctx := context.Background()
	username := "testuser"
	postID := 1
	initialTitle := "Old Title"
	initialContent := "Old Content"

	t.Run("success full update", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, Username: username, Title: initialTitle, Content: initialContent}
		newTitle := "New Title"
		newContent := "New Content"

		mockRepo.EXPECT().GetByID(ctx, username, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Update(ctx, username, gomock.Any()).DoAndReturn(func(ctx context.Context, username string, post *domain.Post) error {
			assert.Equal(t, newTitle, post.Title)
			assert.Equal(t, newContent, post.Content)
			return nil
		})
		mockLog.EXPECT().Info(ctx, "post updated successfully", "postID", postID, "username", username)

		post, err := svc.UpdatePost(ctx, username, postID, newTitle, newContent)

		assert.NoError(t, err)
		assert.Equal(t, newTitle, post.Title)
		assert.Equal(t, newContent, post.Content)
	})

	t.Run("success partial update - title only", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, Username: username, Title: initialTitle, Content: initialContent}
		newTitle := "New Title"

		mockRepo.EXPECT().GetByID(ctx, username, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Update(ctx, username, gomock.Any()).DoAndReturn(func(ctx context.Context, username string, post *domain.Post) error {
			assert.Equal(t, newTitle, post.Title)
			assert.Equal(t, initialContent, post.Content)
			return nil
		})
		mockLog.EXPECT().Info(ctx, "post updated successfully", "postID", postID, "username", username)

		post, err := svc.UpdatePost(ctx, username, postID, newTitle, "")

		assert.NoError(t, err)
		assert.Equal(t, newTitle, post.Title)
		assert.Equal(t, initialContent, post.Content)
	})

	t.Run("success partial update - content only", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, Username: username, Title: initialTitle, Content: initialContent}
		newContent := "New Content"

		mockRepo.EXPECT().GetByID(ctx, username, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Update(ctx, username, gomock.Any()).DoAndReturn(func(ctx context.Context, username string, post *domain.Post) error {
			assert.Equal(t, initialTitle, post.Title)
			assert.Equal(t, newContent, post.Content)
			return nil
		})
		mockLog.EXPECT().Info(ctx, "post updated successfully", "postID", postID, "username", username)

		post, err := svc.UpdatePost(ctx, username, postID, "", newContent)

		assert.NoError(t, err)
		assert.Equal(t, initialTitle, post.Title)
		assert.Equal(t, newContent, post.Content)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, username, postID).Return(nil, domain.ErrPostNotFound)

		post, err := svc.UpdatePost(ctx, username, postID, "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
	})

	t.Run("invalid id", func(t *testing.T) {
		post, err := svc.UpdatePost(ctx, username, 0, "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, domain.ErrInvalidPost))
	})
}

func TestPostUseCase_DeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog)

	ctx := context.Background()
	username := "testuser"
	postID := 1

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().Delete(ctx, username, postID).Return(nil)
		mockLog.EXPECT().Info(ctx, "post deleted successfully", "postID", postID, "username", username)

		err := svc.DeletePost(ctx, username, postID)

		assert.NoError(t, err)
	})

	t.Run("repository error on delete", func(t *testing.T) {
		mockRepo.EXPECT().Delete(ctx, username, postID).Return(errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to delete post", "postID", postID, "error", gomock.Any())

		err := svc.DeletePost(ctx, username, postID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInternalServer))
	})

	t.Run("invalid id", func(t *testing.T) {
		err := svc.DeletePost(ctx, username, 0)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPost))
	})
}

func TestPostUseCase_LikePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog)

	ctx := context.Background()
	username := "testuser"
	postID := 1

	t.Run("success - new like", func(t *testing.T) {
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, Username: "postowner", LikeCount: 4}, nil)
		mockLikeRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().IncrementLikeCount(ctx, postID, 1).Return(nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, "postowner").Return(&domain.User{ID: 2, Username: "postowner"}, nil)
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(&domain.User{ID: 1, Username: username}, nil)
		mockPublisher.EXPECT().Publish(ctx, gomock.Any()).Return(nil)
		mockLog.EXPECT().Info(ctx, "post liked successfully", "postID", postID, "username", username)

		count, err := svc.LikePost(ctx, postID, username)
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("success idempotent - already liked", func(t *testing.T) {
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 4}, nil)
		mockLikeRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrAlreadyLiked)

		count, err := svc.LikePost(ctx, postID, username)
		assert.NoError(t, err)
		assert.Equal(t, 4, count)
	})

	t.Run("post not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(nil, domain.ErrPostNotFound)

		count, err := svc.LikePost(ctx, postID, username)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
		assert.Equal(t, 0, count)
	})
}

func TestPostUseCase_UnlikePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog)

	ctx := context.Background()
	username := "testuser"
	postID := 1

	t.Run("success - unlike existing", func(t *testing.T) {
		mockLikeRepo.EXPECT().Delete(ctx, postID, username).Return(nil)
		mockRepo.EXPECT().IncrementLikeCount(ctx, postID, -1).Return(nil)
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 3}, nil)
		mockLog.EXPECT().Info(ctx, "post unliked successfully", "postID", postID, "username", username)

		count, err := svc.UnlikePost(ctx, postID, username)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("success idempotent - not liked", func(t *testing.T) {
		mockLikeRepo.EXPECT().Delete(ctx, postID, username).Return(domain.ErrNotLiked)
		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 4}, nil)

		count, err := svc.UnlikePost(ctx, postID, username)
		assert.NoError(t, err)
		assert.Equal(t, 4, count)
	})
}
