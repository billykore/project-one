package usecase

import (
	"context"
	"errors"
	"testing"

	featureflagdomain "github.com/billykore/project-one/internal/featureflags/domain"
	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostCommandUseCase_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	mockEvaluator := mocks.NewMockFeatureFlagEvaluator(ctrl)
	svc := NewPostCommandUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog, mockEvaluator)

	ctx := context.Background()
	user := &identitydomain.User{ID: 42, Username: "testuser"}
	title := "Test Title"
	content := "Test Content"
	tags := []string{"tag1", "tag2"}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			Save(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, post *domain.Post) error {
				assert.Equal(t, user.ID, post.UserID)
				assert.Equal(t, user.Username, post.Username)
				post.ID = 1
				return nil
			})
		mockLog.EXPECT().Info(ctx, "post created successfully", "postID", gomock.Any(), "username", user.Username)
		mockEvaluator.EXPECT().Evaluate(ctx, "post-creation", user.Username).Return(featureflagdomain.FeatureFlagDecision{Key: "post-creation", Enabled: true, Source: featureflagdomain.SourceEnabledAll})

		post, err := svc.CreatePost(ctx, user, title, content, tags)

		assert.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, 1, post.ID)
		assert.Equal(t, title, post.Title)
		assert.Equal(t, user.Username, post.Username)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			Save(ctx, gomock.Any()).
			Return(errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to create post", "username", user.Username, "error", gomock.Any())
		mockEvaluator.EXPECT().Evaluate(ctx, "post-creation", user.Username).Return(featureflagdomain.FeatureFlagDecision{Key: "post-creation", Enabled: true, Source: featureflagdomain.SourceEnabledAll})

		post, err := svc.CreatePost(ctx, user, title, content, tags)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})
}

func TestPostCommandUseCase_CreatePost_RejectsUnknownFeatureFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	mockEvaluator := mocks.NewMockFeatureFlagEvaluator(ctrl)
	mockEvaluator.EXPECT().Evaluate(gomock.Any(), "post-creation", "testuser").Return(featureflagdomain.FeatureFlagDecision{Source: featureflagdomain.SourceUnknown})

	svc := NewPostCommandUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog, mockEvaluator)
	post, err := svc.CreatePost(context.Background(), &identitydomain.User{ID: 42, Username: "testuser"}, "title", "content", nil)

	assert.Nil(t, post)
	assert.ErrorIs(t, err, problem.ErrFeatureDisabled)
}

func TestPostQueryUseCase_GetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostQueryRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostQueryUseCase(mockRepo, mockLikeRepo, mockLog)

	ctx := context.Background()
	username := "testuser"
	postID := 1

	t.Run("success", func(t *testing.T) {
		expectedPost := &domain.Post{ID: postID, Username: username, Title: "Test Title"}
		mockRepo.EXPECT().GetByID(ctx, postID).Return(expectedPost, nil)

		post, err := svc.GetPostByID(ctx, postID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPost, post)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, postID).Return(nil, problem.ErrPostNotFound)

		post, err := svc.GetPostByID(ctx, postID)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, problem.ErrPostNotFound))
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, postID).Return(nil, errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to get post by id", "postID", postID, "error", gomock.Any())

		post, err := svc.GetPostByID(ctx, postID)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})

	t.Run("invalid id", func(t *testing.T) {
		post, err := svc.GetPostByID(ctx, 0)

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, problem.ErrInvalidPost))
	})
}

func TestPostQueryUseCase_GetPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostQueryRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostQueryUseCase(mockRepo, mockLikeRepo, mockLog)

	ctx := context.Background()
	username := "testuser"
	userID := 1
	limit := 10

	t.Run("success", func(t *testing.T) {
		expectedPosts := []*domain.Post{
			{ID: 1, Username: username, Title: "Post 1"},
			{ID: 2, Username: username, Title: "Post 2"},
		}
		mockRepo.EXPECT().GetUserPosts(ctx, userID, nil, limit+1).Return(expectedPosts, nil)

		posts, nextCursor, hasMore, err := svc.GetPosts(ctx, userID, nil, limit)

		assert.NoError(t, err)
		assert.Equal(t, expectedPosts, posts)
		assert.Nil(t, nextCursor)
		assert.False(t, hasMore)
	})

	t.Run("empty results", func(t *testing.T) {
		mockRepo.EXPECT().GetUserPosts(ctx, userID, nil, limit+1).Return([]*domain.Post{}, nil)

		posts, _, _, err := svc.GetPosts(ctx, userID, nil, limit)

		assert.NoError(t, err)
		assert.Empty(t, posts)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetUserPosts(ctx, userID, nil, limit+1).Return(nil, errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to get posts for user", "userID", userID, "error", gomock.Any())

		posts, _, _, err := svc.GetPosts(ctx, userID, nil, limit)

		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})

	t.Run("pagination defaults", func(t *testing.T) {
		mockRepo.EXPECT().GetUserPosts(ctx, userID, nil, 11).Return([]*domain.Post{}, nil)

		_, _, _, err := svc.GetPosts(ctx, userID, nil, 0)

		assert.NoError(t, err)
	})
}

func TestPostCommandUseCase_UpdatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	mockEvaluator := mocks.NewMockFeatureFlagEvaluator(ctrl)
	svc := NewPostCommandUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog, mockEvaluator)

	ctx := context.Background()
	username := "testuser"
	userID := 1
	postID := 1
	initialTitle := "Old Title"
	initialContent := "Old Content"

	t.Run("success full update", func(t *testing.T) {
		existingPost := &domain.Post{ID: postID, UserID: userID, Username: username, Title: initialTitle, Content: initialContent}
		newTitle := "New Title"
		newContent := "New Content"

		mockRepo.EXPECT().Load(ctx, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Save(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, post *domain.Post) error {
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
		existingPost := &domain.Post{ID: postID, UserID: userID, Username: username, Title: initialTitle, Content: initialContent}
		newTitle := "New Title"

		mockRepo.EXPECT().Load(ctx, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Save(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, post *domain.Post) error {
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
		existingPost := &domain.Post{ID: postID, UserID: userID, Username: username, Title: initialTitle, Content: initialContent}
		newContent := "New Content"

		mockRepo.EXPECT().Load(ctx, postID).Return(existingPost, nil)
		mockRepo.EXPECT().Save(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, post *domain.Post) error {
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

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(nil, problem.ErrPostNotFound)

		post, err := svc.UpdatePost(ctx, userID, postID, "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, problem.ErrPostNotFound))
	})

	t.Run("invalid id", func(t *testing.T) {
		post, err := svc.UpdatePost(ctx, userID, 0, "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, problem.ErrInvalidPost))
	})
}

func TestPostCommandUseCase_DeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	mockEvaluator := mocks.NewMockFeatureFlagEvaluator(ctrl)
	svc := NewPostCommandUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog, mockEvaluator)

	ctx := context.Background()
	username := "testuser"
	userID := 1
	postID := 1

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID, UserID: userID, Username: username}, nil)
		mockRepo.EXPECT().Delete(ctx, gomock.Any()).Return(nil)
		mockLog.EXPECT().Info(ctx, "post deleted successfully", "postID", postID, "userID", userID)

		err := svc.DeletePost(ctx, userID, postID)

		assert.NoError(t, err)
	})

	t.Run("repository error on delete", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID, UserID: userID, Username: username}, nil)
		mockRepo.EXPECT().Delete(ctx, gomock.Any()).Return(errors.New("db error"))
		mockLog.EXPECT().Error(ctx, "failed to delete post", "postID", postID, "error", gomock.Any())

		err := svc.DeletePost(ctx, userID, postID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrRepositoryFailure))
	})

	t.Run("invalid id", func(t *testing.T) {
		err := svc.DeletePost(ctx, userID, 0)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrInvalidPost))
	})
}

func TestPostCommandUseCase_LikePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	mockEvaluator := mocks.NewMockFeatureFlagEvaluator(ctrl)
	svc := NewPostCommandUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog, mockEvaluator)

	ctx := context.Background()
	username := "testuser"
	actor := &identitydomain.User{ID: 1, Username: username}
	postID := 1

	t.Run("success - new like", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID, UserID: 2, Username: "postowner", LikeCount: 4}, nil)
		mockLikeRepo.EXPECT().SetLiked(ctx, postID, actor.ID, true).Return(5, true, nil)
		mockUserRepo.EXPECT().GetUserByID(ctx, 2).Return(&identitydomain.User{ID: 2, Username: "postowner"}, nil)
		mockPublisher.EXPECT().Publish(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, event platformports.Event) error {
			assert.Equal(t, "user:2", event.Key)
			return nil
		})
		mockLog.EXPECT().Info(ctx, "post liked successfully", "postID", postID, "userID", actor.ID)

		count, err := svc.LikePost(ctx, postID, actor)
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("success idempotent - already liked", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 4}, nil)
		mockLikeRepo.EXPECT().SetLiked(ctx, postID, actor.ID, true).Return(4, false, nil)

		count, err := svc.LikePost(ctx, postID, actor)
		assert.NoError(t, err)
		assert.Equal(t, 4, count)
	})

	t.Run("post not found", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(nil, problem.ErrPostNotFound)

		count, err := svc.LikePost(ctx, postID, actor)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, problem.ErrPostNotFound))
		assert.Equal(t, 0, count)
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := svc.LikePost(ctx, 0, actor)
		assert.ErrorIs(t, err, problem.ErrInvalidPostID)
		_, err = svc.LikePost(ctx, postID, nil)
		assert.ErrorIs(t, err, problem.ErrInvalidUsername)
	})

	t.Run("write failures", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID}, nil)
		mockLikeRepo.EXPECT().SetLiked(ctx, postID, actor.ID, true).Return(0, false, errors.New("write failed"))
		mockLog.EXPECT().Error(ctx, "failed to set like state", "postID", postID, "userID", actor.ID, "error", gomock.Any())
		_, err := svc.LikePost(ctx, postID, actor)
		assert.Error(t, err)
	})
}

func TestPostCommandUseCase_UnlikePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostCommandRepository(ctrl)
	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	mockEvaluator := mocks.NewMockFeatureFlagEvaluator(ctrl)
	svc := NewPostCommandUseCase(mockRepo, mockLikeRepo, mockUserRepo, mockPublisher, mockLog, mockEvaluator)

	ctx := context.Background()
	username := "testuser"
	actor := &identitydomain.User{ID: 1, Username: username}
	postID := 1

	t.Run("success - unlike existing", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 4}, nil)
		mockLikeRepo.EXPECT().SetLiked(ctx, postID, actor.ID, false).Return(3, true, nil)
		mockLog.EXPECT().Info(ctx, "post unliked successfully", "postID", postID, "userID", actor.ID)

		count, err := svc.UnlikePost(ctx, postID, actor)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("success idempotent - not liked", func(t *testing.T) {
		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 4}, nil)
		mockLikeRepo.EXPECT().SetLiked(ctx, postID, actor.ID, false).Return(4, false, nil)

		count, err := svc.UnlikePost(ctx, postID, actor)
		assert.NoError(t, err)
		assert.Equal(t, 4, count)
	})

	t.Run("invalid input and failures", func(t *testing.T) {
		_, err := svc.UnlikePost(ctx, 0, actor)
		assert.ErrorIs(t, err, problem.ErrInvalidPost)
		_, err = svc.UnlikePost(ctx, postID, nil)
		assert.ErrorIs(t, err, problem.ErrInvalidUsername)

		mockRepo.EXPECT().Load(ctx, postID).Return(nil, problem.ErrPostNotFound)
		_, err = svc.UnlikePost(ctx, postID, actor)
		assert.ErrorIs(t, err, problem.ErrPostNotFound)

		mockRepo.EXPECT().Load(ctx, postID).Return(&domain.Post{ID: postID}, nil)
		mockLikeRepo.EXPECT().SetLiked(ctx, postID, actor.ID, false).Return(0, false, errors.New("delete failed"))
		mockLog.EXPECT().Error(ctx, "failed to set like state", "postID", postID, "userID", actor.ID, "error", gomock.Any())
		_, err = svc.UnlikePost(ctx, postID, actor)
		assert.Error(t, err)
	})
}

func TestPostCommandUseCase_PublishLikeNotificationFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	posts := mocks.NewMockPostCommandRepository(ctrl)
	likes := mocks.NewMockLikeRepository(ctrl)
	users := mocks.NewMockUserRepository(ctrl)
	publisher := mocks.NewMockPublisher(ctrl)
	log := mocks.NewMockLogger(ctrl)
	evaluator := mocks.NewMockFeatureFlagEvaluator(ctrl)
	uc := NewPostCommandUseCase(posts, likes, users, publisher, log, evaluator).(*postCommandUseCase)
	ctx := context.Background()
	post := &domain.Post{ID: 1, UserID: 1, Username: "owner"}
	actor := &identitydomain.User{ID: 2, Username: "liker"}

	users.EXPECT().GetUserByID(ctx, 1).Return(nil, errors.New("lookup failed"))
	log.EXPECT().Error(ctx, "failed to resolve post owner for like notification", "userID", 1, "error", gomock.Any())
	uc.publishLikeNotification(ctx, post, actor)

	users.EXPECT().GetUserByID(ctx, 1).Return(nil, nil)
	uc.publishLikeNotification(ctx, post, actor)

	users.EXPECT().GetUserByID(ctx, 1).Return(&identitydomain.User{ID: 1, Username: "owner"}, nil)
	publisher.EXPECT().Publish(ctx, gomock.Any()).Return(errors.New("publish failed"))
	log.EXPECT().Error(ctx, "failed to publish like notification", "error", gomock.Any())
	uc.publishLikeNotification(ctx, post, actor)
}
