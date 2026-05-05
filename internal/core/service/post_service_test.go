package service

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostService_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	mockLog := mocks.NewMockLogger(ctrl)
	svc := NewPostService(mockRepo, mockLog)

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
